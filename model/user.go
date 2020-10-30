package model

type User struct {
	ID             uint   `gorm:"primary_key" json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Links          []Link `json:"links"`
	Votes          []Vote `json:"votes"`
	HashedPassword []byte `json:"-"`
}
