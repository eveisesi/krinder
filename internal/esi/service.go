package esi

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eveisesi/krinder/pkg/roundtripper"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type API interface {
	// Characters
	Character(ctx context.Context, id uint) (*CharacterOk, error)
	// Killmails
	KillmailByIDHash(ctx context.Context, id int64, hash string) (*KillmailOk, error)
	// Search
	Search(ctx context.Context, category, term string) (*SearchOk, error)
	// Universe
	Names(ctx context.Context, ids []int) ([]*NamesOk, error)
	System(ctx context.Context, id uint) (*SystemOk, error)
}

type Service struct {
	url    string
	client *http.Client
	cache  *redis.Client
}

const (
	HeaderTimestampFormat = "Mon, 02 Jan 2006 15:04:05 MST"
)

func New(userAgent string, cache *redis.Client) *Service {
	return &Service{
		url: "https://esi.evetech.net",
		client: &http.Client{
			Transport: roundtripper.UserAgent(userAgent, http.DefaultTransport),
		},
		cache: cache,
	}
}

// Execute a request to the ESI API using the provided Method, Path, and Body. If the response status != the exepected status
// the response body is decoded to a slice of bytes, converted to a string, and appended to the end of an error message

type Out struct {
	Data    interface{} `json:"data"`
	Headers http.Header `json:"headers"`
	Status  int         `json:"status"`
}

func (s *Service) request(ctx context.Context, method, path string, body io.Reader, expected int, cacheDuration time.Duration, out *Out, reqMods []requestFunc, respMods []responseFunc) error {

	url := fmt.Sprintf("%s%s", s.url, path)

	if cacheDuration != 0 {
		if err := s.getResponseCache(ctx, url, out); err == nil {
			return nil
		}
	}
	var res = new(http.Response)
	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return errors.Wrap(err, "failed to create request")
		}

		if len(reqMods) > 0 {
			for _, mod := range reqMods {
				mod(req)
			}
		}

		res, err = s.client.Do(req)
		if err != nil {
			return errors.Wrap(err, "failed to execute request")
		}

		if res.StatusCode == http.StatusTooManyRequests {
			fmt.Println("received 429, sleeping for 10 seconds")
			time.Sleep(time.Second * 10)
			continue
		}

		if res.StatusCode < http.StatusInternalServerError {
			break
		}

		time.Sleep(time.Second)
	}

	defer func(requestID string, body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			fmt.Printf("failed to close requst body for %s\n", requestID)
		}
	}(fmt.Sprintf("%s %s", method, path), res.Body)

	if res.StatusCode > 399 || (res.StatusCode != expected && res.StatusCode != http.StatusNotModified) {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.Wrapf(err, "expected status %d, got %d: unable to parse request body", expected, res.StatusCode)
		}

		return errors.Errorf("expected status %d, got %d: %s", expected, res.StatusCode, string(data))
	}

	out.Status = res.StatusCode
	out.Headers = res.Header

	if out.Status == http.StatusOK {
		if cacheDuration == time.Duration(-1) {
			expiresHeader := res.Header.Get("Expires")
			if expiresHeader != "" {
				expires, err := time.Parse(HeaderTimestampFormat, expiresHeader)
				if err == nil {
					cacheDuration = time.Until(expires)
				}
			}
		}

		err := json.NewDecoder(res.Body).Decode(out.Data)
		if err != nil {
			return errors.Wrap(err, "failed to decode request body to json")
		}
	}

	if len(respMods) > 0 {
		for _, mod := range respMods {
			mod(out)
		}
	}

	if method == http.MethodGet {
		_ = s.setResponseCache(ctx, url, cacheDuration, out)

	}

	return nil

}

// if duration == -2, results will be cached permenantly with no duration,
// if duration == -1, results will be cached temporarily according to the Expires header on the result if the result response code == expected
// if duration == 0, results will not be cached
// if duration > 0, results will be cached accordding to the provided value multiple by time.Second
func (s *Service) setResponseCache(ctx context.Context, url string, duration time.Duration, out *Out) error {

	var d time.Duration
	if duration == -2 {
		d = time.Duration(0)
	} else if duration == -1 {
		return errors.New("invalid duration")
	} else if duration == 0 {
		return nil
	}

	payload, err := json.Marshal(out)
	if err != nil {
		return errors.Wrap(err, "failed to encode payload to json")
	}

	key := fmt.Sprintf("%x", sha1.Sum([]byte(url)))
	_, err = s.cache.Set(ctx, key, string(payload), d).Result()

	return errors.Wrap(err, "failed to cache response")

}

func (s *Service) getResponseCache(ctx context.Context, url string, out *Out) error {

	key := fmt.Sprintf("%x", sha1.Sum([]byte(url)))
	b, err := s.cache.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, out)

	return err

}
