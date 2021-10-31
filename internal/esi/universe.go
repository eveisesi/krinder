package esi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	var out = &Out{Data: &names}
	err = s.request(ctx, http.MethodPost, "/v3/universe/names", bytes.NewReader(data), http.StatusOK, time.Duration(0), out, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute /v2/search on ESI API")
	}

	return names, nil

}

type SystemOk struct {
	ID              uint    `json:"system_id"`
	Name            string  `json:"name"`
	ConstellationID uint    `json:"constellation_id"`
	SecurityStatus  float64 `json:"security_status"`
}

func (s *Service) System(ctx context.Context, id uint) (*SystemOk, error) {

	var systemOk = new(SystemOk)
	var out = &Out{Data: systemOk}

	path := fmt.Sprintf("/v4/universe/systems/%d/", id)
	err := s.request(ctx, http.MethodGet, path, nil, http.StatusOK, time.Duration(time.Hour*24), out, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch system")
	}

	return systemOk, nil

}
