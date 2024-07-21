package model

import "time"

type Task struct {
	ID     int       `json:"id"`
	Time   time.Time `json:"time"`
	UserID string    `json:"user_id"`
	Data   []byte    `json:"data"`
}
