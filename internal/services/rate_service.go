package services

import (
	"time"

	"backend-assignment/internal/models"
	"backend-assignment/internal/store"
)

type RateService struct {
	store *store.RateStore
}

func NewRateService(store *store.RateStore) *RateService {
	return &RateService{store: store}
}

func (s *RateService) Accept(userID string) bool {
	return s.store.Allow(userID, time.Now().UTC())
}

func (s *RateService) Stats() map[string]models.UserStats {
	return s.store.Stats(time.Now().UTC())
}
