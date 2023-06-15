// external_test.go

package external_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"mocking/external"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/keploy/go-sdk/integrations/khttpclient"
	"github.com/keploy/go-sdk/keploy"
	"github.com/keploy/go-sdk/mock"
)

var (
	server *httptest.Server
	ext    external.External
)

func fatal(t *testing.T, want, got interface{}) {
	t.Helper()
	t.Fatalf(`want: %v, got: %v`, want, got)
}

func mockFetchDataEndpoint(w http.ResponseWriter, r *http.Request) {
	ids, ok := r.URL.Query()["id"]

	sc := http.StatusOK
	m := make(map[string]interface{})

	if !ok || len(ids[0]) == 0 {
		sc = http.StatusBadRequest
	} else {
		m["id"] = "mock"
		m["name"] = "mock"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(sc)
	json.NewEncoder(w).Encode(m)
}

func TestMain(m *testing.M) {
	fmt.Println("mocking server")
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/":
			mockFetchDataEndpoint(w, r)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))

	fmt.Println("mocking external")
	// wrap the http client with the Keploy SDK
	interceptor := khttpclient.NewInterceptor(http.DefaultTransport)
	client := &http.Client{
		Transport: interceptor,
	}

	ext = external.New(server.URL, client, time.Second)

	fmt.Println("run tests")
	m.Run()
}

func TestExternal_FetchData(t *testing.T) {
	tt := []struct {
		name     string
		id       string
		wantData *external.Data
		wantErr  error
	}{
		{
			name:     "response not ok",
			id:       "",
			wantData: nil,
			wantErr:  external.ErrResponseNotOK,
		},
		{
			name: "data found",
			id:   "mock",
			wantData: &external.Data{
				ID:   "mock",
				Name: "mock",
			},
			wantErr: nil,
		},
	}

	ctx := mock.NewContext(mock.Config{
		Name: "hello",          // It is unique for every mock/stub. If you dont provide during record it would be generated. Its compulsory during tests.
		Mode: keploy.MODE_TEST, // It can be MODE_TEST or MODE_OFF. Default is MODE_TEST
	})

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotData, gotErr := ext.FetchData(ctx, tc.id)

			if !errors.Is(gotErr, tc.wantErr) {
				fatal(t, tc.wantErr, gotErr)
			}

			if !reflect.DeepEqual(gotData, tc.wantData) {
				fatal(t, tc.wantData, gotData)
			}
		})
	}
}
