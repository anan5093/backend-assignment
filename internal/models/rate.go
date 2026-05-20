package models

import "time"

type UserRateData struct {
	AcceptedTimestamps []time.Time
	RejectedCount      int
}

type RateRequest struct {
	UserID  string
	Payload any
}

type UserStats struct {
	AcceptedRequestsCurrentWindow int `json:"accepted_requests_current_window"`
	RejectedRequestsTotal         int `json:"rejected_requests_total"`
}
