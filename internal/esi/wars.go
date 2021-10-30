package esi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/eveisesi/krinder"
	"github.com/pkg/errors"
)

func (s *Service) Wars(ctx context.Context) ([]int, error) {

	var warIDs = make([]int, 0)

	err := s.request(ctx, http.MethodGet, "/v1/wars/", nil, http.StatusOK, time.Duration(time.Hour), &warIDs)

	return warIDs, errors.Wrap(err, "failed to fetch wars")

}

func (s *Service) War(ctx context.Context, id uint) (*krinder.ESIWar, error) {

	var war = new(krinder.ESIWar)

	path := fmt.Sprintf("/v1/wars/%d/", id)
	err := s.request(ctx, http.MethodGet, path, nil, http.StatusOK, time.Duration(time.Hour), war)

	return war, err

}
