package esi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/eveisesi/krinder"
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

type GroupOk struct {
	Expires time.Time
	Etag    string
	Group   *krinder.ESIGroup
}

func (s *Service) Group(ctx context.Context, id uint, reqFuncs ...RequestFunc) (*GroupOk, error) {

	var group = new(krinder.ESIGroup)
	var out = &Out{Data: group}

	path := fmt.Sprintf("/v1/universe/groups/%d/", id)
	err := s.request(ctx, http.MethodGet, path, nil, http.StatusOK, time.Duration(time.Hour*24), out, reqFuncs, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch group")
	}

	var etag = out.Headers.Get("etag")
	var expires time.Time
	if xexpires := out.Headers.Get("expires"); xexpires != "" {
		parsed, err := time.Parse(HeaderTimestampFormat, xexpires)
		if err == nil {
			expires = parsed
		}

	}

	return &GroupOk{
		Expires: expires,
		Etag:    etag,
		Group:   group,
	}, nil

}

type GroupsOk struct {
	Pages uint
	IDs   []uint
}

func (s *Service) Groups(ctx context.Context, page uint) (*GroupsOk, error) {

	var ids = make([]uint, 0)
	var out = &Out{Data: &ids}

	if page == 0 {
		page = 1
	}

	path := fmt.Sprintf("/v1/universe/groups/?page=%d", page)
	err := s.request(ctx, http.MethodGet, path, nil, http.StatusOK, time.Duration(time.Hour*24), out, nil, nil)

	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch groups")
	}

	var pages uint64 = 1
	if xpages := out.Headers.Get("x-pages"); xpages != "" {
		ppages, err := strconv.ParseUint(xpages, 10, 32)
		if err == nil {
			pages = ppages
		}
	}

	return &GroupsOk{Pages: uint(pages), IDs: ids}, nil

}
