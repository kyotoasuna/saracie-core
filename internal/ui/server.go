package ui

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/kyotoasuna/saracie-core/internal/chain"
	"github.com/kyotoasuna/saracie-core/internal/consensus"
	"github.com/kyotoasuna/saracie-core/internal/node"
	"github.com/kyotoasuna/saracie-core/internal/wallet"
)

//go:embed assets/*
var assets embed.FS

type Server struct {
	store     *chain.Store
	dataDir   string
	peers     []string
	mu        sync.Mutex
	miner     MinerState
	cancelMin context.CancelFunc
	syncing   bool
	lastSync  time.Time
	syncError string
}

type MinerState struct {
	Running     bool   `json:"running"`
	Address     string `json:"address"`
	BlocksMined uint64 `json:"blocks_mined"`
	LastHeight  uint64 `json:"last_height"`
	LastHash    string `json:"last_hash"`
	LastReward  string `json:"last_reward"`
	LastError   string `json:"last_error"`
	StartedAt   int64  `json:"started_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type StatusResponse struct {
	Chain  chain.Status `json:"chain"`
	Miner  MinerState   `json:"miner"`
	Peers  []string     `json:"peers"`
	Uptime int64        `json:"uptime_seconds"`
}

func New(store *chain.Store, dataDir string, peers []string) *Server {
	return &Server{
		store:   store,
		dataDir: dataDir,
		peers:   peers,
	}
}

func (s *Server) Handler() (http.Handler, error) {
	static, err := fs.Sub(assets, "assets")
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(static)))
	mux.HandleFunc("/api/status", s.status)
	mux.HandleFunc("/api/wallet/create", s.walletCreate)
	mux.HandleFunc("/api/wallet/open", s.walletOpen)
	mux.HandleFunc("/api/balance", s.balance)
	mux.HandleFunc("/api/send-file", s.sendFile)
	mux.HandleFunc("/api/miner/start", s.minerStart)
	mux.HandleFunc("/api/miner/stop", s.minerStop)
	return mux, nil
}

func (s *Server) ListenAndServe(addr string) error {
	handler, err := s.Handler()
	if err != nil {
		return err
	}
	return http.ListenAndServe(addr, handler)
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.syncFromPeersIfStale()
	s.mu.Lock()
	resp := StatusResponse{
		Chain: s.store.Status(),
		Miner: s.miner,
		Peers: append([]string(nil), s.peers...),
	}
	if s.miner.StartedAt > 0 {
		resp.Uptime = time.Now().Unix() - s.miner.StartedAt
	}
	s.mu.Unlock()
	writeJSON(w, resp)
}

func (s *Server) walletCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Wallet     string `json:"wallet"`
		Passphrase string `json:"passphrase"`
		Index      uint32 `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path := s.walletPath(req.Wallet)
	info, err := wallet.CreateEncryptedWalletFile(path, req.Passphrase, req.Index)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, info)
}

func (s *Server) walletOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Wallet     string `json:"wallet"`
		Passphrase string `json:"passphrase"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path := s.walletPath(req.Wallet)
	keyPair, file, err := wallet.LoadEncryptedKeyPair(path, req.Passphrase)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, wallet.WalletFileInfo{
		Wallet:    path,
		Address:   keyPair.Account.Address,
		Path:      keyPair.Account.Path,
		Created:   file.CreatedAt,
		Encrypted: true,
	})
}

func (s *Server) balance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address is required", http.StatusBadRequest)
		return
	}
	s.syncFromPeersIfStale()
	s.mu.Lock()
	balance, err := s.store.Balance(address)
	s.mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, balance)
}

func (s *Server) sendFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Wallet     string `json:"wallet"`
		Passphrase string `json:"passphrase"`
		To         string `json:"to"`
		Amount     string `json:"amount"`
		Fee        string `json:"fee"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Fee == "" {
		req.Fee = "0.00001000"
	}

	amount, err := consensus.ParseAmount(req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fee, err := consensus.ParseAmount(req.Fee)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	keyPair, _, err := wallet.LoadEncryptedKeyPair(s.walletPath(req.Wallet), req.Passphrase)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	toPubKeyHash, err := wallet.AddressToPubKeyHash(req.To)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.syncFromPeersIfStale()
	s.mu.Lock()
	utxos, err := s.store.SpendableUTXOs(keyPair.PubKeyHash)
	if err == nil {
		tx, txErr := chain.NewSignedTransaction(utxos, keyPair.PrivateKey, keyPair.PublicKey, toPubKeyHash, keyPair.PubKeyHash, amount, fee)
		if txErr == nil {
			err = s.store.AddMempoolTx(tx)
			if err == nil {
				s.mu.Unlock()
				for _, peer := range s.peers {
					_ = node.SubmitTransaction(peer, tx)
				}
				writeJSON(w, tx)
				return
			}
		} else {
			err = txErr
		}
	}
	s.mu.Unlock()

	http.Error(w, err.Error(), http.StatusBadRequest)
}

