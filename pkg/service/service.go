package service

import (
	"context"

	"github.com/jeffreybolle/inttest-example-go/pkg/api"
	"github.com/jeffreybolle/inttest-example-go/pkg/creditscore"
	"github.com/jeffreybolle/inttest-example-go/pkg/store"
)

type Service struct {
	store *store.Store
	cs    *creditscore.CreditStore
}

func NewService(s *store.Store, cs *creditscore.CreditStore) *Service {
	return &Service{
		store: s,
		cs:    cs,
	}
}

func (s *Service) CreateUser(ctx context.Context, req *api.CreateUserRequest) (*api.CreateUserResponse, error) {
	userID, err := s.store.CreateUser(ctx, req.FirstName, req.LastName, req.DateOfBirth)
	if err != nil {
		return nil, err
	}
	return &api.CreateUserResponse{
		ID: userID,
	}, nil
}

func (s *Service) GetUser(ctx context.Context, req *api.GetUserRequest) (*api.GetUserResponse, error) {
	user, err := s.store.GetUser(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	score, err := s.cs.GetScore(ctx, user.FirstName, user.LastName)
	if err != nil {
		return nil, err
	}
	return &api.GetUserResponse{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DateOfBirth,
		CreditScore: score,
	}, nil
}
