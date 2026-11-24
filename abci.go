package main

import "fmt"

// AdjustDynamicGasLimit calculates a moving capacity ceiling for the EVM block window.
// It expands drastically during burst loads to absorb heavy post-quantum cryptographic byte footprints.
func AdjustDynamicGasLimit(historicalAverageGas uint64) uint64 {
	// Expanded limits to safely process large data chunks without network congestion
	const TargetGasLimit uint64 = 90000000
	const MaxGrowthFactor float64 = 4.0

	// Return baseline scale if the transaction congestion is running optimally
	if historicalAverageGas <= TargetGasLimit {
		return TargetGasLimit
	}

	// Compute scaling coefficients accurately based on payload pressures
	demandRatio := float64(historicalAverageGas) / float64(TargetGasLimit)

	// Caps limits strictly to ensure block computation footprints remain sustainable for nodes
	if demandRatio > MaxGrowthFactor {
		demandRatio = MaxGrowthFactor
	}

	// Generate the fresh elastic transaction resource parameter
	dynamicGasLimit := uint64(float64(TargetGasLimit) * demandRatio)

	return dynamicGasLimit
}

// FinalizeAndPruneBlock executes dynamic gas scaling alongside UTXO commitment updates
// and historical block pruning at the end of every block.
func FinalizeAndPruneBlock(
	height uint64,
	historicalAverageGas uint64,
	rawBlockBytes []byte,
	spentUTXOKeys []string,
	createdUTXOs []UTXO,
) (uint64, string) {
	// 1. Calculate the dynamic block gas limit for the next window
	nextGasLimit := AdjustDynamicGasLimit(historicalAverageGas)

	// 2. Fallback initialization if storage manager wasn't set by CLI
	if ActiveStorageManager == nil {
		defaultCfg, _ := NewNodeStorageConfig(string(DetectDefaultStorageMode()), 1000)
		ActiveStorageManager = NewStorageManager(defaultCfg)
	}

	// 3. Apply UTXO state transitions and prune/discard old blocks based on node mode
	ActiveStorageManager.ApplyBlockState(height, rawBlockBytes, spentUTXOKeys, createdUTXOs)
	utxoRootHex := ActiveStorageManager.GetAccumulatorHex()

	fmt.Printf(
		"[Block %d] GasLimit=%d | StorageMode=%s | UTXORoot=%s\n",
		height,
		nextGasLimit,
		ActiveStorageManager.Config.Mode,
		utxoRootHex,
	)

	return nextGasLimit, utxoRootHex
}
