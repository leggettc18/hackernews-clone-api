package model

type Vote struct {
	ID   uint `json:"id"`
	User User `json:"user"`
	Link Link `json:"link"`
}
