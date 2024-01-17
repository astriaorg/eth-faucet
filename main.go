package main

import (
	"github.com/astriaorg/eth-faucet/cmd"
)

//go:generate npm run build-web
func main() {
	cmd.Execute()
}
