package esi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type NamesOk struct {
	Category string `json:"category"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
}

// HTTP Post /v3/universe/names
func (s *Service) Names(ctx context.Context, ids []int) ([]*NamesOk, error) {

	data, err := json.Marshal(ids)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode slice of ids to json")
	}

	var names = make([]*NamesOk, 0, len(ids))
	err = s.request(ctx, http.MethodPost, "/v3/universe/names", bytes.NewReader(data), http.StatusOK, &names)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute /v2/search on ESI API")
	}

	return names, nil

}
