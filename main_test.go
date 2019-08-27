package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

func TestCreateNetwork(t *testing.T) {
	errNotFound := &googleapi.Error{Code: http.StatusNotFound, Message: "boom!"}
	errConflict := &googleapi.Error{Code: http.StatusConflict, Message: "boom!"}

	type args struct {
		ctx     context.Context
		project string
		name    string
	}

	cases := map[string]struct {
		handler http.Handler
		args    args
		want    error
	}{
		"Successful": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Body.Close()

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&compute.Operation{})
			}),
			args: args{
				ctx:     context.Background(),
				project: "coolProject",
				name:    "coolNetwork",
			},
		},
		"ProjectDoesNotExist": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Body.Close()

				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(&ErrorReply{Error: errNotFound})
			}),
			args: args{
				ctx:     context.Background(),
				project: "coolProject",
				name:    "coolNetwork",
			},
			want: errNotFound,
		},
		"NetworkAlreadyExists": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Body.Close()

				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(&ErrorReply{Error: errConflict})
			}),
			args: args{
				ctx:     context.Background(),
				project: "coolProject",
				name:    "coolNetwork",
			},
			want: errConflict,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			svc, _ := compute.NewService(tc.args.ctx, option.WithEndpoint(srv.URL))

			err := CreateNetwork(tc.args.ctx, svc.Networks, tc.args.project, tc.args.name)

			if diff := cmp.Diff(tc.want, err, EquateErrors()); diff != "" {
				t.Errorf("CreateNetwork(): -want error, +got error:\n%s", diff)
			}
		})
	}
}

// https://github.com/googleapis/google-api-go-client/blob/d1c9f49/googleapi/googleapi.go#L112
type ErrorReply struct {
	Error *googleapi.Error `json:"error"`
}

// Copied from Crossplane.
func EquateErrors() cmp.Option {
	return cmp.Comparer(func(a, b error) bool {
		if a == nil || b == nil {
			return a == nil && b == nil
		}

		av := reflect.ValueOf(a)
		bv := reflect.ValueOf(b)
		if av.Type() != bv.Type() {
			return false
		}

		return a.Error() == b.Error()
	})
}
