package model

import "time"

type Link struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	Url         string    `json:"url"`
	PosterID    uint      `json:"poster_id"`
	Votes       []Vote    `json:"votes"`
}
