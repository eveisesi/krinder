package universe

import (
	"context"
	"database/sql"
	"time"

	"github.com/eveisesi/krinder"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type UniverseAPI interface {
	// Entity(ctx context.Context, entityID uint) (*krinder.MongoEntity, error)
	Groups(ctx context.Context, operators ...*krinder.Operator) ([]*krinder.MySQLGroup, error)
}

type Service struct {
	cache  *redis.Client
	logger *logrus.Logger

	esi      esi.API
	universe krinder.UniverseRepository
	UniverseAPI
}

var _ UniverseAPI = new(Service)

func New(logger *logrus.Logger, cache *redis.Client, esi esi.API, universe krinder.UniverseRepository) *Service {
	return &Service{
		logger:      logger,
		cache:       cache,
		esi:         esi,
		universe:    universe,
		UniverseAPI: universe,
	}
}

func (s *Service) Entity(ctx context.Context, entityID uint) (*krinder.MongoEntity, error) {

	entry := s.logger.WithFields(logrus.Fields{
		"entityID": entityID,
		"service":  "universe",
	})

	entity, err := s.universe.Entity(ctx, entityID)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.Wrap(err, "failed to fetch entity from datastore")
	}

	if entity.Expires.Unix() > time.Now().Unix() {
		return entity, err
	}

	var create = false
	if errors.Is(err, mongo.ErrNoDocuments) {
		create = true
	}

	esiEntity, err := s.esi.Type(ctx, entityID, esi.AddIfNoneMatchHeader(entity.Etag))
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch type from ESI")
	}

	if entity.Etag != esiEntity.Etag {
		entity = esiEntity.Type.ToMongoEntity()
	}
	entity.Etag = esiEntity.Etag
	entity.Expires = esiEntity.Expires

	switch create {
	case true:
		entry.Info("creeting entity")
		_, err = s.universe.CreateEntity(ctx, entity)
	case false:
		entry.Info("updating entity")
		_, err = s.universe.UpdateEntity(ctx, entity)
	}
	if err != nil {
		entry.WithError(err).Error("encountered error mutating entity in database")
	}

	return entity, nil
}

func (s *Service) Run() {

	var ctx = context.Background()

	var page uint = 1
	ids := make([]uint, 0, 4000)
	for {
		groupIDs, err := s.esi.Groups(ctx, page)
		if err != nil {
			s.logger.WithError(err).Error("failed to fetch groups from ESI")
			return
		}

		ids = append(ids, groupIDs.IDs...)

		if groupIDs.Pages == page {
			break
		}

		page++
		time.Sleep(time.Second)

	}

	for _, groupID := range ids {

		entry := s.logger.WithFields(logrus.Fields{
			"groupID": groupID,
			"service": "universe",
		})

		group, err := s.universe.Group(ctx, groupID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			entry.WithError(err).Error("failed to fetch group from mongo")
			return
		}

		if !group.Expires.IsZero() && group.Expires.Unix() > time.Now().UTC().Unix() {
			continue
		}
		var create = false
		if errors.Is(err, sql.ErrNoRows) {
			create = true
		}

		esiGroup, err := s.esi.Group(ctx, groupID, esi.AddIfNoneMatchHeader(group.Etag))
		if err != nil {
			entry.WithError(err).Error("failed to fetch group from ESI")
			time.Sleep(time.Millisecond * 250)
			continue
		}

		if esiGroup.Etag != group.Etag {
			group = esiGroup.Group.ToMongoGroup()
		}
		group.Etag = esiGroup.Etag
		group.Expires = esiGroup.Expires

		switch create {
		case true:
			entry.Info("creating group")
			_, err = s.universe.CreateGroup(ctx, group)
		case false:
			entry.Info("updating group")
			_, err = s.universe.UpdateGroup(ctx, group)
		}
		if err != nil {
			entry.WithError(err).Error("encountered error mutating group in database")
			return
		}

	}

}
