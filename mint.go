package main

import (
	"://github.com"
)

// CalculateBlockSubsidy determines the miner reward for the current block height.
// It decreases the reward gradually with every single block until hitting the tail floor.
func CalculateBlockSubsidy(height int64) types.Coin {
	// Base units representation (e.g., 1 coin = 1,000,000 base units)
	const BaseUnit uint64 = 1000000
	const InitialSubsidy uint64 = 50 * BaseUnit
	const TailEmissionFloor uint64 = 5000 // 0.005 * 1,000,000 = 5,000 base units

	// Deterministic decay calculation using integer approximation:
	// Instead of floating-point math.Exp(-0.000005 * height), we use a fixed precision reduction loop.
	// For a decay factor of 0.000005 per block, the remaining subsidy shrinks by factor of (1 - 5/1000000).
	decayBasis := types.NewDec(999995).Quo(types.NewDec(1000000))
	
	// Compute exponential decay using Cosmos SDK's deterministic fixed-point Power function
	remainingFactor := decayBasis.Power(uint64(height))
	
	// Calculate the actual current base subsidy integer value safely
	currentSubsidyDec := types.NewDec(int64(InitialSubsidy)).Mul(remainingFactor)
	currentSubsidy := uint64(currentSubsidyDec.TruncateInt64())

	// Enforce the permanent minimum tail emission subsidy floor
	if currentSubsidy < TailEmissionFloor {
		currentSubsidy = TailEmissionFloor
	}

	return types.NewCoin("native", types.NewInt(int64(currentSubsidy)))
}
