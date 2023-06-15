// external.go
package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	ErrResponseNotOK error = errors.New("response not ok")
)

type (
	Data struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	External interface {
		FetchData(ctx context.Context, id string) (*Data, error)
	}

	v1 struct {
		baseURL string
		client  *http.Client
		timeout time.Duration
	}
)

func New(baseURL string, client *http.Client, timeout time.Duration) *v1 {
	return &v1{
		baseURL: baseURL,
		client:  client,
		timeout: timeout,
	}
}

func (v *v1) FetchData(ctx context.Context, id string) (*Data, error) {
	url := fmt.Sprintf("%s/?id=%s", v.baseURL, id)

	ctx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w. %s", ErrResponseNotOK, http.StatusText(resp.StatusCode))
	}

	var d *Data
	return d, json.NewDecoder(resp.Body).Decode(&d)
}
