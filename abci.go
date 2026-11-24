package main

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"://github.com"
)

const (
	// TargetBlockTime enforces the 1-minute block interval
	TargetBlockTime = 60 * time.Second

	// PreSignLockTimeout enforces < 2-second validation to lock in a valid block
	PreSignLockTimeout = 2 * time.Second

	// PenaltyFreeBlockBytes is the baseline block size (300 KB) before Monero-style reward penalties apply
	PenaltyFreeBlockBytes uint64 = 300 * 1024
)

// BlockLockState tracks locked heights so nodes immediately stop hashing once a valid block locks in
type BlockLockState struct {
	mu           sync.RWMutex
	lockedHeight uint64
	lockedHash   string
	cancelMining context.CancelFunc
}

var ActiveBlockLock = &BlockLockState{}

// RegisterMiningCancel allows pow.go to register a cancel function for the active block height
func (b *BlockLockState) RegisterMiningCancel(height uint64, cancel context.CancelFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.lockedHeight < height {
		b.cancelMining = cancel
	}
}

// IsHeightLocked returns true if this block height is already verified and locked in
func (b *BlockLockState) IsHeightLocked(height uint64) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lockedHeight >= height
}

// PreSignAndLockBlock validates a candidate block's pre-signature under a strict 2-second deadline.
// Once verified, it locks the height and immediately cancels ongoing CPU hashing for that block.
func PreSignAndLockBlock(
	height uint64,
	blockHash string,
	headerBytes []byte,
	preSig []byte,
	proposerPubKey ed25519.PublicKey,
) error {
	start := time.Now()

	// Reject immediately if another valid block already locked in this height
	if ActiveBlockLock.IsHeightLocked(height) {
		return fmt.Errorf("block height %d is already locked in; halting duplicate work", height)
	}

	// Verify the cryptographic pre-signature
	if len(proposerPubKey) != ed25519.PublicKeySize || !ed25519.Verify(proposerPubKey, headerBytes, preSig) {
		return errors.New("invalid pre-sign signature on candidate block")
	}

	// Enforce the strict < 2 second validation ceiling
	elapsed := time.Since(start)
	if elapsed >= PreSignLockTimeout {
		return fmt.Errorf("pre-sign validation exceeded 2s limit (%v); rejecting stale lock-in", elapsed)
	}

	// Lock the height atomically and signal the miner to stop hashing this height
	ActiveBlockLock.mu.Lock()
	if ActiveBlockLock.lockedHeight < height {
		ActiveBlockLock.lockedHeight = height
		ActiveBlockLock.lockedHash = blockHash
		if ActiveBlockLock.cancelMining != nil {
			ActiveBlockLock.cancelMining()
			ActiveBlockLock.cancelMining = nil
		}
	}
	ActiveBlockLock.mu.Unlock()

	fmt.Printf("[Lock-In] Height %d verified & locked in %v (Hash: %s) — CPU hashing halted\n", height, elapsed, blockHash)
	return nil
}

// AdjustDynamicGasLimit calculates a moving capacity ceiling for the EVM block window.
// It expands dynamically during burst loads while remaining bounded for consumer hardware.
func AdjustDynamicGasLimit(historicalAverageGas uint64) uint64 {
	const TargetGasLimit uint64 = 90000000
	const MaxGrowthFactor float64 = 4.0

	if historicalAverageGas <= TargetGasLimit {
		return TargetGasLimit
	}

	demandRatio := float64(historicalAverageGas) / float64(TargetGasLimit)
	if demandRatio > MaxGrowthFactor {
		demandRatio = MaxGrowthFactor
	}

	return uint64(float64(TargetGasLimit) * demandRatio)
}

// CalculateBlockRewardWithPenalty applies Monero's quadratic block-size penalty to CalculateBlockSubsidy.
// Blocks up to medianBytes (or PenaltyFreeBlockBytes) keep 100% of the reward.
// Blocks larger than medianBytes (up to 2x max) have their reward slashed quadratically to stop state-bloat spam.
func CalculateBlockRewardWithPenalty(height int64, blockSizeBytes uint64, medianBlockBytes uint64) types.Coin {
	baseCoin := CalculateBlockSubsidy(height)
	baseReward := baseCoin.Amount.BigInt()

	if medianBlockBytes < PenaltyFreeBlockBytes {
		medianBlockBytes = PenaltyFreeBlockBytes
	}

	// No penalty if within the median window
	if blockSizeBytes <= medianBlockBytes {
		return baseCoin
	}

	// Hard cap at 2x median size (100% reward penalty if exceeded)
	if blockSizeBytes >= medianBlockBytes*2 {
		return types.NewCoin(baseCoin.Denom, types.ZeroInt())
	}

	// Deterministic Monero quadratic penalty using math/big (no floats):
	// Penalty = BaseReward * ((blockSize - medianSize)^2) / (medianSize^2)
	excess := new(big.Int).SetUint64(blockSizeBytes - medianBlockBytes)
	excessSquared := new(big.Int).Mul(excess, excess)

	median := new(big.Int).SetUint64(medianBlockBytes)
	medianSquared := new(big.Int).Mul(median, median)

	penalty := new(big.Int).Mul(baseReward, excessSquared)
	penalty.Div(penalty, medianSquared)

	finalReward := new(big.Int).Sub(baseReward, penalty)
	if finalReward.Sign() < 0 {
		finalReward.SetInt64(0)
	}

	return types.NewCoin(baseCoin.Denom, types.NewIntFromBigInt(finalReward))
}

// FinalizeAndPruneBlock executes dynamic gas scaling, calculates spam-penalized block rewards,
// updates UTXO light commitments, and prunes historical blocks at the end of every 1-minute block.
func FinalizeAndPruneBlock(
	height uint64,
	historicalAverageGas uint64,
	rawBlockBytes []byte,
	spentUTXOKeys []string,
	createdUTXOs []UTXO,
) (uint64, string) {
	// 1. Calculate the dynamic block gas limit for the next 1-minute window
	nextGasLimit := AdjustDynamicGasLimit(historicalAverageGas)

	// 2. Compute block reward with Monero-style dynamic size penalty (prevents state bloat)
	blockSize := uint64(len(rawBlockBytes))
	minerReward := CalculateBlockRewardWithPenalty(int64(height), blockSize, PenaltyFreeBlockBytes)

	// 3. Fallback initialization if storage manager wasn't set by CLI
	if ActiveStorageManager == nil {
		defaultCfg, _ := NewNodeStorageConfig(string(DetectDefaultStorageMode()), 1000)
		ActiveStorageManager = NewStorageManager(defaultCfg)
	}

	// 4. Apply UTXO light commitments and prune/discard old blocks based on node mode
	ActiveStorageManager.ApplyBlockState(height, rawBlockBytes, spentUTXOKeys, createdUTXOs)
	utxoRootHex := ActiveStorageManager.GetAccumulatorHex()

	fmt.Printf(
		"[Block %d | 1m] Size=%dB | Reward=%s | GasLimit=%d | Mode=%s | UTXORoot=%s\n",
		height,
		blockSize,
		minerReward.String(),
		nextGasLimit,
		ActiveStorageManager.Config.Mode,
		utxoRootHex,
	)

	return nextGasLimit, utxoRootHex
}
