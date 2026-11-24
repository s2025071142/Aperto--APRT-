package main

import (
	"math/big"

	"://github.com"
)

// CalculateBlockSubsidy determines the exact miner reward for a 1-minute block height.
// Over the first 20 years (10,519,200 blocks), it follows a deterministic upside-down U
// (bell curve) that mints ~21,000,000 APTO total, peaking at ~2.992 APTO/block at Year 10.
// After Year 20, it locks into a permanent 0.005 APTO tail emission floor forever.
func CalculateBlockSubsidy(height int64) types.Coin {
	// Base units representation: 1 APTO = 1,000,000 base units
	const BaseUnit int64 = 1000000
	const TailEmissionFloor int64 = 5000 // 0.005 APTO = 5,000 base units

	// 1-minute block time = 525,960 blocks/year (accounting for leap years)
	// 20-year bell curve duration = 10,519,200 blocks
	const CurveDurationBlocks int64 = 10519200
	const PeakHeight int64 = CurveDurationBlocks / 2 // Block 5,259,600 (Year 10)

	// Peak bell bonus calibrated so the 20-year integral equals exactly 21,000,000 APTO
	// Peak total reward = 2,987,024 (bonus) + 5,000 (tail) = 2,992,024 base units (~2.992 APTO)
	const PeakBellBonus int64 = 2987024

	// Before block 1 or once the 20-year curve completes, return the permanent 0.005 APTO tail floor
	if height <= 0 || height >= CurveDurationBlocks {
		return types.NewCoin("apto", types.NewInt(TailEmissionFloor))
	}

	// Deterministic exact integer parabola using math/big (prevents 64-bit integer overflow):
	// bonus = PeakBellBonus * height * (CurveDurationBlocks - height) / (PeakHeight^2)
	h := big.NewInt(height)
	rem := big.NewInt(CurveDurationBlocks - height)
	peakBonus := big.NewInt(PeakBellBonus)

	peakH := big.NewInt(PeakHeight)
	peakHSquared := new(big.Int).Mul(peakH, peakH)

	numerator := new(big.Int).Mul(peakBonus, h)
	numerator.Mul(numerator, rem)

	bonus := new(big.Int).Div(numerator, peakHSquared).Int64()
	totalSubsidy := TailEmissionFloor + bonus

	return types.NewCoin("apto", types.NewInt(totalSubsidy))
}
