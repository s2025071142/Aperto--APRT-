package main

import (
	"bytes"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/opd-ai/go-randomx"
)

// GlobalChainhasher stores the pre-allocated 2GB RandomX dataset to prevent initialization freezes.
type BlockchainVerifier struct {
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
		CurrentSeed: initialSeed,
	}, nil
}

// VerifyRandomXProof uses the pre-allocated verifier instance to validate blocks in milliseconds.
func (bv *BlockchainVerifier) VerifyRandomXProof(blockData []byte, nonce uint64, targetLimit *big.Int, seedHash []byte) (bool, error) {
	// 1. Efficiently update the cache dataset only if the epoch seed has changed
	if !bytes.Equal(bv.CurrentSeed, seedHash) {
		if err := bv.Hasher.UpdateCacheKey(seedHash); err != nil {
			return false, errors.New("failed to transition RandomX seed epoch")
		}
		bv.CurrentSeed = seedHash
	}

	// 2. Format the payload inputs
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	inputData := append(blockData, nonceBytes...)

	// 3. Compute the cryptographic hash instantly using the pre-pooled dataset
	resultHash := bv.Hasher.Hash(inputData)

	// 4. Convert the full 32-byte matrix into a big integer for absolute safety checks
	hashInt := new(big.Int).SetBytes(resultHash[:])

	// 5. Check target bounds: The numeric hash value must be less than or equal to the target limit
	if hashInt.Cmp(targetLimit) <= 0 {
		return true, nil
	}

	return false, nil
}