func (s *Server) minerStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := wallet.AddressToPubKeyHash(req.Address); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.syncFromPeersIfStale()
	s.mu.Lock()
	if s.miner.Running {
		state := s.miner
		s.mu.Unlock()
		writeJSON(w, state)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	now := time.Now().Unix()
	s.cancelMin = cancel
	s.miner = MinerState{
		Running:   true,
		Address:   req.Address,
		StartedAt: now,
		UpdatedAt: now,
	}
	state := s.miner
	s.mu.Unlock()

	go s.mineLoop(ctx, req.Address)
	writeJSON(w, state)
}

func (s *Server) minerStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	if s.cancelMin != nil {
		s.cancelMin()
	}
	s.miner.Running = false
	s.miner.UpdatedAt = time.Now().Unix()
	state := s.miner
	s.mu.Unlock()
	writeJSON(w, state)
}

func (s *Server) mineLoop(ctx context.Context, address string) {
	for {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			s.miner.Running = false
			s.miner.UpdatedAt = time.Now().Unix()
			s.mu.Unlock()
			return
		default:
		}

		s.syncFromPeersIfStale()
		s.mu.Lock()
		block, err := s.store.MineNext(ctx, address)
		if err != nil {
			s.miner.LastError = err.Error()
			s.miner.Running = false
			s.miner.UpdatedAt = time.Now().Unix()
			s.mu.Unlock()
			return
		}
		s.miner.BlocksMined++
		s.miner.LastHeight = block.Header.Height
		s.miner.LastHash = block.Hash()
		s.miner.LastReward = consensus.FormatAmount(coinbaseValue(block))
		s.miner.UpdatedAt = time.Now().Unix()
		peers := append([]string(nil), s.peers...)
		s.mu.Unlock()

		for _, peer := range peers {
			_ = node.SubmitBlock(peer, block)
		}
	}
}

func (s *Server) syncFromPeersIfStale() {
	s.mu.Lock()
	if len(s.peers) == 0 || s.syncing || time.Since(s.lastSync) < 5*time.Second {
		s.mu.Unlock()
		return
	}
	peers := append([]string(nil), s.peers...)
	s.syncing = true
	s.mu.Unlock()

	var firstErr error
	for _, peer := range peers {
		blocks, err := node.FetchBlocks(peer)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
		} else {
			s.mu.Lock()
			_, err = s.store.ReplaceIfValidLonger(blocks)
			s.mu.Unlock()
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}

		txs, err := node.FetchMempool(peer)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		s.mu.Lock()
		for _, tx := range txs {
			_ = s.store.AddMempoolTx(tx)
		}
		s.mu.Unlock()
	}

	s.mu.Lock()
	s.syncing = false
	s.lastSync = time.Now()
	if firstErr != nil {
		s.syncError = firstErr.Error()
	} else {
		s.syncError = ""
	}
	s.mu.Unlock()
}

func (s *Server) walletPath(name string) string {
	if name == "" {
		name = "saracie.wallet"
	}
	if filepath.IsAbs(name) {
		return name
	}
	return filepath.Join(s.dataDir, name)
}

func coinbaseValue(block chain.Block) int64 {
	if len(block.Transactions) == 0 {
		return 0
	}
	var total int64
	for _, out := range block.Transactions[0].Outputs {
		total += out.Value
	}
	return total
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func LocalURL(addr string) string {
	if addr == "" {
		return "http://127.0.0.1:7340"
	}
	if addr[0] == ':' {
		return fmt.Sprintf("http://127.0.0.1%s", addr)
	}
	return "http://" + addr
}
