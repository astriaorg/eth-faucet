package store

import (
	"context"
	"errors"
	"regexp"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type RollupStoreManager interface {
	Rollups() ([]RollupDoc, error)
	FindRollupByName(name string) (RollupDoc, error)
	RollupByNameWithPrivate(name string) (RollupDoc, error)
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
	ID        string               `firestore:"id" json:"id"`
	Name      string               `firestore:"name" json:"name"`
	NetworkID uint32               `firestore:"networkId" json:"networkId"`
	Status    RollupDocumentStatus `firestore:"status" json:"status"`

	RollupPublicDetails
	PrivateDetails RollupPrivateDoc
}

type RollupPrivateDoc struct {
	RollupAccountPrivKey    string `firestore:"rollupAccountPrivKey"`
	SequencerAccountPrivKey string `firestore:"sequencerAccountPrivKey"`
}

type RollupPublicDetails struct {
	RollupAccountAddress      string `firestore:"rollupAccountAddress" json:"rollupAccountAddress"`
	RollupAccountPublicKey    string `firestore:"rollupAccountPublicKey" json:"rollupAccountPublicKey"`
	SequencerAccountAddress   string `firestore:"sequencerAccountAddress" json:"sequencerAccountAddress"`
	SequencerAccountPublicKey string `firestore:"sequencerAccountPublicKey" json:"sequencerAccountPublicKey"`
}

// Rollups queries the store to find all deployed rollups
func (m *Manager) Rollups() ([]RollupDoc, error) {
	ctx := context.Background()

	iter := m.client.CollectionGroup(m.rollupsCollection).Where("status", "==", StatusDeployed).Documents(ctx)
	defer iter.Stop()

	var rollups []RollupDoc
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return []RollupDoc{}, err
		}

		var rollup RollupDoc
		if err := doc.DataTo(&rollup); err != nil {
			return []RollupDoc{}, err
		}
		rollup.ID = doc.Ref.ID
		rollups = append(rollups, rollup)
	}

	return rollups, nil
}

// FindRollupByName queries the store to find a rollup with the given name
func (m *Manager) FindRollupByName(name string) (RollupDoc, error) {
	snapshot, err := m.RollupDocSnapshotByName(name)
	if err != nil {
		return RollupDoc{}, err
	}

	var rollup RollupDoc
	if err := snapshot.DataTo(&rollup); err != nil {
		return RollupDoc{}, err
	}
	rollup.ID = snapshot.Ref.ID
	return rollup, nil
}

// RollupByNameWithPrivate finds a rollup by name and returns it along with its private details
func (m *Manager) RollupByNameWithPrivate(name string) (RollupDoc, error) {
	snapshot, err := m.RollupDocSnapshotByName(name)
	if err != nil {
		return RollupDoc{}, err
	}
	var rollup RollupDoc
	if err := snapshot.DataTo(&rollup); err != nil {
		return RollupDoc{}, err
	}

	privateDoc, err := snapshot.Ref.Collection(m.rollupPrivateCollection).Doc(m.rollupPrivateDoc).Get(context.Background())
	if err != nil {
		return RollupDoc{}, err
	}
	var priv RollupPrivateDoc
	if err := privateDoc.DataTo(&priv); err != nil {
		return RollupDoc{}, err
	}

	rollup.ID = snapshot.Ref.ID
	rollup.PrivateDetails = priv
	return rollup, nil
}

// RollupDocSnapshotByName queries the store to find a rollup with the given name
func (m *Manager) RollupDocSnapshotByName(name string) (*firestore.DocumentSnapshot, error) {
	ctx := context.Background()

	iter := m.client.CollectionGroup(m.rollupsCollection).Where("name", "==", name).Where("status", "==", StatusDeployed).Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			return nil, errors.New("rollup not found")
		}
		if err != nil {
			return nil, err
		}
		return doc, nil
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
