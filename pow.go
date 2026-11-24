package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"
	"sync"

	"github.com/opd-ai/go-randomx"
)

// AuxPoWProof encapsulates a Monero merge-mining submission for Aperto.
type AuxPoWProof struct {
	MoneroBlockHeader []byte   // Serialized Monero parent block header
	MoneroNonce       uint64   // Winning Monero RandomX nonce
	MoneroSeedHash    []byte   // Current Monero 2048-block RandomX epoch seed
	ApertoBlockHash   [32]byte // Aperto block header hash committed inside Monero's coinbase/extra field
	CoinbaseBranch    [][32]byte // Merkle proof linking Aperto's hash to Monero's Merkle root
}

// BlockchainVerifier stores the pre-allocated RandomX dataset to prevent initialization freezes.
type BlockchainVerifier struct {
	mu          sync.RWMutex
	Hasher      *randomx.Hasher
	CurrentSeed []byte
}

// NewVerifier initializes the global dataset once during daemon startup.
func NewVerifier(initialSeed []byte) (*BlockchainVerifier, error) {
	config := randomx.Config{
		Mode:     randomx.FastMode,
		CacheKey: initialSeed,
	}
	hasher, err := randomx.New(config)
	if err != nil {
		return nil, errors.New("failed to initialize global RandomX dataset")
	}
	return &BlockchainVerifier{
		Hasher:      hasher,
		CurrentSeed: append([]byte(nil), initialSeed...),
	}, nil
}

// ensureSeed safely updates the RandomX cache dataset if the epoch seed has changed.
func (bv *BlockchainVerifier) ensureSeed(seedHash []byte) error {
	bv.mu.RLock()
	same := bytes.Equal(bv.CurrentSeed, seedHash)
	bv.mu.RUnlock()
	if same {
		return nil
	}

	bv.mu.Lock()
	defer bv.mu.Unlock()
	if !bytes.Equal(bv.CurrentSeed, seedHash) {
		if err := bv.Hasher.UpdateCacheKey(seedHash); err != nil {
			return errors.New("failed to transition RandomX seed epoch")
		}
		bv.CurrentSeed = append([]byte(nil), seedHash...)
	}
	return nil
}

// VerifyRandomXProof uses the pre-allocated verifier instance to validate blocks in milliseconds.
func (bv *BlockchainVerifier) VerifyRandomXProof(blockData []byte, nonce uint64, targetLimit *big.Int, seedHash []byte) (bool, error) {
	if err := bv.ensureSeed(seedHash); err != nil {
		return false, err
	}

	// Safely copy blockData + 8-byte nonce without mutating the caller's underlying slice
	inputData := make([]byte, len(blockData)+8)
	copy(inputData, blockData)
	binary.BigEndian.PutUint64(inputData[len(blockData):], nonce)

	bv.mu.RLock()
	resultHash := bv.Hasher.Hash(inputData)
	bv.mu.RUnlock()

	hashInt := new(big.Int).SetBytes(resultHash[:])
	return hashInt.Cmp(targetLimit) <= 0, nil
}

// VerifyMergeMinedProof validates a Monero AuxPoW submission against Aperto's target difficulty.
// This allows Monero miners to simultaneously secure Aperto with zero extra energy cost.
func (bv *BlockchainVerifier) VerifyMergeMinedProof(aux AuxPoWProof, targetLimit *big.Int) (bool, error) {
	if len(aux.MoneroBlockHeader) == 0 {
		return false, errors.New("empty Monero parent block header")
	}

	// 1. Verify that Aperto's block hash is committed inside the Monero parent header / Merkle branch
	currentHash := aux.ApertoBlockHash
	for _, sibling := range aux.CoinbaseBranch {
		combined := append(currentHash[:], sibling[:]...)
		currentHash = sha256.Sum256(combined)
	}
	if !bytes.Contains(aux.MoneroBlockHeader, aux.ApertoBlockHash[:]) && !bytes.Contains(aux.MoneroBlockHeader, currentHash[:]) {
		return false, errors.New("Aperto block commitment missing from Monero AuxPoW header")
	}

	// 2. Validate the Monero parent header's RandomX hash against Aperto's difficulty target
	return bv.VerifyRandomXProof(aux.MoneroBlockHeader, aux.MoneroNonce, targetLimit, aux.MoneroSeedHash)
}

// MineBlockWithLockIn hashes nonces for a 1-minute block height, but immediately aborts
// if another valid block passes the <2s pre-sign lock-in verification via ActiveBlockLock.
func (bv *BlockchainVerifier) MineBlockWithLockIn(
	parentCtx context.Context,
	height uint64,
	blockData []byte,
	targetLimit *big.Int,
	seedHash []byte,
) (uint64, [32]byte, error) {
	var zeroHash [32]byte

	// Abort immediately if this height was already locked in
	if ActiveBlockLock != nil && ActiveBlockLock.IsHeightLocked(height) {
		return 0, zeroHash, errors.New("height already locked in before mining started")
	}

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	if ActiveBlockLock != nil {
		ActiveBlockLock.RegisterMiningCancel(height, cancel)
	}

	if err := bv.ensureSeed(seedHash); err != nil {
		return 0, zeroHash, err
	}

	inputData := make([]byte, len(blockData)+8)
	copy(inputData, blockData)

	for nonce := uint64(0); ; nonce++ {
		// Check every iteration if <2s pre-sign lock-in canceled this height
		select {
		case <-ctx.Done():
			return 0, zeroHash, errors.New("CPU hashing halted: block height locked in")
		default:
		}

		binary.BigEndian.PutUint64(inputData[len(blockData):], nonce)

		bv.mu.RLock()
		resultHash := bv.Hasher.Hash(inputData)
		bv.mu.RUnlock()

		hashInt := new(big.Int).SetBytes(resultHash[:])
		if hashInt.Cmp(targetLimit) <= 0 {
			return nonce, resultHash, nil
		}
	}
}
