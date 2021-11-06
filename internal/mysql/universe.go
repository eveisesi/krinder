package mysql

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/davecgh/go-spew/spew"
	"github.com/eveisesi/krinder"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type UniverseRepository struct {
	db *sqlx.DB
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
	groupsTable = "groups"
)

func NewUniverseRepository(db *sqlx.DB) *UniverseRepository {
	return &UniverseRepository{
		db,
	}
}

func (r *UniverseRepository) Group(ctx context.Context, groupID uint) (*krinder.MySQLGroup, error) {

	query, args, err := sq.Select(groupsTableColumns...).From(groupsTable).
		Where(sq.Eq{GroupGroupID: groupID}).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate query")
	}

	var group = new(krinder.MySQLGroup)

	err = r.db.GetContext(ctx, group, query, args...)

	return group, err

}

func (r *UniverseRepository) Groups(ctx context.Context, operators ...*krinder.Operator) ([]*krinder.MySQLGroup, error) {

	query, args, err := BuildFilters(
		sq.Select(groupsTableColumns...).From(groupsTable),
		operators...).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate query")
	}

	var groups = make([]*krinder.MySQLGroup, 0)

	err = r.db.SelectContext(ctx, &groups, query, args...)

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

	_, err = r.db.ExecContext(ctx, query, args...)

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

	_, err = r.db.ExecContext(ctx, query, args...)
	return group, err

}
