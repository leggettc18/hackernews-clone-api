package resolvers

import (
	"fmt"
	"github.com/graph-gophers/graphql-go"
	"github.com/leggettc18/hackernews-clone-api/db"
	"github.com/leggettc18/hackernews-clone-api/model"
)

type LinkResolver struct {
	DB   *db.DB
	Link model.Link
}

func (r *LinkResolver) ID() graphql.ID {
	return graphql.ID(fmt.Sprint(r.Link.ID))
}

func (r *LinkResolver) CreatedAt() graphql.Time {
	return graphql.Time{Time: r.Link.CreatedAt}
}

func (r *LinkResolver) Description() string {
	return r.Link.Description
}

func (r *LinkResolver) Url() string {
	return r.Link.Url
}

func (r *LinkResolver) PostedBy() (*UserResolver, error) {
	user, err := r.DB.GetUserById(r.Link.PosterID)
	if err != nil {
		return nil, err
	}
	return &UserResolver{r.DB, *user}, nil
}

func (r *LinkResolver) Votes() (*[]*VoteResolver, error) {
	var (
		resolvers []*VoteResolver
	)
	if err := r.DB.Model(&r.Link).Related(&r.Link.Votes).Error; err != nil {
		return nil, err
	}
	for _, vote := range r.Link.Votes {
		if vote.LinkID == r.Link.ID {
			resolver := VoteResolver{r.DB, vote}
			resolvers = append(resolvers, &resolver)
		}
	}
	return &resolvers, nil
}
