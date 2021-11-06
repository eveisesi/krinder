package krinder

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type WarRepository interface {
	War(ctx context.Context, warID uint) (*MongoWar, error)
	Wars(ctx context.Context, operators ...*Operator) ([]*MongoWar, error)
	CreateWar(ctx context.Context, war *MongoWar) (*MongoWar, error)
	CreateWarBulk(ctx context.Context, wars []*MongoWar) error
	UpdateWar(ctx context.Context, war *MongoWar) error
}

type ESIWar struct {
	ID        int              `json:"id"`
	Aggressor *ESIWarAggressor `json:"aggressor"`
	Allies    []*ESIWarAlly    `json:"allies"`
	Defender  *ESIWarDefender  `json:"defender"`

	Mutual        bool `json:"mutual"`
	OpenForAllies bool `json:"open_for_allies"`

	Declared      time.Time `json:"declared"`
	Started       time.Time `json:"started"`
	Retracted     null.Time `json:"retracted"`
	Finished      null.Time `json:"finished"`
	ExpiresAt     null.Time `json:"expiresAt,omitempty"`
	IntegrityHash string    `json:"integrityHash,omitempty"`
}

func (u *ESIWar) ToMongoWar() *MongoWar {
	return &MongoWar{
		ID:        uint(u.ID),
		Aggressor: u.Aggressor.ToMongoWarAggressor(),
		Defender:  u.Defender.ToMongoWarDefender(),
		Allies: func(a []*ESIWarAlly) []*MongoWarAlly {
			var b = make([]*MongoWarAlly, 0, len(a))
			for _, ally := range u.Allies {
				b = append(b, ally.ToMongoWarAlly())
			}
			return b
		}(u.Allies),
		Mutual:        u.Mutual,
		OpenForAllies: u.OpenForAllies,
		Declared:      u.Declared,
		Started:       u.Started,
		Retracted:     u.Retracted.Ptr(),
		Finished:      u.Finished.Ptr(),
		ExpiresAt:     u.ExpiresAt.Ptr(),
		IntegrityHash: u.IntegrityHash,
	}
}

type ESIWarAggressor struct {
	AllianceID    null.Uint `json:"alliance_id"`
	CorporationID null.Uint `json:"corporation_id"`
	IskDestroyed  float64   `json:"isk_destroyed"`
	ShipsKilled   uint      `json:"ships_killed"`
}

func (i *ESIWarAggressor) ToMongoWarAggressor() *MongoWarAggressor {
	return &MongoWarAggressor{
		AllianceID:    i.AllianceID.Ptr(),
		CorporationID: i.CorporationID.Ptr(),
		IskDestroyed:  i.IskDestroyed,
		ShipsKilled:   i.ShipsKilled,
	}

}

type ESIWarDefender struct {
	AllianceID    null.Uint `json:"alliance_id"`
	CorporationID null.Uint `json:"corporation_id"`
	IskDestroyed  float64   `json:"isk_destroyed"`
	ShipsKilled   uint      `json:"ships_killed"`
}

func (i *ESIWarDefender) ToMongoWarDefender() *MongoWarDefender {
	return &MongoWarDefender{
		AllianceID:    i.AllianceID.Ptr(),
		CorporationID: i.CorporationID.Ptr(),
		IskDestroyed:  i.IskDestroyed,
		ShipsKilled:   i.ShipsKilled,
	}
}

type ESIWarAlly struct {
	AllianceID    null.Uint `bson:"allianceID"`
	CorporationID null.Uint `bson:"corporationID"`
}

func (i *ESIWarAlly) ToMongoWarAlly() *MongoWarAlly {
	return &MongoWarAlly{
		AllianceID:    i.AllianceID.Ptr(),
		CorporationID: i.CorporationID.Ptr(),
	}
}

type MongoWar struct {
	// ID of the specified war
	ID uint `bson:"id"`
	// allied corporations or alliances, each object contains either corporation_id or alliance_id
	Allies []*MongoWarAlly `bson:"allies"`
	// Time that the war was declared
	Declared time.Time `bson:"declared"`
	// Time the war ended and shooting was no longer allowed
	Finished *time.Time `bson:"finished,omitempty"`
	// Was the war declared mutual by both parties
	Mutual bool `bson:"mutual"`
	// Is the war currently open for allies or not
	OpenForAllies bool `bson:"openForAllies"`
	// Time the war was retracted but both sides could still shoot each other
	Retracted *time.Time `bson:"retracted,omitempty"`
	// Time when the war started and both sides could shoot each other
	Started time.Time `bson:"started"`

	Aggressor *MongoWarAggressor `bson:"aggressor,omitempty"`
	Defender  *MongoWarDefender  `bson:"defender,omitempty"`

	// DateTime the record was inserted into the DB
	CreatedAt time.Time `bson:"createdAt"`
	// DateTime the record in the database was last updated
	UpdatedAt time.Time `bson:"updatedAt"`
	// DateTime the record is considered stale and should receive an update.
	// This can be null is the record will never be considered stale
	ExpiresAt *time.Time `bson:"expiresAt,omitempty"`
	// This is an ETag from the API that sourced this information.
	// Please do not ask anybody how to calculate this
	IntegrityHash string `bson:"integrityHash,omitempty"`
}

type MongoWarAggressor struct {
	// Alliance ID if and only if the aggressor is an alliance
	AllianceID *uint `bson:"allianceID,omitempty"`
	// Corporation ID if and only if the aggressor is a corporation
	CorporationID *uint `bson:"corporationID,omitempty"`
	// ISK value of ships the aggressor has destroyed
	IskDestroyed float64 `bson:"iskDestroyed"`
	// The number of ships the aggressor has killed
	ShipsKilled uint `bson:"shipsKilled"`
}

type MongoWarDefender struct {
	// Alliance ID if and only if the defender is an alliance
	AllianceID *uint `bson:"allianceID,omitempty"`
	// Corporation ID if and only if the defender is a corporation
	CorporationID *uint `bson:"corporationID,omitempty"`
	// ISK value of ships the aggressor has destroyed
	IskDestroyed float64 `bson:"iskDestroyed"`
	// The number of ships the aggressor has killed
	ShipsKilled uint `bson:"shipsKilled"`
}

type MongoWarAlly struct {
	// Alliance ID if and only if this ally is an alliance
	AllianceID *uint `bson:"allianceID,omitempty"`
	// Corporation ID if and only if this ally is a corporation
	CorporationID *uint `bson:"corporationID,omitempty"`
}
