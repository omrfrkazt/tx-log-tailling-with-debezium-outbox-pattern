package service

import (
	"context"
	"encoding/json"
	"github.com/go-kit/log"
	authdb "testapp/internal/store"
)

type Service interface {
	AddUser(ctx context.Context, user authdb.CreateUserParams) (*authdb.User, error)
}
type service struct {
	db *authdb.Queries
	l  log.Logger
}

func NewService(l log.Logger, db *authdb.Queries) Service {
	return &service{
		db: db,
		l:  l,
	}
}

func (s *service) AddUser(ctx context.Context, user authdb.CreateUserParams) (*authdb.User, error) {
	res, err := s.db.CreateUser(ctx, user)
	if err == nil {
		payload, _ := json.Marshal(res)
		_, err = s.db.AddOutbox(ctx, authdb.AddOutboxParams{
			Aggregatetype: "CreateUser",
			Aggregateid:   "create_user",
			Type:          "CREATE",
			Payload:       string(payload),
		})
		if err != nil {
			s.l.Log(err)
			return nil, err
		}
		return &res, nil
	}
	s.l.Log(err)
	return nil, err
}
