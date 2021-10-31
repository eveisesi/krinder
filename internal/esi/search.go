package esi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

type SearchOk struct {
	Character []int `json:"character,omitempty"`
}

// HTTP Get /v2/search/?categories={category}&term={term}
func (s *Service) Search(ctx context.Context, category, term string, strict bool) (*SearchOk, error) {

	var searchOK = new(SearchOk)
	var out = &Out{Data: searchOK}
	v := url.Values{}
	v.Add("categories", category)
	v.Add("search", term)
	if strict {
		v.Add("strict", "true")
	}

	path := fmt.Sprintf("/v2/search/?%s", v.Encode())

	err := s.request(ctx, http.MethodGet, path, nil, http.StatusOK, time.Hour, out, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute /v2/search on ESI API")
	}

	return searchOK, nil

}
