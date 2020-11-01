package resolvers

import (
	"fmt"
	"github.com/graph-gophers/graphql-go"
	"github.com/leggettc18/hackernews-clone-api/db"
	"github.com/leggettc18/hackernews-clone-api/model"
)

type VoteResolver struct {
	DB   *db.DB
	Vote model.Vote
}

func (r *VoteResolver) ID() graphql.ID {
	return graphql.ID(fmt.Sprint(r.Vote.ID))
}

func (r *VoteResolver) User() (*UserResolver, error) {
	user, err := r.DB.GetUserById(r.Vote.UserID)
	if err != nil {
		return nil, err
	}
	return &UserResolver{r.DB, *user}, nil
}

func (r *VoteResolver) Link() (*LinkResolver, error) {
	link, err := r.DB.GetLinkById(r.Vote.LinkID)
	if err != nil {
		return nil, err
	}
	return &LinkResolver{r.DB, *link}, nil
}
