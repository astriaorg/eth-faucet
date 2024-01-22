package server

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/LK4D4/trylock"
	"github.com/astriaorg/eth-faucet/internal/store"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"

	"github.com/astriaorg/eth-faucet/internal/chain"
	"github.com/astriaorg/eth-faucet/web"
)

type Server struct {
	mutex trylock.Mutex
	cfg   *Config
	queue chan claimRequest
	sm    store.RollupStoreManager
}

func NewServer(sm store.RollupStoreManager, cfg *Config) *Server {
	return &Server{
		cfg:   cfg,
		queue: make(chan claimRequest, cfg.queueCap),
		sm:    sm,
	}
}

func (s *Server) setupRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	api := r.PathPrefix("/api").Subrouter()
	limiter := NewLimiter(s.cfg.proxyCount, time.Duration(s.cfg.interval)*time.Minute)
	api.Handle("/claim", negroni.New(limiter, negroni.Wrap(s.handleClaim()))).Methods("POST")
	api.Handle("/info/{rollupName}", s.handleInfo()).Methods("GET")

	fs := http.FileServer(web.Dist())

	// NOTE - serving static files from /static_assets allows us to handle wildcard routes properly.
	//  requires vite.config.js `base` property to be set to the same pattern as below.
	// NOTE - rollup names don't support `_`, so using uncommon name `static_assets`
	r.PathPrefix("/static_assets").Handler(http.StripPrefix("/static_assets", fs))

	// serve the svelte app, via the filesystem, for any other route
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NOTE - http.FS has the same paths as the filesystem. if we didn't strip
		//  the path, then it would try to find a file with the same name as the path
		http.StripPrefix(r.URL.Path, fs).ServeHTTP(w, r)
	})

	return r
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
		claimRequest := <-s.queue

		txBuilder, err := s.txBuilderFromRollupName(claimRequest.RollupName)
		if err != nil {
			log.WithError(err).Error("Failed to create transaction builder while processing claim in queue")
		}

		txHash, err := txBuilder.Transfer(context.Background(), claimRequest.Address, chain.EtherToWei(int64(s.cfg.payout)))
		if err != nil {
			log.WithError(err).Error("Failed to handle transaction in the queue")
		} else {
			log.WithFields(log.Fields{
				"txHash":  txHash,
				"address": claimRequest.Address,
			}).Info("Consume from queue successfully")
		}
	}
}

func (s *Server) handleClaim() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The error always be nil since it has already been handled in limiter
		claimRequest, _ := readClaimRequest(r)
		// Try to lock mutex if the work queue is empty
		if len(s.queue) != 0 || !s.mutex.TryLock() {
			// TODO - need to store the rollup name with the address
			select {
			case s.queue <- claimRequest:
				log.WithFields(log.Fields{
					"address":    claimRequest.Address,
					"rollupName": claimRequest.RollupName,
				}).Info("Added to queue successfully")
				resp := claimResponse{Message: fmt.Sprintf("Added %s to the queue", claimRequest.Address)}
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

		txHash, err := txBuilder.Transfer(ctx, claimRequest.Address, chain.EtherToWei(int64(s.cfg.payout)))
		s.mutex.Unlock()
		if err != nil {
			log.WithError(err).Error("Failed to send transaction")
			_ = renderJSON(w, claimResponse{Message: err.Error()}, http.StatusInternalServerError)
			return
		}

		log.WithFields(log.Fields{
			"txHash":  txHash,
			"address": claimRequest.Address,
		}).Info("Funded directly successfully")
		resp := claimResponse{Message: fmt.Sprintf("Txhash: %s", txHash)}
		_ = renderJSON(w, resp, http.StatusOK)
	}
}

// handleInfo returns some details of the rollup.
// It fetches data from Firestore and returns it as JSON.
func (s *Server) handleInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		rollupName := vars["rollupName"]

		isValid := store.IsRollupNameValid(rollupName)
		if !isValid {
			msg := fmt.Sprintf("Invalid rollup name: %v", rollupName)
			log.Warn(msg)
			_ = renderJSON(w, errorResponse{Message: msg, Status: http.StatusBadRequest}, http.StatusBadRequest)
			return
		}

		rDoc, err := s.sm.FindRollupByName(rollupName)
		if err != nil {
			log.WithError(err).Warn("Failed to find rollup by name")
			_ = renderJSON(w, errorResponse{Message: err.Error(), Status: http.StatusInternalServerError}, http.StatusInternalServerError)
			return
		}

		if rDoc.Status != store.StatusDeployed {
			msg := "Rollup is not deployed"
			log.Warnf("%v: %v", msg, rDoc)
			_ = renderJSON(w, errorResponse{Message: "msg", Status: http.StatusInternalServerError}, http.StatusInternalServerError)
			return
		}

		_ = renderJSON(w, infoResponse{
			Payout:         strconv.Itoa(s.cfg.payout),
			FundingAddress: rDoc.RollupAccountAddress,
			RollupName:     rDoc.Name,
			NetworkID:      rDoc.NetworkID,
		}, http.StatusOK)
	}
}

// txBuilderFromRequest creates and returns a TxBuilder from the request
func (s *Server) txBuilderFromRequest(r *http.Request) (chain.TxBuilder, error) {
	claimRequest, err := readClaimRequest(r)
	if err != nil {
		err = fmt.Errorf("failed to read claim request: %w", err)
		return nil, err
	}

	return s.txBuilderFromRollupName(claimRequest.RollupName)
}

// txBuilderFromRollupName creates and returns a TxBuilder from the given name
func (s *Server) txBuilderFromRollupName(name string) (chain.TxBuilder, error) {
	rollup, err := s.sm.RollupByNameWithPrivate(name)
	if err != nil {
		err = fmt.Errorf("failed to find rollup by name: %w", err)
		return nil, err
	}

	hexkey := rollup.PrivateDetails.RollupAccountPrivKey
	if chain.Has0xPrefix(hexkey) {
		hexkey = hexkey[2:]
	}

	privKey, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		err = fmt.Errorf("failed to parse private key: %w", err)
		return nil, err
	}

	// FIXME - generate url from template string passed in as flag?
	rpcURL := fmt.Sprintf("http://%v.rpc.localdev.me", rollup.Name)
	txBuilder, err := chain.NewTxBuilder(rpcURL, privKey, big.NewInt(int64(rollup.NetworkID)))
	if err != nil {
		return nil, err
	}

	return txBuilder, nil
}
