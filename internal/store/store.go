package store

import (
	"context"
	"errors"
	"regexp"

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
	ID        string               `firestore:"id"`
	Name      string               `firestore:"name"`
	NetworkID uint32               `firestore:"networkId"`
	Status    RollupDocumentStatus `firestore:"status"`

	RollupPublicDetails
	// FIXME - private isn't a field on the doc but is another collection, so this won't work
	PrivateDetails RollupPrivateDoc `firestore:"private"`
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

	iter := m.client.CollectionGroup(m.rollupsCollection).Where("name", "==", name).Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			return RollupDoc{}, errors.New("rollup not found")
		}
		if err != nil {
			return RollupDoc{}, err
		}

		var rollup RollupDoc
		err = doc.DataTo(&rollup)
		if err != nil {
			return RollupDoc{}, err
		}
		if rollup.Name != "" {
			rollup.ID = doc.Ref.ID
			return rollup, nil
		}
	}
}

// IsRollupNameValid checks against a regex to ensure the rollup name is valid
func IsRollupNameValid(name string) bool {
	pattern := "^[a-z]+[a-z0-9]*(?:-[a-z0-9]+)*$"
	matched, err := regexp.MatchString(pattern, name)
	if err != nil {
		return false
	}
	if !matched {
		return false
	}
	return true
}
