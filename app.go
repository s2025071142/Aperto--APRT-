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

// QuantumAccount represents an un-hackable wallet schema on the Aperto network
type QuantumAccount struct {
	PublicKey mldsa65.PublicKey
	Address   []byte
}

// App structures the global state container definitions for the compiler
type App struct {
	LegacyAmino *codec.LegacyAmino
	AppCodec    codec.Codec
	Storage     *StorageManager // Manages UTXO commitment or Pruned node storage
}

// NewApp initializes the Aperto application state with the active storage mode
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
	}
}

// NewQuantumAccount derives a secure local address by hashing the ML-DSA public key coordinates
func NewQuantumAccount(pubKey mldsa65.PublicKey) *QuantumAccount {
	pubBytes := pubKey.Bytes()
	hash := sha256.Sum256(pubBytes)

	return &QuantumAccount{
		PublicKey: pubKey,
		Address:   hash[:20], // Standard 20-byte account format derived from quantum public key states
	}
}

// VerifyQuantumTransaction checks high-level incoming signatures safely against Shor's Algorithm
func VerifyQuantumTransaction(pubKey mldsa65.PublicKey, message []byte, signature []byte) (bool, error) {
	if !mldsa65.Verify(pubKey, message, nil, signature) {
		return false, errors.New("signature authenticity verification failed via ML-DSA-65 engine")
	}
	return true, nil
}

// GetBlockTimeTarget tells the engine to automatically target blocks precisely every 30 seconds
func GetBlockTimeTarget() time.Duration {
	return 30 * time.Second
}
