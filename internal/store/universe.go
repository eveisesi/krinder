package store

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/davecgh/go-spew/spew"
	"github.com/eveisesi/krinder"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UniverseRepository struct {
	mysql struct {
		db *sqlx.DB
	}
	mongo struct {
		entities *mongo.Collection
	}
}

var _ krinder.UniverseRepository = new(UniverseRepository)

var (
	groupsTableColumns = []string{GroupCategoryID,
		GroupGroupID,
		GroupName,
		GroupEtag,
		GroupPublished,
		GroupExpires,
		GroupCreatedAt,
		GroupUpdatedAt,
	}
	groupsTable      = "groups"
	entityCollection = "entities"
)

func NewUniverseRepository(sqldb *sqlx.DB, mongodb *mongo.Database) (*UniverseRepository, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	entities := mongodb.Collection(entityCollection)

	_, err := entities.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: primitive.D{
				primitive.E{Key: EntityID, Value: 1},
			},
			Options: &options.IndexOptions{
				Sparse: null.BoolFrom(true).Ptr(),
				Unique: null.BoolFrom(true).Ptr(),
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create index")
	}

	repo := new(UniverseRepository)
	repo.mongo.entities = entities
	repo.mysql.db = sqldb

	return repo, nil

}

func (r *UniverseRepository) Group(ctx context.Context, groupID uint) (*krinder.MySQLGroup, error) {

	query, args, err := sq.Select(groupsTableColumns...).From(groupsTable).
		Where(sq.Eq{GroupGroupID: groupID}).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate query")
	}

	var group = new(krinder.MySQLGroup)

	err = r.mysql.db.GetContext(ctx, group, query, args...)

	return group, err

}

func (r *UniverseRepository) Groups(ctx context.Context, operators ...*krinder.Operator) ([]*krinder.MySQLGroup, error) {

	query, args, err := BuildSQLFilters(
		sq.Select(groupsTableColumns...).From(groupsTable),
		operators...).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate query")
	}

	var groups = make([]*krinder.MySQLGroup, 0)

	err = r.mysql.db.SelectContext(ctx, &groups, query, args...)

	return groups, err

}

func (r *UniverseRepository) CreateGroup(ctx context.Context, group *krinder.MySQLGroup) (*krinder.MySQLGroup, error) {

	group.CreatedAt = time.Now().UTC()
	group.UpdatedAt = time.Now().UTC()

	query, args, err := sq.Insert(groupsTable).SetMap(map[string]interface {
	}{
		GroupCategoryID: group.CategoryID,
		GroupGroupID:    group.GroupID,
		GroupName:       group.Name,
		GroupPublished:  group.Published,
		GroupEtag:       group.Etag,
		GroupExpires:    group.Expires,
		GroupCreatedAt:  group.CreatedAt,
		GroupUpdatedAt:  group.UpdatedAt,
	}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate query")
	}

	_, err = r.mysql.db.ExecContext(ctx, query, args...)

	return group, err

}

func (r *UniverseRepository) UpdateGroup(ctx context.Context, group *krinder.MySQLGroup) (*krinder.MySQLGroup, error) {

	group.UpdatedAt = time.Now()

	query, args, err := sq.Update(groupsTable).SetMap(map[string]interface {
	}{
		GroupCategoryID: group.CategoryID,
		GroupName:       group.Name,
		GroupPublished:  group.Published,
		GroupEtag:       group.Etag,
		GroupExpires:    group.Expires,
		GroupUpdatedAt:  group.UpdatedAt,
	}).Where(sq.Eq{GroupGroupID: group.GroupID}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate query")
	}

	spew.Dump(query)

	_, err = r.mysql.db.ExecContext(ctx, query, args...)
	return group, err

}

func (r *UniverseRepository) Entity(ctx context.Context, entityID uint) (*krinder.MongoEntity, error) {

	var entity = new(krinder.MongoEntity)

	err := r.mongo.entities.FindOne(ctx, primitive.D{primitive.E{Key: EntityID, Value: entityID}}).Decode(entity)

	return entity, err

}
func (r *UniverseRepository) Entitys(ctx context.Context, operators ...*krinder.Operator) ([]*krinder.MongoEntity, error) {

	filters := BuildMongoFilters(operators...)
	options := BuildMongoFindOptions(operators...)

	var entities = make([]*krinder.MongoEntity, 0)
	result, err := r.mongo.entities.Find(ctx, filters, options)
	if err != nil {
		return entities, err
	}

	return entities, result.All(ctx, &entities)

}
func (r *UniverseRepository) CreateEntity(ctx context.Context, entity *krinder.MongoEntity) (*krinder.MongoEntity, error) {

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()

	_, err := r.mongo.entities.InsertOne(ctx, entity)
	if err != nil {
		if !mongo.IsDuplicateKeyError(err) {
			return nil, err
		}

	}

	return entity, nil

}

func (r *UniverseRepository) UpdateEntity(ctx context.Context, entity *krinder.MongoEntity) (*krinder.MongoEntity, error) {

	entity.UpdatedAt = time.Now()

	filter := BuildMongoFilters(krinder.NewEqualOperator(EntityID, entity.ID))
	_, err := r.mongo.entities.UpdateOne(ctx, filter, primitive.D{primitive.E{Key: "$set", Value: entity}})

	return entity, err

}
