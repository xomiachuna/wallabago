package handlers

import (
	"github.com/andriihomiak/wallabago/internal/database"
)

type TimeService struct {
	querier database.Querier
}

// todo: rename
func NewService(querier database.Querier) *TimeService {
	return &TimeService{
		querier: querier,
	}
}
