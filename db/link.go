package db

import (
	"fmt"
	"github.com/leggettc18/hackernews-clone-api/model"
	"github.com/pkg/errors"
)

func (db *DB) GetLinkById(id uint) (*model.Link, error) {
	var link model.Link
	return &link, errors.Wrap(db.First(&link, id).Error, "unable to get link")
}

func (db *DB) SearchLinksByDescription(search string) ([]*model.Link, error) {
	var links []*model.Link
	return links, errors.Wrap(db.Where("description LIKE ?", fmt.Sprintf("%%%s%%", search)).Find(&links).Error, "unable to get links")
}

func (db *DB) SearchLinksByUrl(search string) ([]*model.Link, error) {
	var links []*model.Link
	return links, errors.Wrap(db.Where("url LIKE ?", fmt.Sprintf("%%%s%%", search)).Find(&links).Error, "unable to get links")
}

func (db *DB) CreateLink(link *model.Link) error {
	return errors.Wrap(db.Create(link).Error, "unable to create link")
}
