package store

import (
	"sync"
	"time"

	"backend-assignment/internal/models"
)

type RateStore struct {
	mu     sync.RWMutex
	users  map[string]*models.UserRateData
	limit  int
	window time.Duration
}

func NewRateStore(limit int, window time.Duration) *RateStore {
	return &RateStore{
		users:  make(map[string]*models.UserRateData),
		limit:  limit,
		window: window,
	}
}

func (s *RateStore) Allow(userID string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := s.getOrCreateUser(userID)
	data.AcceptedTimestamps = pruneTimestamps(data.AcceptedTimestamps, now, s.window)

	if len(data.AcceptedTimestamps) >= s.limit {
		data.RejectedCount++
		return false
	}

	data.AcceptedTimestamps = append(data.AcceptedTimestamps, now)
	return true
}

func (s *RateStore) Stats(now time.Time) map[string]models.UserStats {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats := make(map[string]models.UserStats, len(s.users))
	for userID, data := range s.users {
		data.AcceptedTimestamps = pruneTimestamps(data.AcceptedTimestamps, now, s.window)
		stats[userID] = models.UserStats{
			AcceptedRequestsCurrentWindow: len(data.AcceptedTimestamps),
			RejectedRequestsTotal:         data.RejectedCount,
		}
	}

	return stats
}

func (s *RateStore) getOrCreateUser(userID string) *models.UserRateData {
	data, ok := s.users[userID]
	if !ok {
		data = &models.UserRateData{}
		s.users[userID] = data
	}
	return data
}

func pruneTimestamps(timestamps []time.Time, now time.Time, window time.Duration) []time.Time {
	cutoff := now.Add(-window)
	writeIndex := 0
	for _, timestamp := range timestamps {
		if timestamp.After(cutoff) {
			timestamps[writeIndex] = timestamp
			writeIndex++
		}
	}
	return timestamps[:writeIndex]
}
