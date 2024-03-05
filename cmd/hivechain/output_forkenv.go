package main

import (
	"fmt"
)

// writeForkEnv writes chain fork configuration in the form that hive expects.
func (g *generator) writeForkEnv() error {
	cfg := g.genesis.Config
	env := make(map[string]string)

	// basic settings
	env["HIVE_NETWORK_ID"] = fmt.Sprint(cfg.NetworkID)

	return g.writeJSON("forkenv.json", env)
}
