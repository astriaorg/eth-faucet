package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/astriaorg/eth-faucet/internal/server"
	"github.com/astriaorg/eth-faucet/internal/store"
)

var (
	appVersion = "v2.0.0"

	httpPortFlag = flag.Int("httpport", 8080, "Listener port to serve HTTP connection")
	proxyCntFlag = flag.Int("proxycount", 0, "Count of reverse proxies in front of the server")
	queueCapFlag = flag.Int("queuecap", 100, "Maximum transactions waiting to be sent")
	versionFlag  = flag.Bool("version", false, "Print version number")

	payoutFlag   = flag.Int("faucet.amount", 1, "Number of Ethers to transfer per user request")
	intervalFlag = flag.Int("faucet.minutes", 1440, "Number of minutes to wait between funding rounds")
	netnameFlag  = flag.String("faucet.name", "testnet", "Network name to display on the frontend")

	firestoreProjectID = flag.String("firestoreprojectid", "some-proj-id", "The Firestore project id.")
)

func init() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(appVersion)
		os.Exit(0)
	}
}

func Execute() {
	smOpts := &store.NewManagerOpts{
		ProjectID: *firestoreProjectID,
	}
	sm, err := store.NewStoreManager(smOpts)
	if err != nil {
		fmt.Printf("Failed to create store manager: %v\n", err)
		os.Exit(1)
	}

	config := server.NewConfig(*netnameFlag, *httpPortFlag, *intervalFlag, *payoutFlag, *proxyCntFlag, *queueCapFlag)
	go server.NewServer(sm, config).Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
