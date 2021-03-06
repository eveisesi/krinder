package esi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type KillmailOk struct {
	Attackers     []*KillmailAttacker `json:"attackers"`
	KillmailID    int                 `json:"killmail_id"`
	KillmailTime  time.Time           `json:"killmail_time"`
	SolarSystemID int                 `json:"solar_system_id"`
	Victim        *KillmailVictim     `json:"victim"`
	SolarSystem   *SystemOk           `json:"system,omitempty"`
}

type KillmailAttacker struct {
	CharacterID    uint64  `json:"character_id"`
	CorporationID  uint    `json:"corporation_id"`
	AllianceID     uint    `json:"alliance_id"`
	DamageDone     uint    `json:"damage_done"`
	FinalBlow      bool    `json:"final_blow"`
	SecurityStatus float64 `json:"security_status"`
	ShipTypeID     uint    `json:"ship_type_id"`
	WeaponTypeID   uint    `json:"weapon_type_id"`
	FactionID      uint    `json:"faction_id,omitempty"`
}

type KillmailVictimItem struct {
	Flag              uint `json:"flag"`
	ItemTypeID        uint `json:"item_type_id"`
	QuantityDestroyed uint `json:"quantity_destroyed,omitempty"`
	Singleton         uint `json:"singleton"`
	QuantityDropped   uint `json:"quantity_dropped,omitempty"`
}

type KillmailVictim struct {
	AllianceID    uint                  `json:"alliance_id"`
	CharacterID   uint64                `json:"character_id"`
	CorporationID uint                  `json:"corporation_id"`
	DamageTaken   uint                  `json:"damage_taken"`
	Items         []*KillmailVictimItem `json:"items"`
	ShipTypeID    uint                  `json:"ship_type_id"`

	Character *CharacterOk `json:"character,omitempty"`
}

// HTTP Get /v1/killmails/{id}/{hash}/
func (s *service) KillmailByIDHash(ctx context.Context, id int64, hash string) (*KillmailOk, error) {

	var killmailOk = new(KillmailOk)
	var out = &Out{Data: killmailOk}
	path := fmt.Sprintf("/v1/killmails/%d/%s/", id, hash)

	err := s.request(ctx, http.MethodGet, path, nil, http.StatusOK, time.Duration(-2), out, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute /v2/search on ESI API")
	}

	return killmailOk, nil

}
