package httpcli

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var HttpClient = &http.Client{
	Timeout: time.Second * 10,
}

func Get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext((ctx), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to construct http.Request: %w", err)
	}

	resp, err := HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to issue request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read body of %s: %w", url, err)
	}

	return bs, nil
}
