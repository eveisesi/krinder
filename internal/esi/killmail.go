package esi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

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

// HTTP Get /v1/killmails/{id}/{hash}/
func (s *Service) KillmailByIDHash(ctx context.Context, id int64, hash string) (*KillmailOk, error) {

	var killmailOk = new(KillmailOk)

	path := fmt.Sprintf("/v1/killmails/%d/%s/", id, hash)

	err := s.request(ctx, http.MethodGet, path, nil, http.StatusOK, killmailOk)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute /v2/search on ESI API")
	}

	return killmailOk, nil

}
