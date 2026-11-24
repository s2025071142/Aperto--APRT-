package main

import (
	"crypto/sha256"
	"errors"
	"time"

	"://github.com"
	"://github.com"
	"://github.com"
	"://github.com"
	"://github.com"
)

var (
	// DefaultNodeHome configures the data state directory for the Aperto network daemon
	DefaultNodeHome = "/.aperto-aprt"

	// ModuleBasics initializes the underlying accounting and token frameworks safely
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
	)
)

// QuantumAccount represents a post-quantum wallet schema on the Aperto network
type QuantumAccount struct {
	PublicKey mldsa65.PublicKey
	Address   []byte
}

// App structures the global state container definitions for the compiler
type App struct {
	LegacyAmino *codec.LegacyAmino
	AppCodec    codec.Codec
	Storage     *StorageManager // Manages UTXO commitment or Pruned node storage
	BlockLock   *BlockLockState // Enforces < 2s pre-sign block lock-in and halts duplicate mining
}

// NewApp initializes the Aperto application state with the active storage mode and block lock state
func NewApp(legacyAmino *codec.LegacyAmino, appCodec codec.Codec) *App {
	// Ensure ActiveStorageManager is initialized even if launched outside main.go CLI flags
	if ActiveStorageManager == nil {
		defaultCfg, _ := NewNodeStorageConfig(string(DetectDefaultStorageMode()), 1000)
		ActiveStorageManager = NewStorageManager(defaultCfg)
	}

	return &App{
		LegacyAmino: legacyAmino,
		AppCodec:    appCodec,
		Storage:     ActiveStorageManager,
		BlockLock:   ActiveBlockLock,
	}
}

// NewQuantumAccount derives a secure local address by hashing the ML-DSA-65 public key coordinates
func NewQuantumAccount(pubKey mldsa65.PublicKey) *QuantumAccount {
	pubBytes := pubKey.Bytes()
	hash := sha256.Sum256(pubBytes)

	return &QuantumAccount{
		PublicKey: pubKey,
		Address:   hash[:20], // Standard 20-byte account format derived from quantum public key states
	}
}

// VerifyQuantumTransaction checks incoming signatures against Shor's Algorithm using ML-DSA-65
func VerifyQuantumTransaction(pubKey mldsa65.PublicKey, message []byte, signature []byte) (bool, error) {
	if !mldsa65.Verify(pubKey, message, nil, signature) {
		return false, errors.New("signature authenticity verification failed via ML-DSA-65 engine")
	}
	return true, nil
}

// VerifyAndLockQuantumBlock verifies a candidate block's ML-DSA-65 pre-signature under 2 seconds,
// locks the height, and immediately cancels ongoing CPU hashing for that height.
func (a *App) VerifyAndLockQuantumBlock(
	height uint64,
	blockHash string,
	headerBytes []byte,
	preSig []byte,
	proposerPubKey mldsa65.PublicKey,
) error {
	start := time.Now()

	if a.BlockLock.IsHeightLocked(height) {
		return errors.New("block height already locked in; halting duplicate hashing")
	}

	valid, err := VerifyQuantumTransaction(proposerPubKey, headerBytes, preSig)
	if err != nil || !valid {
		return errors.New("quantum pre-sign verification failed")
	}

	elapsed := time.Since(start)
	if elapsed >= PreSignLockTimeout {
		return errors.New("quantum pre-sign verification exceeded 2-second limit")
	}

	a.BlockLock.mu.Lock()
	if a.BlockLock.lockedHeight < height {
		a.BlockLock.lockedHeight = height
		a.BlockLock.lockedHash = blockHash
		if a.BlockLock.cancelMining != nil {
			a.BlockLock.cancelMining()
			a.BlockLock.cancelMining = nil
		}
	}
	a.BlockLock.mu.Unlock()

	return nil
}

// GetBlockTimeTarget enforces the 1-minute (60s) target block interval across the network
func GetBlockTimeTarget() time.Duration {
	return TargetBlockTime // 60 * time.Second
}
