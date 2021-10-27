package esi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

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
		url: "https://esi.evetech.net",
		client: &http.Client{
			Transport: roundtripper.UserAgent(userAgent, http.DefaultTransport),
		},
	}

}

func (s *Service) request(ctx context.Context, method, path string, body io.Reader, expected int, out interface{}) error {

	url := fmt.Sprintf("%s%s", s.url, path)
	fmt.Println(url)
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

		return errors.Errorf("expected status %d, got %d: %s", expected, res.StatusCode, string(data))
	}

	err = json.NewDecoder(res.Body).Decode(out)
	if err != nil {
		return errors.Wrap(err, "failed to decode request body to json")
	}

	return nil

}

type SearchOk struct {
	Character []int `json:"character"`
}

func (s *Service) Search(ctx context.Context, category, term string) (*SearchOk, error) {

	var searchOK = new(SearchOk)

	v := url.Values{}
	v.Add("categories", category)
	v.Add("search", term)

	path := fmt.Sprintf("/v2/search/?%s", v.Encode())

	err := s.request(ctx, http.MethodGet, path, nil, http.StatusOK, searchOK)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute /v2/search on ESI API")
	}

	return searchOK, nil

}

type NamesOk struct {
	Category string `json:"category"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
}

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

type KillmailOk struct {
	Attackers     []KillmailAttacker `json:"attackers"`
	KillmailID    int                `json:"killmail_id"`
	KillmailTime  time.Time          `json:"killmail_time"`
	SolarSystemID int                `json:"solar_system_id"`
	Victim        *KillmailVictim    `json:"victim"`
}

type KillmailAttacker struct {
	CharacterID    int     `json:"character_id"`
	CorporationID  int     `json:"corporation_id"`
	DamageDone     int     `json:"damage_done"`
	FinalBlow      bool    `json:"final_blow"`
	SecurityStatus float64 `json:"security_status"`
	ShipTypeID     int     `json:"ship_type_id"`
	WeaponTypeID   int     `json:"weapon_type_id"`
	FactionID      int     `json:"faction_id,omitempty"`
}

type KillmailVictimItem struct {
	Flag              int `json:"flag"`
	ItemTypeID        int `json:"item_type_id"`
	QuantityDestroyed int `json:"quantity_destroyed,omitempty"`
	Singleton         int `json:"singleton"`
	QuantityDropped   int `json:"quantity_dropped,omitempty"`
}

type KillmailVictim struct {
	AllianceID    int                   `json:"alliance_id"`
	CharacterID   int                   `json:"character_id"`
	CorporationID int                   `json:"corporation_id"`
	DamageTaken   int                   `json:"damage_taken"`
	Items         []*KillmailVictimItem `json:"items"`
	ShipTypeID    int                   `json:"ship_type_id"`
}

func (s *Service) KillmailByIDHash(ctx context.Context, id int64, hash string) (*KillmailOk, error) {

	var killmailOk = new(KillmailOk)

	path := fmt.Sprintf("/v1/killmails/%d/%s/", id, hash)

	err := s.request(ctx, http.MethodGet, path, nil, http.StatusOK, killmailOk)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute /v2/search on ESI API")
	}

	return killmailOk, nil

}
