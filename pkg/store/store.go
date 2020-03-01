package store

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"github.com/jeffreybolle/inttest-example-go/pkg/model"
)

var (
	ErrNoSuchEntity = datastore.ErrNoSuchEntity
)

type Store struct {
	c *datastore.Client
}

func NewStore(ctx context.Context, projectID string) (*Store, error) {
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error while creating store: %v", err)
	}
	return &Store{
		c: client,
	}, nil
}

func (s *Store) CreateUser(ctx context.Context, firstName, lastName string, dob time.Time) (string, error) {
	id := mintID()
	key := userKey(id)
	user := model.User{
		ID:          key,
		FirstName:   firstName,
		LastName:    lastName,
		DateOfBirth: dob,
	}
	_, err := s.c.Put(ctx, key, &user)
	if err != nil {
		return "", fmt.Errorf("error while writing to datastore: %v", err)
	}
	return id, nil
}

func (s *Store) GetUser(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	key := userKey(id)
	err := s.c.Get(ctx, key, &user)
	if err == datastore.ErrNoSuchEntity {
		return nil, ErrNoSuchEntity
	}
	if err != nil {
		return nil, fmt.Errorf("error while reading from datatore: %v", err)
	}
	return &user, nil
}

func userKey(id string) *datastore.Key {
	return datastore.NameKey("User", id, nil)
}

func mintID() string {
	return uuid.New().String()
}
