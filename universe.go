package krinder

import (
	"context"
	"time"
)

type UniverseRepository interface {
	groupRespository
	entityRepository
}

type groupRespository interface {
	Group(ctx context.Context, groupID uint) (*MySQLGroup, error)
	Groups(ctx context.Context, operators ...*Operator) ([]*MySQLGroup, error)
	CreateGroup(ctx context.Context, group *MySQLGroup) (*MySQLGroup, error)
	UpdateGroup(ctx context.Context, group *MySQLGroup) (*MySQLGroup, error)
}

type ESIGroup struct {
	CategoryID uint   `json:"category_id"`
	GroupID    uint   `json:"group_id"`
	Name       string `json:"name"`
	Published  bool   `json:"published"`
}

func (e *ESIGroup) ToMongoGroup() *MySQLGroup {
	return &MySQLGroup{
		CategoryID: e.CategoryID,
		GroupID:    e.GroupID,
		Name:       e.Name,
		Published:  e.Published,
	}
}

type MySQLGroup struct {
	CategoryID uint      `db:"category_id"`
	GroupID    uint      `db:"group_id"`
	Name       string    `db:"name"`
	Published  bool      `db:"published"`
	Expires    time.Time `db:"expires"`
	Etag       string    `db:"etag"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type entityRepository interface {
	Entity(ctx context.Context, typeID uint) (*MongoEntity, error)
	Entitys(ctx context.Context, operators ...*Operator) ([]*MongoEntity, error)
	CreateEntity(ctx context.Context, t *MongoEntity) (*MongoEntity, error)
	UpdateEntity(ctx context.Context, t *MongoEntity) (*MongoEntity, error)
}

type ESIEntity struct {
	ID        uint   `json:"type_id"`
	Name      string `json:"name"`
	Published bool   `json:"published"`
}

func (e *ESIEntity) ToMongoEntity() *MongoEntity {
	return &MongoEntity{
		ID:        e.ID,
		Name:      e.Name,
		Published: e.Published,
	}
}

type MongoEntity struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Published bool      `json:"published"`
	Expires   time.Time `db:"expires"`
	Etag      string    `db:"etag"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
