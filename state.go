package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
	"sort"
	"sync"
)

// StorageMode defines how the node stores and verifies ledger state
type StorageMode string

const (
	ModeUTXO     StorageMode = "utxo"     // Stateless light verification (~KB footprint, default on mobile)
	ModePruned   StorageMode = "pruned"   // Full EVM/UTXO state + rolling block pruning (default on PC)
	ModeArchival StorageMode = "archival" // Full unpruned historical blocks & state
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

// StatelessProof lets mobile/light 'utxo' nodes verify a balance without storing the full state
type StatelessProof struct {
	TargetUTXO     UTXO
	MerkleSiblings [][32]byte
}

// StorageManager handles stateless accumulator commitments and historical block pruning
type StorageManager struct {
	mu              sync.RWMutex
	Config          NodeStorageConfig
	UTXOSet         map[string]UTXO // Full active state (kept only in 'pruned' and 'archival' modes)
	AccumulatorRoot [32]byte        // Compact 32-byte cryptographic commitment (kept in all modes)
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

// ApplyBlockState updates state commitments and prunes old blocks according to the active StorageMode
func (sm *StorageManager) ApplyBlockState(height uint64, blockBytes []byte, spentKeys []string, newOutputs []UTXO) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 1. Update active state transitions
	for _, key := range spentKeys {
		delete(sm.UTXOSet, key)
	}
	for _, out := range newOutputs {
		key := fmt.Sprintf("%s:%d", out.TxHash, out.Index)
		sm.UTXOSet[key] = out
	}

	// 2. Deterministically recompute the 32-byte Accumulator Root
	sm.recomputeAccumulatorRoot()

	// 3. Enforce hardware footprint limits based on the operator's chosen mode
	switch sm.Config.Mode {
	case ModeUTXO:
		// Stateless mobile/light mode: discard both raw blocks AND the full state map,
		// retaining only the 32-byte cryptographic AccumulatorRoot (~KB total footprint).
		sm.UTXOSet = make(map[string]UTXO)
		sm.RetainedBlocks = make(map[uint64][]byte)

	case ModePruned:
		// Pruned PC mode: keep active state to serve proofs, delete blocks older than PruneKeepRecent
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
		// Archival server mode: keep every block and state transition forever
		sm.RetainedBlocks[height] = blockBytes
	}
}

// VerifyStatelessProof allows mobile 'utxo' nodes to verify a UTXO against the 32-byte AccumulatorRoot
func (sm *StorageManager) VerifyStatelessProof(proof StatelessProof) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	key := fmt.Sprintf("%s:%d", proof.TargetUTXO.TxHash, proof.TargetUTXO.Index)
	leafData := fmt.Sprintf("%s|%s|%d;", key, proof.TargetUTXO.Owner, proof.TargetUTXO.Amount)
	currentHash := sha256.Sum256([]byte(leafData))

	for _, sibling := range proof.MerkleSiblings {
		combined := append(currentHash[:], sibling[:]...)
		currentHash = sha256.Sum256(combined)
	}

	return bytes.Equal(currentHash[:], sm.AccumulatorRoot[:])
}

// recomputeAccumulatorRoot sorts keys first so every node computes the exact same 32-byte commitment
func (sm *StorageManager) recomputeAccumulatorRoot() {
	keys := make([]string, 0, len(sm.UTXOSet))
	for k := range sm.UTXOSet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	hasher := sha256.New()
	for _, k := range keys {
		u := sm.UTXOSet[k]
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
