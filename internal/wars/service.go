package wars

import (
	"context"
	"fmt"
	"time"

	"github.com/eveisesi/krinder"
	"github.com/eveisesi/krinder/internal/esi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Service struct {
	logger *logrus.Logger

	esi esi.API

	wars krinder.WarRepository
}

func NewService(logger *logrus.Logger, esi esi.API, wars krinder.WarRepository) *Service {
	return &Service{
		logger: logger,
		esi:    esi,

		wars: wars,
	}
}

func (s *Service) Run() {
	s.checkForNewWars()
	s.updateWars()
}

type Entity struct {
	T  string // Entity Type, must be either corporation or alliance
	ID uint
}

func (s *Service) EntitiesAtWar(ctx context.Context, entityA, entityB Entity, killTime time.Time) (bool, error) {

	const (
		aggressorAllianceID    = "aggressor.allianceID"
		aggressorCorporationID = "aggressor.corporationID"
		defenderAllianceID     = "defender.allianceID"
		defenderCorporationID  = "defender.corporationID"
	)

	filters := make([]*krinder.Operator, 0, 4)
	// Handle Entity A first
	switch entityA.T {
	case "corporation":
		filters = append(filters, krinder.NewOrOperator(krinder.NewEqualOperator(aggressorCorporationID, entityA.ID), krinder.NewEqualOperator(defenderCorporationID, entityA.ID)))
	case "alliance":
		filters = append(filters, krinder.NewOrOperator(krinder.NewEqualOperator(aggressorAllianceID, entityA.ID), krinder.NewEqualOperator(defenderAllianceID, entityA.ID)))
	}

	switch entityB.T {
	case "corporation":
		filters = append(filters, krinder.NewOrOperator(krinder.NewEqualOperator(aggressorCorporationID, entityB.ID), krinder.NewEqualOperator(defenderCorporationID, entityB.ID)))
	case "alliance":
		filters = append(filters, krinder.NewOrOperator(krinder.NewEqualOperator(aggressorAllianceID, entityB.ID), krinder.NewEqualOperator(defenderAllianceID, entityB.ID)))
	}

	filters = append(filters, krinder.NewLessThanOperator("started", killTime), krinder.NewOrOperator(krinder.NewExistsOperator("finished", false), krinder.NewGreaterThanOperator("finished", killTime)))

	wars, err := s.wars.Wars(ctx, krinder.NewAndOperator(filters...))
	if err != nil {
		return false, errors.Wrap(err, "failed to fetch wars for provided entities")
	}

	return len(wars) > 0, nil

}

func (s *Service) updateWars() {

	ctx := context.Background()
	esiWars, err := s.wars.Wars(ctx, krinder.NewExistsOperator("finished", false), krinder.NewLessThanOperator("expiresAt", time.Now().UTC()))
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch wars to update")
		return
	}

	s.logger.WithField("updatedableWars", len(esiWars)).Info("updating wars")

	var updatedMongoWars = make([]*krinder.MongoWar, 0, len(esiWars))
	for i, esiWar := range esiWars {
		if i%50 == 0 {
			s.logger.WithField("iteration", i).Infoln()
		}

		war, err := s.esi.War(ctx, esiWar.ID, esi.AddIfNoneMatchHeader(esiWar.IntegrityHash))
		if err != nil {
			s.logger.WithError(err).WithField("id", esiWar.ID).Error("failed to fetch War from ESI")
			continue
		}

		if war == nil {
			continue
		}

		updatedMongoWars = append(updatedMongoWars, war.ToMongoWar())

	}

	s.logger.WithField("countUpdatedWars", len(updatedMongoWars)).Infoln()

	for _, mongoWar := range updatedMongoWars {
		s.logger.WithField("id", mongoWar.ID).Info("updating war")
		err := s.wars.UpdateWar(ctx, mongoWar)
		if err != nil {
			s.logger.WithError(err).Error("failed to update war")
		}
	}

}

func (s *Service) checkForNewWars() {
	s.logger.Info("initializing wars service")
	s.logger.Info("fetching known wars from mongo")

	ctx := context.Background()
	wars, err := s.wars.Wars(ctx, krinder.NewOrderOperator("id", krinder.SortDesc))
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch known wars from mongo")
		return
	}

	var lastKnownWar int64
	if len(wars) > 0 {
		lastKnownWar = int64(wars[0].ID)
	}

	s.logger.Info("fetching wars from ESI")

	warIDs, err := s.esi.Wars(ctx)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch warIDs from ESI")
		return
	}

	var newIDs = make([]int, 0, len(warIDs))
	for _, id := range warIDs {
		if id > int(lastKnownWar) {
			newIDs = append(newIDs, id)
		}
	}

	if len(newIDs) == 0 {
		s.logger.Info("no new wars returns from ESI")
		return
	}

	s.logger.WithField("numNewWars", len(newIDs)).Info("fetching new wars from ESI. ")
	s.logger.Info("This could take a minute, especially if redis has been cleared recently and mongo is empty")
	var newWars = make([]*krinder.ESIWar, 0, len(newIDs))

	for _, id := range newIDs {
		war, err := s.esi.War(ctx, uint(id))
		if err != nil {
			fmt.Println(err)
			continue
		}

		newWars = append(newWars, war)
	}

	mongoWars := make([]*krinder.MongoWar, 0, len(newWars))
	for _, war := range newWars {
		mongoWars = append(mongoWars, war.ToMongoWar())
	}

	err = s.wars.CreateWarBulk(ctx, mongoWars)
	if err != nil {
		s.logger.WithError(err).Error("failed to save wars to mongo")
	}
}
