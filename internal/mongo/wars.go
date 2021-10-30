package mdb

import (
	"context"
	"time"

	"github.com/eveisesi/krinder"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WarAPI interface {
	War(ctx context.Context, warID uint) (*krinder.MongoWar, error)
	Wars(ctx context.Context, operators ...*krinder.Operator) ([]*krinder.MongoWar, error)
	CreateWar(ctx context.Context, war *krinder.MongoWar) (*krinder.MongoWar, error)
	CreateWarBulk(ctx context.Context, wars []*krinder.MongoWar) error
}

type WarRepository struct {
	wars *mongo.Collection
}

func NewWarRepository(database *mongo.Database) (*WarRepository, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	wars := database.Collection("wars")

	_, err := wars.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				primitive.E{Key: "id", Value: 1},
			},
			Options: &options.IndexOptions{
				Name:   null.StringFrom("wars_id_unique").Ptr(),
				Unique: null.BoolFrom(true).Ptr(),
			},
		},
		{
			Keys: bson.D{
				primitive.E{Key: "aggressor.allianceID", Value: 1},
			},
			Options: &options.IndexOptions{
				Name:   null.StringFrom("wars_aggressor_allianceID_index").Ptr(),
				Sparse: null.BoolFrom(true).Ptr(),
			},
		},
		{
			Keys: bson.D{
				primitive.E{Key: "aggressor.corporation_ID", Value: 1},
			},
			Options: &options.IndexOptions{
				Name:   null.StringFrom("wars_aggressor_corporation_ID_index").Ptr(),
				Sparse: null.BoolFrom(true).Ptr(),
			},
		},
		{
			Keys: bson.D{
				primitive.E{Key: "defender.allianceID", Value: 1},
			},
			Options: &options.IndexOptions{
				Name:   null.StringFrom("wars_defender_allianceID_index").Ptr(),
				Sparse: null.BoolFrom(true).Ptr(),
			},
		},
		{
			Keys: bson.D{
				primitive.E{Key: "defender.corporationID", Value: 1},
			},
			Options: &options.IndexOptions{
				Name:   null.StringFrom("wars_defender_corporationID_index").Ptr(),
				Sparse: null.BoolFrom(true).Ptr(),
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create index")
	}

	return &WarRepository{
		wars: wars,
	}, nil

}

func (r *WarRepository) War(ctx context.Context, warID uint) (*krinder.MongoWar, error) {

	var war = new(krinder.MongoWar)

	err := r.wars.FindOne(ctx, bson.D{primitive.E{Key: "id", Value: warID}}).Decode(war)

	return war, err

}

func (r *WarRepository) Wars(ctx context.Context, operators ...*krinder.Operator) ([]*krinder.MongoWar, error) {

	filters := BuildFilters(operators...)
	options := BuildFindOptions(operators...)

	var wars = make([]*krinder.MongoWar, 0)
	result, err := r.wars.Find(ctx, filters, options)
	if err != nil {
		return wars, err
	}

	return wars, result.All(ctx, &wars)

}

func (r *WarRepository) CreateWar(ctx context.Context, war *krinder.MongoWar) (*krinder.MongoWar, error) {
	war.CreatedAt = time.Now().UTC()
	war.UpdatedAt = time.Now().UTC()

	_, err := r.wars.InsertOne(ctx, war)
	if err != nil {
		if !isUniqueConstrainViolation(err) {
			return nil, err
		}

	}

	return war, nil

}

func (r *WarRepository) CreateWarBulk(ctx context.Context, wars []*krinder.MongoWar) error {

	now := time.Now().UTC()
	documents := make([]interface{}, len(wars))
	for i, war := range wars {
		war.CreatedAt = now
		war.UpdatedAt = now

		documents[i] = war
	}

	results, err := r.wars.InsertMany(ctx, documents)
	if err != nil {
		if !isUniqueConstrainViolation(err) {
			return errors.Wrap(err, "failed to insert wars in bulk")
		}
	}

	if len(results.InsertedIDs) != len(wars) {
		return errors.Errorf("length of inserted ids (%d) does not match length of provided documents (%d)", len(results.InsertedIDs), len(documents))
	}

	return nil

}
