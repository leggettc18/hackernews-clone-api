package resolvers

import (
	"fmt"
	"github.com/graph-gophers/graphql-go"
	"github.com/leggettc18/hackernews-clone-api/db"
	"github.com/leggettc18/hackernews-clone-api/model"
)

type UserResolver struct {
	DB   *db.DB
	User model.User
}

func (r *UserResolver) ID() graphql.ID {
	return graphql.ID(fmt.Sprint(r.User.ID))
}

func (r *UserResolver) Name() string {
	return r.User.Name
}

func (r *UserResolver) Email() string {
	return r.User.Email
}

func (r *UserResolver) Links() (*[]*LinkResolver, error) {
	resolvers := make([]*LinkResolver, len(r.User.Links))
	for _, link := range r.User.Links {
		if link.PosterID == r.User.ID {
			resolver := LinkResolver{r.DB, link}
			resolvers = append(resolvers, &resolver)
		}
	}
	return &resolvers, nil
}

func (r *UserResolver) Votes() (*[]*VoteResolver, error) {
	resolvers := make([]*VoteResolver, len(r.User.Votes))
	for _, vote := range r.User.Votes {
		if vote.UserID == r.User.ID {
			resolver := VoteResolver{r.DB, vote}
			resolvers = append(resolvers, &resolver)
		}
	}
	return &resolvers, nil
}
