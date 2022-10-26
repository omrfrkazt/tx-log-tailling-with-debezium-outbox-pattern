package endpoints

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"testapp/internal/service"
	authdb "testapp/internal/store"
)

// Endpoints represents service endpoints
type Endpoints struct {
	UserAddEndpoint endpoint.Endpoint
}

// MakeEndpoints makes and returns endpoints
func MakeEndpoints(s service.Service) Endpoints {
	return Endpoints{
		UserAddEndpoint: MakeUserAddEndpoint(s),
	}
}

// MakeUserAddEndpoint returns user endpoint
func MakeUserAddEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*authdb.CreateUserParams)
		res, _ := s.AddUser(ctx, *req)
		return res, nil
	}
}
