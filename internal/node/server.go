package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kyotoasuna/saracie-core/internal/chain"
	"github.com/kyotoasuna/saracie-core/internal/consensus"
)

type Server struct {
	Store *chain.Store
	Self  string
	mu    sync.Mutex
	peers map[string]bool
}

func New(store *chain.Store) *Server {
	return NewWithPeers(store, "", nil)
}

func NewWithPeers(store *chain.Store, self string, peers []string) *Server {
	server := &Server{
		Store: store,
		Self:  normalizePeer(self),
		peers: make(map[string]bool),
	}
	for _, peer := range peers {
		server.AddPeer(peer)
	}
	return server
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", s.status)
	mux.HandleFunc("/params", s.params)
	mux.HandleFunc("/blocks", s.blocks)
	mux.HandleFunc("/transactions", s.transactions)
	mux.HandleFunc("/mempool", s.mempool)
	mux.HandleFunc("/peers", s.peersHandler)
	mux.HandleFunc("/scarcity", s.scarcity)
	return mux
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}

func (s *Server) status(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	status := s.Store.Status()
	s.mu.Unlock()
	writeJSON(w, status)
}

func (s *Server) params(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, consensus.Mainnet)
}

func (s *Server) blocks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.Lock()
		blocks := append([]chain.Block(nil), s.Store.Blocks...)
		s.mu.Unlock()
		writeJSON(w, blocks)
	case http.MethodPost:
		var block chain.Block
		if err := json.NewDecoder(r.Body).Decode(&block); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.mu.Lock()
		knownBefore := s.Store.KnowsBlock(block)
		err := s.Store.AddBlock(block)
		status := s.Store.Status()
		s.mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !knownBefore {
			go s.BroadcastBlock(block)
		}
		writeJSON(w, status)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) transactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var tx chain.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		knownBefore := s.Store.HasTransaction(tx.ID) || s.Store.HasMempoolTx(tx.ID)
		err := s.Store.AddMempoolTx(tx)
		status := s.Store.Status()
		s.mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !knownBefore {
			go s.BroadcastTransaction(tx)
		}
		writeJSON(w, status)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) mempool(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.Lock()
		mempool := append([]chain.Transaction(nil), s.Store.Mempool...)
		s.mu.Unlock()
		writeJSON(w, mempool)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) peersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, s.Peers())
	case http.MethodPost:
		var req struct {
			Peer string `json:"peer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Peer == "" {
			http.Error(w, "peer is required", http.StatusBadRequest)
			return
		}
		s.AddPeer(req.Peer)
		writeJSON(w, s.Peers())
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) scarcity(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	height := s.Store.Tip().Header.Height
	s.mu.Unlock()
	writeJSON(w, consensus.Mainnet.ScarcityAt(height))
}

func (s *Server) AddPeer(peer string) bool {
	peer = normalizePeer(peer)
	if peer == "" || peer == s.Self {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.peers[peer] {
		return false
	}
	s.peers[peer] = true
	return true
}

func (s *Server) Peers() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	peers := make([]string, 0, len(s.peers))
	for peer := range s.peers {
		peers = append(peers, peer)
	}
	sort.Strings(peers)
	return peers
}

func (s *Server) BroadcastBlock(block chain.Block) {
	for _, peer := range s.Peers() {
		_ = SubmitBlock(peer, block)
	}
}

func (s *Server) BroadcastTransaction(tx chain.Transaction) {
	for _, peer := range s.Peers() {
		_ = SubmitTransaction(peer, tx)
	}
}

func (s *Server) DiscoverPeers() int {
	added := 0
	for _, peer := range s.Peers() {
		remotePeers, err := FetchPeers(peer)
		if err == nil {
			for _, remotePeer := range remotePeers {
				if s.AddPeer(remotePeer) {
					added++
				}
			}
		}
		if s.Self != "" {
			_ = AnnouncePeer(peer, s.Self)
		}
	}
	return added
}

func (s *Server) SyncNetwork() (int, int, error) {
	peers := s.Peers()
	replaced := 0
	txAdded := 0
	var firstErr error

	for _, peer := range peers {
		blocks, err := FetchBlocks(peer)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
		} else {
			s.mu.Lock()
			ok, err := s.Store.ReplaceIfValidLonger(blocks)
			s.mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
			} else if ok {
				replaced++
			}
		}

		txs, err := FetchMempool(peer)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		s.mu.Lock()
		for _, tx := range txs {
			before := len(s.Store.Mempool)
			if err := s.Store.AddMempoolTx(tx); err != nil {
				continue
			}
			if len(s.Store.Mempool) > before {
				txAdded++
			}
		}
		s.mu.Unlock()
	}

	return replaced, txAdded, firstErr
}

func FetchBlocks(peer string) ([]chain.Block, error) {
	url := strings.TrimRight(peer, "/") + "/blocks"
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("peer %s returned %s: %s", peer, resp.Status, strings.TrimSpace(string(body)))
	}

	var blocks []chain.Block
	if err := json.NewDecoder(resp.Body).Decode(&blocks); err != nil {
		return nil, err
	}
	return blocks, nil
}

func SyncFromPeers(store *chain.Store, peers []string) (int, error) {
	replaced := 0
	for _, peer := range peers {
		peer = strings.TrimSpace(peer)
		if peer == "" {
			continue
		}

		blocks, err := FetchBlocks(peer)
		if err != nil {
			return replaced, err
		}
		ok, err := store.ReplaceIfValidLonger(blocks)
		if err != nil {
			return replaced, err
		}
		if ok {
			replaced++
		}
	}
	return replaced, nil
}

func FetchMempool(peer string) ([]chain.Transaction, error) {
	url := strings.TrimRight(peer, "/") + "/mempool"
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("peer %s returned %s: %s", peer, resp.Status, strings.TrimSpace(string(body)))
	}

	var txs []chain.Transaction
	if err := json.NewDecoder(resp.Body).Decode(&txs); err != nil {
		return nil, err
	}
	return txs, nil
}

func SyncMempoolFromPeers(store *chain.Store, peers []string) (int, error) {
	added := 0
	for _, peer := range peers {
		peer = strings.TrimSpace(peer)
		if peer == "" {
			continue
		}

		txs, err := FetchMempool(peer)
		if err != nil {
			return added, err
		}
		for _, tx := range txs {
			before := len(store.Mempool)
			if err := store.AddMempoolTx(tx); err != nil {
				continue
			}
			if len(store.Mempool) > before {
				added++
			}
		}
	}
	return added, nil
}

func FetchPeers(peer string) ([]string, error) {
	url := strings.TrimRight(peer, "/") + "/peers"
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("peer %s returned %s: %s", peer, resp.Status, strings.TrimSpace(string(body)))
	}

	var peers []string
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return nil, err
	}
	return peers, nil
}

func AnnouncePeer(peer, self string) error {
	self = normalizePeer(self)
	if self == "" {
		return nil
	}
	url := strings.TrimRight(peer, "/") + "/peers"
	raw, err := json.Marshal(struct {
		Peer string `json:"peer"`
	}{Peer: self})
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("peer %s returned %s: %s", peer, resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func SubmitBlock(peer string, block chain.Block) error {
	url := strings.TrimRight(peer, "/") + "/blocks"
	raw, err := json.Marshal(block)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("peer %s returned %s: %s", peer, resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func normalizePeer(peer string) string {
	peer = strings.TrimSpace(peer)
	if peer == "" {
		return ""
	}
	return strings.TrimRight(peer, "/")
}

func SubmitTransaction(peer string, tx chain.Transaction) error {
	url := strings.TrimRight(peer, "/") + "/transactions"
	raw, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("peer %s returned %s: %s", peer, resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
