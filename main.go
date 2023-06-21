package main

import (
	"github.com/chainflag/eth-faucet/cmd"
)

//go:generate npm run build-web
func main() {
	cmd.Execute()
}
