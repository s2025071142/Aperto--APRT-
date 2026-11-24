package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
	"sync"
)

// StorageMode defines how the node stores and verifies ledger state
type StorageMode string

const (
	ModeUTXO     StorageMode = "utxo"     // Lightweight UTXO accumulator (default on mobile)
	ModePruned   StorageMode = "pruned"   // Rolling pruned state/blocks (default on PC)
	ModeArchival StorageMode = "archival" // Full unpruned historical blocks
)

// NodeStorageConfig holds user-defined or auto-detected storage settings
type NodeStorageConfig struct {
	Mode            StorageMode
	PruneKeepRecent uint64
}

// UTXO represents an unspent transparent transaction output
type UTXO struct {
	TxHash string
	Index  uint32
	Owner  string
	Amount uint64
}

// StorageManager handles UTXO commitments and historical block pruning
type StorageManager struct {
	mu              sync.RWMutex
	Config          NodeStorageConfig
	UTXOSet         map[string]UTXO // Active unspent outputs (used in UTXO/Pruned modes)
	AccumulatorRoot [32]byte        // Compact cryptographic commitment of all active UTXOs
	RetainedBlocks  map[uint64][]byte
}

// ActiveStorageManager is accessed by app.go and abci.go during block execution
var ActiveStorageManager *StorageManager

// DetectDefaultStorageMode defaults to 'utxo' on mobile OS and 'pruned' on desktop/server
func DetectDefaultStorageMode() StorageMode {
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return ModeUTXO
	}
	return ModePruned
}

// NewNodeStorageConfig validates CLI inputs and builds the storage configuration
func NewNodeStorageConfig(modeStr string, keepRecent uint64) (NodeStorageConfig, error) {
	mode := StorageMode(modeStr)
	switch mode {
	case ModeUTXO, ModePruned, ModeArchival:
		if keepRecent == 0 {
			keepRecent = 1000
		}
		return NodeStorageConfig{
			Mode:            mode,
			PruneKeepRecent: keepRecent,
		}, nil
	default:
		return NodeStorageConfig{}, fmt.Errorf("unknown --storage-mode %q (must be 'utxo', 'pruned', or 'archival')", modeStr)
	}
}

// NewStorageManager initializes a clean storage backend
func NewStorageManager(cfg NodeStorageConfig) *StorageManager {
	return &StorageManager{
		Config:         cfg,
		UTXOSet:        make(map[string]UTXO),
		RetainedBlocks: make(map[uint64][]byte),
	}
}

// ApplyBlockState updates UTXOs and prunes old blocks according to the active StorageMode
func (sm *StorageManager) ApplyBlockState(height uint64, blockBytes []byte, spentKeys []string, newOutputs []UTXO) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 1. Remove spent UTXOs and add newly created UTXOs
	for _, key := range spentKeys {
		delete(sm.UTXOSet, key)
	}
	for _, out := range newOutputs {
		key := fmt.Sprintf("%s:%d", out.TxHash, out.Index)
		sm.UTXOSet[key] = out
	}

	// 2. Recompute compact UTXO Accumulator Root (lightweight fingerprint for UTXO mode)
	sm.recomputeAccumulatorRoot()

	// 3. Handle block retention based on the operator's chosen mode
	switch sm.Config.Mode {
	case ModeUTXO:
		// UTXO-only mode: discard raw historical blocks immediately to keep footprint near zero
		sm.RetainedBlocks = make(map[uint64][]byte)

	case ModePruned:
		// Pruned mode: store current block, delete blocks older than PruneKeepRecent
		sm.RetainedBlocks[height] = blockBytes
		if height > sm.Config.PruneKeepRecent {
			cutoff := height - sm.Config.PruneKeepRecent
			for h := range sm.RetainedBlocks {
				if h < cutoff {
					delete(sm.RetainedBlocks, h)
				}
			}
		}

	case ModeArchival:
		// Archival mode: keep every block forever
		sm.RetainedBlocks[height] = blockBytes
	}
}

// recomputeAccumulatorRoot hashes the active UTXO set into a single 32-byte commitment
func (sm *StorageManager) recomputeAccumulatorRoot() {
	hasher := sha256.New()
	for k, u := range sm.UTXOSet {
		hasher.Write([]byte(fmt.Sprintf("%s|%s|%d;", k, u.Owner, u.Amount)))
	}
	copy(sm.AccumulatorRoot[:], hasher.Sum(nil))
}

// GetAccumulatorHex returns the current UTXO commitment root as a hex string
func (sm *StorageManager) GetAccumulatorHex() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return hex.EncodeToString(sm.AccumulatorRoot[:])
}
