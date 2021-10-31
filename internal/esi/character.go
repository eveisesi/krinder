package esi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type CharacterOk struct {
	ID             uint64  `json:"id"`
	Name           string  `json:"name"`
	SecurityStatus float64 `json:"security_status"`
	CorporationID  uint    `json:"corporation_id"`
	AllianceID     uint    `json:"alliance_id,omitempty"`
}

func (s *Service) Character(ctx context.Context, id uint64) (*CharacterOk, error) {

	var characterOk = new(CharacterOk)
	var out = &Out{Data: characterOk}
	path := fmt.Sprintf("/v5/characters/%d/", id)
	err := s.request(ctx, http.MethodGet, path, nil, http.StatusOK, time.Duration(-1), out, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch character")
	}

	characterOk.ID = id

	return characterOk, nil

}
