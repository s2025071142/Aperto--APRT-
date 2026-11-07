package main

// AdjustDynamicGasLimit calculates a moving target for the EVM execution engine.
// It dynamically increases capacity during congestion and prevents network spam.
func AdjustDynamicGasLimit(historicalAverageGas uint64) uint64 {
	const TargetGasLimit uint64 = 30000000 
	const MaxGrowthFactor float64 = 2.0 

	// Base case: Traffic is normal or below target limits
	if historicalAverageGas <= TargetGasLimit {
		return TargetGasLimit
	}

	// Calculate the surge ratio of network demand
	demandRatio := float64(historicalAverageGas) / float64(TargetGasLimit)
	
	// Enforce growth boundary to stabilize maximum execution footprints
	if demandRatio > MaxGrowthFactor {
		demandRatio = MaxGrowthFactor
	}

	// Scale gas limits proportionally to transaction congestion
	dynamicGasLimit := uint64(float64(TargetGasLimit) * demandRatio)

	return dynamicGasLimit
}
