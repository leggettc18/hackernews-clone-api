package db

import (
	"github.com/leggettc18/hackernews-clone-api/model"
	"github.com/pkg/errors"
)

func (db *DB) GetVoteById(id uint) (*model.Vote, error) {
	var vote model.Vote
	return &vote, errors.Wrap(db.First(&vote, id).Error, "unable to get vote")
}

func (db *DB) GetVotesByLinkId(linkId uint) ([]*model.Vote, error) {
	var votes []*model.Vote
	return votes, errors.Wrap(db.Where("link_id = ?", linkId).Find(&links).Error, "unable to get votes")
}
