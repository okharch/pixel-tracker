package model

import "time"

// row event feed to endpoint
type TrackEvent struct {
	EmailHash string `json:"email_hash"`
	EventHash string `json:"event_hash"`
}

// processed event in database
type TrackEventU struct {
	UserId    int64 // this is hash of EmailHash
	EventHash string
	CreatedAt time.Time
}
