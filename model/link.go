package model

import "time"

type Link struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	Url         string    `json:"url"`
	PostedBy    User      `json:"postedBy"`
	Votes       []Vote    `json:"votes"`
}
