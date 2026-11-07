package main

import (
	"time"

	"://github.com"
	"://github.com"
	"://github.com"
	"://github.com"
)

var (
	// DefaultNodeHome configures the default data directory inside user space
	DefaultNodeHome = "/.mycustoml1"
	
	// ModuleBasics initializes the core asset routing and account systems
	ModuleBasics    = module.NewBasicManager(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
	)
)

type App struct {
	LegacyAmino *codec.LegacyAmino
	AppCodec    codec.Codec
}

// GetBlockTimeTarget tells the consensus engine to automatically adjust the 
// mining difficulty so that a new block is produced exactly every 30 seconds.
func GetBlockTimeTarget() time.Duration {
	return 30 * time.Second
}
