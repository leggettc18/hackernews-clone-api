package resolvers

import (
	"context"
	"errors"
	"fmt"
	"github.com/graph-gophers/graphql-go"
	"github.com/leggettc18/hackernews-clone-api/db"
	"github.com/leggettc18/hackernews-clone-api/model"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"strings"
	"time"
)

type RootResolver struct {
	DB       *db.DB
	NewLinks chan *LinkResolver
}

func NewRoot(db *db.DB) (*RootResolver, error) {
	return &RootResolver{DB: db, NewLinks: make(chan *LinkResolver)}, nil
}

func (r *RootResolver) NewLink() (chan *LinkResolver, error) {
	fmt.Println("subscribing to new links")
	return r.NewLinks, nil
}

type LinkQueryArgs struct {
	ID graphql.ID
}

func (r RootResolver) Link(args LinkQueryArgs) (*LinkResolver, error) {
	id, err := getUintFromGraphqlId(args.ID)
	if err != nil {
		return nil, err
	}
	link, err := r.DB.GetLinkById(id)
	if err != nil {
		return nil, err
	}
	linkResolver := LinkResolver{
		DB:   r.DB,
		Link: *link,
	}

	return &linkResolver, nil
}

type PostArgs struct {
	Description string
	Url         string
}

func (r *RootResolver) Post(ctx context.Context, args PostArgs) (*LinkResolver, error) {
	token, ok := ctx.Value("token").(string)
	if !ok {
		return &LinkResolver{}, errors.New("post: no key 'token' in context")
	}
	author, errAuthor := r.DB.GetUserFromToken(token)
	if errAuthor != nil {
		return &LinkResolver{}, errAuthor
	}
	newLink := model.Link{
		CreatedAt:   time.Now(),
		Description: args.Description,
		Url:         args.Url,
		PosterID:    author.ID,
		Votes:       []model.Vote{},
	}

	if err := r.DB.CreateLink(&newLink); err != nil {
		return nil, err
	}
	linkResolver := &LinkResolver{DB: r.DB, Link: newLink}

	select {
	case r.NewLinks <- linkResolver:
		// values are being read from r.Events
		fmt.Println("r.NewLinks: inserted link")
	default:
		//no subscribers, link not in channel
		fmt.Println("r.NewLinks: link created, not inserted")
	}

	return linkResolver, nil
}

type LinksQueryArgs struct {
	Or  *[]string
	And *[]string
}

func (r RootResolver) Links(args LinksQueryArgs) (*[]*LinkResolver, error) {
	var (
		links   []model.Link
		results []model.Link
	)
	if err := r.DB.Find(&links).Error; err != nil {
		return nil, err
	}
	if args.And != nil {
		for _, link := range links {
			hasAllTerms := true
			for _, term := range *args.And {
				if hasAllTerms == false {
					break
				}
				if strings.Contains(link.Description, term) || strings.Contains(link.Url, term) {
					hasAllTerms = true
				} else {
					hasAllTerms = false
				}
			}
			if hasAllTerms == true {
				results = append(results, link)
			}
		}
	} else if args.Or != nil {
		for _, link := range links {
			for _, term := range *args.Or {
				if strings.Contains(link.Description, term) || strings.Contains(link.Url, term) {
					results = append(results, link)
					break
				}
			}
		}
	} else {
		results = links
	}
	var resolvers []*LinkResolver
	for _, link := range results {
		resolvers = append(resolvers, &LinkResolver{r.DB, link})
	}
	return &resolvers, nil
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
		return &VoteResolver{}, errors.New("post: no key 'token' in context")
	}
	voter, errVoter := r.DB.GetUserFromToken(token)
	if errVoter != nil {
		return &VoteResolver{}, errVoter
	}
	id, err := getUintFromGraphqlId(args.LinkID)
	if err != nil {
		return nil, err
	}
	link, err := r.DB.GetLinkById(id)
	if err != nil {
		return nil, err
	}
	vote := model.Vote{LinkID: link.ID, UserID: voter.ID}
	if err := r.DB.CreateVote(&vote); err != nil {
		return nil, err
	}
	return &VoteResolver{DB: r.DB, Vote: vote}, nil
}

func (r *RootResolver) Signup(args SignupArgs) (*AuthResolver, error) {
	passwordHash, errHash := bcrypt.GenerateFromPassword(
		[]byte(args.Password),
		bcrypt.DefaultCost,
	)
	if errHash != nil {
		return nil, errHash
	}

	newUser := model.User{
		Email:          args.Email,
		HashedPassword: passwordHash,
		Name:           args.Name,
	}

	if err := r.DB.CreateUser(&newUser); err != nil {
		return nil, err
	}

	token, errToken := GenerateToken(&newUser)

	if errToken != nil {
		return nil, errToken
	}

	payload := AuthPayload{
		Token: &token,
		User:  &newUser,
	}

	return &AuthResolver{r.DB, payload}, nil
}

type LoginArgs struct {
	Email    string
	Password string
}

func (r *RootResolver) Login(args LoginArgs) (*AuthResolver, error) {
	user, errUser := r.DB.GetUserByEmail(args.Email)
	if errUser != nil {
		return nil, errUser
	}
	model.ComparePasswordHash(user.HashedPassword, []byte(args.Password))

	token, errToken := GenerateToken(user)
	if errToken != nil {
		return nil, errToken
	}
	payload := AuthPayload{
		Token: &token,
		User:  user,
	}
	return &AuthResolver{r.DB, payload}, nil
}

//Helpers
func getUintFromGraphqlId(gqlid graphql.ID) (uint, error) {
	id, err := strconv.ParseUint(string(gqlid), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
