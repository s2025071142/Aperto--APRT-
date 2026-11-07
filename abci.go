package main

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
