package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyotoasuna/saracie-core/internal/chain"
	"github.com/kyotoasuna/saracie-core/internal/wallet"
)

func TestUIStatusAndWalletCreate(t *testing.T) {
	dataDir := t.TempDir()
	store, err := chain.Open(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := New(store, dataDir, nil).Handler()
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	var status StatusResponse
	getJSON(t, server.URL+"/api/status", &status)
	if status.Chain.Ticker != "SRCE" {
		t.Fatalf("ticker = %s, want SRCE", status.Chain.Ticker)
	}

	var info wallet.WalletFileInfo
	postJSON(t, server.URL+"/api/wallet/create", map[string]any{
		"wallet":     filepath.Base("test.wallet"),
		"passphrase": "testpass",
	}, &info)
	if info.Address == "" {
		t.Fatal("missing wallet address")
	}
}

func TestUIMinerStartStop(t *testing.T) {
	dataDir := t.TempDir()
	store, err := chain.Open(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := New(store, dataDir, nil).Handler()
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	account, err := wallet.NewAccount(0)
	if err != nil {
		t.Fatal(err)
	}

	var started MinerState
	postJSON(t, server.URL+"/api/miner/start", map[string]any{
		"address": account.Address,
	}, &started)
	if !started.Running {
		t.Fatal("miner did not start")
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		var status StatusResponse
		getJSON(t, server.URL+"/api/status", &status)
		if status.Chain.Height > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	var stopped MinerState
	postJSON(t, server.URL+"/api/miner/stop", map[string]any{}, &stopped)
	if stopped.Running {
		t.Fatal("miner did not stop")
	}
}

func getJSON(t *testing.T, url string, out any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned %s", url, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}

func postJSON(t *testing.T, url string, body any, out any) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST %s returned %s", url, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}
