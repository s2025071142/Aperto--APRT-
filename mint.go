package main

import (
	"://github.com"
)

// CalculateBlockSubsidy determines the miner reward for the current block height.
// It follows a deterministic 20-year upside-down U (bell curve) that peaks at year 10,
// then tapers down smoothly until locking into a permanent 0.005 APTO tail emission floor.
func CalculateBlockSubsidy(height int64) types.Coin {
	// Base units representation (1 APTO = 1,000,000 base units)
	const BaseUnit uint64 = 1000000
	const PeakSubsidy uint64 = 50 * BaseUnit  // Maximum reward at the 10-year midpoint
	const TailEmissionFloor uint64 = 5000     // 0.005 * 1,000,000 = 5,000 base units

	// Assuming ~2-minute block times (~262,800 blocks/year):
	// 20-year bell curve spans 5,256,000 blocks, peaking at block 2,628,000 (Year 10).
	const CurveDurationBlocks int64 = 5256000
	const PeakHeight int64 = CurveDurationBlocks / 2

	// After the 20-year curve completes (or on invalid negative heights), lock at tail floor
	if height <= 0 || height >= CurveDurationBlocks {
		return types.NewCoin("native", types.NewInt(int64(TailEmissionFloor)))
	}

	// Deterministic upside-down U formula using Cosmos SDK fixed-point decimals (no floats):
	// Progress factor: 1 - ((height - PeakHeight) / PeakHeight)^2
	// Equals 0 at block 0, rises smoothly to 1.0 at PeakHeight (Year 10), and returns to 0 at Year 20.
	hDec := types.NewDec(height)
	peakDec := types.NewDec(PeakHeight)

	offset := hDec.Sub(peakDec).Quo(peakDec)
	bellFactor := types.OneDec().Sub(offset.Mul(offset))

	// Scale the reward range between TailEmissionFloor and PeakSubsidy
	rewardRange := types.NewDec(int64(PeakSubsidy - TailEmissionFloor))
	scaledBonus := rewardRange.Mul(bellFactor).TruncateInt64()

	currentSubsidy := TailEmissionFloor + uint64(scaledBonus)

	// Enforce the permanent minimum tail emission subsidy floor
	if currentSubsidy < TailEmissionFloor {
		currentSubsidy = TailEmissionFloor
	}

	return types.NewCoin("native", types.NewInt(int64(currentSubsidy)))
}
