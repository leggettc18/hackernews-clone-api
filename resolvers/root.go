package resolvers

import (
	"context"
	"errors"
	"fmt"
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

type RootResolver struct{}

func NewRoot() (*RootResolver, error) {
	return &RootResolver{}, nil
}

type LinkQueryArgs struct {
	ID graphql.ID
}

func (r RootResolver) Link(ctx context.Context, args LinkQueryArgs) (*LinkResolver, error) {
	return NewLink(ctx, NewLinkArgs{ID: args.ID})
}

type PostArgs struct {
	Description string
	Url         string
}

func (r *RootResolver) Post(ctx context.Context, args PostArgs) (*LinkResolver, error) {
	token, ok := ctx.Value("token").(string)
	if !ok {
		return &LinkResolver{}, errors.New("Post: no key 'token' in context")
	}
	author, errAuthor := GetUserFromToken(token)
	if errAuthor != nil {
		return &LinkResolver{}, errAuthor
	}
	newLink := Link{
		ID:          graphql.ID(fmt.Sprint(len(links))),
		CreatedAt:   graphql.Time{time.Now()},
		Description: args.Description,
		URL:         args.Url,
		PostedBy:    author,
		Votes:       []Vote{},
	}

	links = append(links, newLink)
	return NewLink(ctx, NewLinkArgs{ID: newLink.ID})
}

type LinksQueryArgs struct {
	Or  *[]string
	And *[]string
}

func (r RootResolver) Links(ctx context.Context, args LinksQueryArgs) (*[]*LinkResolver, error) {
	return NewLinks(ctx, NewLinksArgs{Or: args.Or, And: args.And})
}

type SignupArgs struct {
	Email    string
	Password string
	Name     string
}

type UpvoteArgs struct {
	LinkID graphql.ID
}

func (r *RootResolver) Upvote(ctx context.Context, args UpvoteArgs) (*VoteResolver, error) {
	token, ok := ctx.Value("token").(string)
	if !ok {
		return &VoteResolver{}, errors.New("Post: no key 'token' in context")
	}
	voter, errVoter := GetUserFromToken(token)
	if errVoter != nil {
		return &VoteResolver{}, errVoter
	}
	var processedLinks = []Link{}
	for index, link := range links {
		processedLinks = append(processedLinks, link)
		for _, vote := range votes {
			if vote.Link.ID == processedLinks[index].ID {
				processedLinks[index].Votes = append(processedLinks[index].Votes, vote)
			}
		}
	}
	var votedLink Link
	for _, link := range processedLinks {
		if link.ID == args.LinkID {
			votedLink = link
		}
	}
	newVote := Vote{
		ID:   graphql.ID(fmt.Sprint((votes))),
		User: voter,
		Link: &votedLink,
	}
	votedLink.Votes = append(votedLink.Votes, newVote)
	votes = append(votes, newVote)
	return NewVote(ctx, NewVoteArgs{ID: newVote.ID})
}

func (r *RootResolver) Signup(args SignupArgs) (*AuthResolver, error) {
	passwordHash, errHash := bcrypt.GenerateFromPassword(
		[]byte(args.Password),
		bcrypt.DefaultCost,
	)
	if errHash != nil {
		return nil, errHash
	}

	newUser := User{
		ID:       graphql.ID(fmt.Sprint(len(users))),
		Email:    args.Email,
		Password: string(passwordHash),
		Name:     args.Name,
	}

	users = append(users, newUser)

	token, errToken := GenerateToken(&newUser)

	if errToken != nil {
		return nil, errToken
	}

	payload := AuthPayload{
		Token: &token,
		User:  &newUser,
	}

	return NewAuth(payload)
}

type LoginArgs struct {
	Email    string
	Password string
}

func (r *RootResolver) Login(args LoginArgs) (*AuthResolver, error) {
	user, errUser := getUser(args.Email, args.Password)
	if errUser != nil {
		return nil, errUser
	}

	token, errToken := GenerateToken(&user)
	if errToken != nil {
		return nil, errToken
	}
	payload := AuthPayload{
		Token: &token,
		User:  &user,
	}
	return NewAuth(payload)
}
