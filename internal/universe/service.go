package universe

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/eveisesi/krinder"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type UniverseAPI interface {
	Groups(ctx context.Context, operators ...*krinder.Operator) ([]*krinder.MySQLGroup, error)
}

type Service struct {
	cache  *redis.Client
	logger *logrus.Logger

	esi      esi.API
	universe krinder.UniverseRepository
	UniverseAPI
}

func New(logger *logrus.Logger, cache *redis.Client, esi esi.API, universe krinder.UniverseRepository) *Service {
	return &Service{
		logger:      logger,
		cache:       cache,
		esi:         esi,
		universe:    universe,
		UniverseAPI: universe,
	}
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
