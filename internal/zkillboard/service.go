package zkillboard

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/eveisesi/krinder/pkg/roundtripper"
	"github.com/pkg/errors"
)

type API interface{}

type Service struct {
	url    string
	client *http.Client
}

func New(userAgent string) *Service {
	return &Service{
		url: "https://zkillboard.com/api",
		client: &http.Client{
			Transport: roundtripper.UserAgent(userAgent, http.DefaultTransport),
		},
	}
}

func (s *Service) request(ctx context.Context, method, path string, body io.Reader, expected int, out interface{}) error {

	url := fmt.Sprintf("%s/%s", s.url, path)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	res, err := s.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}

	defer func(requestID string, body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			fmt.Printf("failed to close requst body for %s\n", requestID)
		}
	}(fmt.Sprintf("%s %s", method, path), res.Body)

	if res.StatusCode > 299 || res.StatusCode != expected {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.Wrapf(err, "expected status %d, got %d: unable to parse request body", expected, res.StatusCode)
		}

		return errors.Wrapf(err, "expected status %d, got %s: %s", expected, res.StatusCode, string(data))
	}

	err = json.NewDecoder(res.Body).Decode(out)
	if err != nil {
		return errors.Wrap(err, "failed to decode request body to json")
	}

	return nil

}
