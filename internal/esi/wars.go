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
	var out = &Out{Data: &warIDs}

	err := s.request(ctx, http.MethodGet, "/v1/wars/", nil, http.StatusOK, time.Duration(time.Hour), out, nil, nil)

	return warIDs, errors.Wrap(err, "failed to fetch wars")

}

func (s *Service) War(ctx context.Context, id uint, etag string) (*krinder.ESIWar, error) {

	var war = new(krinder.ESIWar)
	var out = &Out{Data: war}

	reqFuncs := make([]requestFunc, 0)
	if etag != "" {
		reqFuncs = append(reqFuncs, WarsAddIfNoneMatchHeader(etag))
	}

	path := fmt.Sprintf("/v1/wars/%d/", id)
	err := s.request(
		ctx,
		http.MethodGet,
		path,
		nil,
		http.StatusOK,
		time.Duration(time.Hour),
		out,
		reqFuncs,
		[]responseFunc{WarAddResponseHeaders()},
	)

	if out.Status == http.StatusNotModified {

		return nil, nil
	}

	fmt.Println(out.Status)

	return war, err

}

func WarsAddIfNoneMatchHeader(etag string) requestFunc {
	return func(req *http.Request) {
		if etag == "" {
			return
		}
		req.Header.Set("If-None-Match", etag)
	}

}

func WarAddResponseHeaders() responseFunc {
	return func(out *Out) {

		war, ok := out.Data.(*krinder.ESIWar)
		if !ok {
			return
		}

		if expiresStr := out.Headers.Get("Expires"); expiresStr != "" {
			expires, err := time.Parse(HeaderTimestampFormat, expiresStr)
			if err == nil {
				war.ExpiresAt.SetValid(expires.Add(time.Hour * 12))
			}
		}

		if etag := out.Headers.Get("Etag"); etag != "" {
			war.IntegrityHash = etag
		}

		out.Data = war

	}
}
