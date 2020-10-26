package resolvers

import (
	"context"
	"github.com/graph-gophers/graphql-go"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	passwordHash, _ = bcrypt.GenerateFromPassword(
		[]byte("password"),
		bcrypt.DefaultCost,
	)
	longForm      = "Jan 2, 2006 at 3:04pm (MST)"
	sampleTime, _ = time.Parse(longForm, "Feb 3, 2013 at 7:54pm (PST)")
	users         = []User{
		{
			ID:       "0",
			Name:     "Christopher Leggett",
			Email:    "chris@leggett.dev",
			Password: string(passwordHash),
		},
	}
	links = []Link{
		{
			ID:          "0",
			CreatedAt:   graphql.Time{Time: sampleTime},
			URL:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
			PostedBy:    &users[0],
		},
	}
	votes = []Vote{
		{
			ID:   "0",
			User: &users[0],
			Link: &links[0],
		},
	}
)

type QueryResolver struct{}

func NewRoot() (*QueryResolver, error) {
	return &QueryResolver{}, nil
}

type LinkQueryArgs struct {
	ID graphql.ID
}

func (r QueryResolver) Link(ctx context.Context, args LinkQueryArgs) (*LinkResolver, error) {
	return NewLink(ctx, NewLinkArgs{ID: args.ID})
}

type LinksQueryArgs struct {
	Or  *[]string
	And *[]string
}

func (r QueryResolver) Links(ctx context.Context, args LinksQueryArgs) (*[]*LinkResolver, error) {
	return NewLinks(ctx, NewLinksArgs{Or: args.Or, And: args.And})
}
