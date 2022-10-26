package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/iris-contrib/schema"
	"net/http"
	"reflect"
	"testapp"
	"testapp/internal/endpoints"
	"testapp/internal/service"
	authdb "testapp/internal/store"
	"testapp/internal/transport"
)

// endpoint names
const (
	CreateUser = "CreateUser"
)

// decoder tags
const (
	headerTag = "header"
	queryTag  = "query"
)

const invalidResponseError = "invalid response"

// MakeHTTPHandler makes and returns http handler
func MakeHTTPHandler(l log.Logger, s service.Service) http.Handler {
	es := endpoints.MakeEndpoints(s)

	r := mux.NewRouter()

	// Create User - POST - /user
	r.Methods("POST").Path("/user").Handler(
		makeCreateUserHandler(es.UserAddEndpoint, makeDefaultServerOptions(l, CreateUser)),
	)

	return r
}

func makeCreateUserHandler(e endpoint.Endpoint, serverOptions []kithttp.ServerOption) http.Handler {
	h := kithttp.NewServer(e, makeDecoder(authdb.CreateUserParams{}), encoder, serverOptions...)

	return h
}
func makeDefaultServerOptions(l log.Logger, endpointName string) []kithttp.ServerOption {
	return []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewErrorHandler(l, endpointName)),
	}
}

func makeDecoder(emptyReq interface{}) kithttp.DecodeRequestFunc {
	return func(_ context.Context, r *http.Request) (interface{}, error) {
		req := reflect.New(reflect.TypeOf(emptyReq)).Interface()

		if err := newHeaderDecoder().Decode(req, r.Header); err != nil {
			return nil, fmt.Errorf("decoding request header failed, %s", err.Error())
		}

		if err := newQueryDecoder().Decode(req, r.URL.Query()); err != nil {
			return nil, fmt.Errorf("decoding request query failed, %s", err.Error())
		}

		if requestHasBody(r) {
			if err := json.NewDecoder(r.Body).Decode(req); err != nil {
				return nil, fmt.Errorf("decoding request body failed, %s", err.Error())
			}
		}

		if err := validate(req); err != nil {
			return nil, err
		}

		return req, nil
	}
}

func newHeaderDecoder() *schema.Decoder {
	return newDecoder(headerTag)
}

func newQueryDecoder() *schema.Decoder {
	return newDecoder(queryTag)
}

func newDecoder(tag string) *schema.Decoder {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	if tag != "" {
		decoder.SetAliasTag(tag)
	}

	return decoder
}

func requestHasBody(r *http.Request) bool {
	return r.Body != http.NoBody
}

func validate(req interface{}) error {
	errs := validator.New().Struct(req)
	if errs == nil {
		return nil
	}

	firstErr := errs.(validator.ValidationErrors)[0]

	return errors.New("validation failed, tag: " + firstErr.Tag() + ", field: " + firstErr.Field())
}

func encoder(_ context.Context, rw http.ResponseWriter, response interface{}) error {
	r, ok := response.(testapp.Response)
	if !ok {
		return errors.New(invalidResponseError)
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.WriteHeader(http.StatusOK)
	return json.NewEncoder(rw).Encode(r)
}
