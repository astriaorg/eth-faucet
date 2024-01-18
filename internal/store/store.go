package store

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type RollupStoreManager interface {
	FindRollupByName(name string) (RollupDoc, error)
}

type Manager struct {
	client                  *firestore.Client
	rollupsCollection       string
	usersCollection         string
	rollupPrivateCollection string
	rollupPrivateDoc        string
}

type NewManagerOpts struct {
	ProjectID string
}

func NewStoreManager(opts *NewManagerOpts) (*Manager, error) {
	ctx := context.Background()

	// NOTE - auth relies on GOOGLE_APPLICATION_CREDENTIALS envar being set
	client, err := firestore.NewClient(ctx, opts.ProjectID)
	if err != nil {
		return nil, err
	}

	return &Manager{
		client:                  client,
		rollupsCollection:       "rollups",
		usersCollection:         "users",
		rollupPrivateCollection: "private",
		rollupPrivateDoc:        "private",
	}, nil
}

// RollupDocumentStatus is a custom type for different document statuses
type RollupDocumentStatus string

// Enum values for RollupDocumentStatus
const (
	StatusCreated   RollupDocumentStatus = "CREATED"
	StatusDeploying RollupDocumentStatus = "DEPLOYING"
	StatusDeployed  RollupDocumentStatus = "DEPLOYED"
	StatusDeleted   RollupDocumentStatus = "DELETED"
	StatusError     RollupDocumentStatus = "ERROR"
)

type RollupDoc struct {
	Name           string               `firestore:"name"`
	NetworkID      uint32               `firestore:"networkId"`
	Status         RollupDocumentStatus `firestore:"status"`
	PrivateDetails RollupPrivateDoc     `firestore:"private"`
}

type RollupPrivateDoc struct {
	RollupAccountPrivKey    string `firestore:"rollupAccountPrivKey"`
	SequencerAccountPrivKey string `firestore:"sequencerAccountPrivKey"`
}

type RollupPublicDetails struct {
	RollupAccountAddress      string `firestore:"rollupAccountAddress"`
	RollupAccountPublicKey    string `firestore:"rollupAccountPublicKey"`
	SequencerAccountAddress   string `firestore:"sequencerAccountAddress"`
	SequencerAccountPublicKey string `firestore:"sequencerAccountPublicKey"`
}

// FindRollupByName queries the store to find a rollup with the given name
func (m *Manager) FindRollupByName(name string) (RollupDoc, error) {
	ctx := context.Background()

	iter := m.client.Collection(m.rollupsCollection).Where("name", "==", name).Documents(ctx)
	defer iter.Stop()
	var rollup RollupDoc
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return RollupDoc{}, err
		}
		err = doc.DataTo(&rollup)
		if err != nil {
			return RollupDoc{}, err
		}
	}
	return rollup, nil
}
