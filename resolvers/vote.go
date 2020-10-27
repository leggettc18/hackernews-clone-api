package resolvers

import (
	"context"
	goErrors "errors"
	"github.com/graph-gophers/graphql-go"
)

type Vote struct {
	ID   graphql.ID
	User *User
	Link *Link
}

type VoteResolver struct {
	Vote Vote
}

type NewVoteArgs struct {
	ID graphql.ID
}

func NewVote(ctx context.Context, args NewVoteArgs) (*VoteResolver, error) {
	for _, vote := range votes {
		if vote.ID == args.ID {
			return &VoteResolver{Vote: vote}, nil
		}
	}
	return &VoteResolver{}, goErrors.New("ID not found")
}

func (r *VoteResolver) ID() graphql.ID {
	return r.Vote.ID
}

func (r *VoteResolver) User(ctx context.Context) (*UserResolver, error) {
	return NewUser(ctx, NewUserArgs{ID: r.Vote.User.ID})
}

func (r *VoteResolver) Link(ctx context.Context) (*LinkResolver, error) {
	return NewLink(ctx, NewLinkArgs{ID: r.Vote.Link.ID})
}
