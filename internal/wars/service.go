package wars

import (
	"context"
	"fmt"

	"github.com/eveisesi/krinder"
	"github.com/eveisesi/krinder/internal/esi"
	mdb "github.com/eveisesi/krinder/internal/mongo"
	"github.com/sirupsen/logrus"
)

type Service struct {
	logger *logrus.Logger

	esi *esi.Service

	mdb.WarAPI
}

func NewService(logger *logrus.Logger, esi *esi.Service, warAPI mdb.WarAPI) *Service {
	return &Service{
		logger: logger,
		esi:    esi,

		WarAPI: warAPI,
	}
}

func (s *Service) Initialize() {

	s.logger.Info("initializing wars service")
	s.logger.Info("fetching known wars from mongo")

	ctx := context.Background()
	wars, err := s.Wars(ctx, krinder.NewOrderOperator("id", krinder.SortDesc))
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

	s.logger.WithField("numNewWars", len(newIDs)).Info("fetching new wars from ESI")
	var newWars = make([]*krinder.ESIWar, 0, len(newIDs))

	for i, id := range newIDs {
		war, err := s.esi.War(ctx, uint(id))
		if err != nil {
			fmt.Println(err)
			continue
		}

		newWars = append(newWars, war)
		if i%50 == 0 {
			fmt.Println(i)
		}
	}

	mongoWars := make([]*krinder.MongoWar, 0, len(newWars))
	for _, war := range newWars {
		mongoWars = append(mongoWars, war.ToMongoWar())
	}

	err = s.CreateWarBulk(ctx, mongoWars)
	if err != nil {
		s.logger.WithError(err).Error("failed to save wars to mongo")
	}

}

// Run is a function that satisfies the Cron Runner Interface
func Run() {

	// wars, err :

}
