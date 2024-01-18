package server

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/astriaorg/eth-faucet/internal/store"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/LK4D4/trylock"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"

	"github.com/astriaorg/eth-faucet/internal/chain"
	"github.com/astriaorg/eth-faucet/web"
)

type Server struct {
	mutex trylock.Mutex
	cfg   *Config
	queue chan string
	sm    store.RollupStoreManager
}

func NewServer(sm store.RollupStoreManager, cfg *Config) *Server {
	return &Server{
		cfg:   cfg,
		queue: make(chan string, cfg.queueCap),
		sm:    sm,
	}
}

func (s *Server) setupRouter() *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/", http.FileServer(web.Dist()))
	limiter := NewLimiter(s.cfg.proxyCount, time.Duration(s.cfg.interval)*time.Minute)
	router.Handle("/api/claim", negroni.New(limiter, negroni.Wrap(s.handleClaim())))
	router.Handle("/api/info", s.handleInfo())

	return router
}

func (s *Server) Run() {
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			s.consumeQueue()
		}
	}()

	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(s.setupRouter())
	log.Infof("Starting http server %d", s.cfg.httpPort)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(s.cfg.httpPort), n))
}

func (s *Server) consumeQueue() {
	if len(s.queue) == 0 {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	for len(s.queue) != 0 {
		address := <-s.queue

		// TODO - need to store the rollup name with the address
		name := "bugbug-fake-rollup-name"
		txBuilder, err := s.txBuilderFromRollupName(name)
		if err != nil {
			log.WithError(err).Error("Failed to create transaction builder while processing claim in queue")
		}

		txHash, err := txBuilder.Transfer(context.Background(), address, chain.EtherToWei(int64(s.cfg.payout)))
		if err != nil {
			log.WithError(err).Error("Failed to handle transaction in the queue")
		} else {
			log.WithFields(log.Fields{
				"txHash":  txHash,
				"address": address,
			}).Info("Consume from queue successfully")
		}
	}
}

func (s *Server) handleClaim() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}

		// The error always be nil since it has already been handled in limiter
		address, _ := readAddress(r)
		// Try to lock mutex if the work queue is empty
		if len(s.queue) != 0 || !s.mutex.TryLock() {
			// TODO - need to store the rollup name with the address
			select {
			case s.queue <- address:
				log.WithFields(log.Fields{
					"address": address,
				}).Info("Added to queue successfully")
				resp := claimResponse{Message: fmt.Sprintf("Added %s to the queue", address)}
				_ = renderJSON(w, resp, http.StatusOK)
			default:
				log.Warn("Max queue capacity reached")
				_ = renderJSON(w, claimResponse{Message: "Faucet queue is too long, please try again later"}, http.StatusServiceUnavailable)
			}
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		txBuilder, err := s.txBuilderFromRequest(r)
		if err != nil {
			log.WithError(err).Error("Failed to create transaction builder")
			_ = renderJSON(w, claimResponse{Message: err.Error()}, http.StatusInternalServerError)
			return
		}

		txHash, err := txBuilder.Transfer(ctx, address, chain.EtherToWei(int64(s.cfg.payout)))
		s.mutex.Unlock()
		if err != nil {
			log.WithError(err).Error("Failed to send transaction")
			_ = renderJSON(w, claimResponse{Message: err.Error()}, http.StatusInternalServerError)
			return
		}

		log.WithFields(log.Fields{
			"txHash":  txHash,
			"address": address,
		}).Info("Funded directly successfully")
		resp := claimResponse{Message: fmt.Sprintf("Txhash: %s", txHash)}
		_ = renderJSON(w, resp, http.StatusOK)
	}
}

func (s *Server) handleInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}

		_ = renderJSON(w, infoResponse{
			Payout: strconv.Itoa(s.cfg.payout),
		}, http.StatusOK)
	}
}

// txBuilderFromRequest creates and returns a TxBuilder from the request
func (s *Server) txBuilderFromRequest(r *http.Request) (chain.TxBuilder, error) {
	claimRequest, err := readClaimRequest(r)
	if err != nil {
		return nil, err
	}

	return s.txBuilderFromRollupName(claimRequest.RollupName)
}

// txBuilderFromRollupName creates and returns a TxBuilder from the given name
func (s *Server) txBuilderFromRollupName(name string) (chain.TxBuilder, error) {
	rollup, err := s.sm.FindRollupByName(name)
	if err != nil {
		return nil, err
	}

	hexkey := rollup.PrivateDetails.RollupAccountPrivKey
	if chain.Has0xPrefix(hexkey) {
		hexkey = hexkey[2:]
	}

	privKey, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		return nil, err
	}

	// TODO - generate proper rpc url
	rpcURL := fmt.Sprintf("https://rollups.%v.rpc.blahblah.com", rollup.Name)
	txBuilder, err := chain.NewTxBuilder(rpcURL, privKey, big.NewInt(int64(rollup.NetworkID)))
	if err != nil {
		return nil, err
	}

	return txBuilder, nil
}
