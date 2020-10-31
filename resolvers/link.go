package resolvers

import (
	goErrors "errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/leggettc18/hackernews-clone-api/db"
	"strings"

	"github.com/leggettc18/hackernews-clone-api/errors"
)

type LinkResolver struct {
	DB   *db.DB
	Link Link
}

type NewLinkArgs struct {
	ID graphql.ID
}

func NewLink(args NewLinkArgs) (*LinkResolver, error) {
	for _, link := range links {
		if link.ID == args.ID {
			for _, vote := range votes {
				if vote.Link.ID == link.ID {
					link.Votes = append(link.Votes, vote)
				}
			}
			return &LinkResolver{Link: link}, nil
		}
	}
	return &LinkResolver{Link{}}, goErrors.New("ID not found")
}

type NewLinksArgs struct {
	Or  *[]string
	And *[]string
}

func NewLinks(args NewLinksArgs) (*[]*LinkResolver, error) {
	var processedLinks []Link

	if args.Or != nil && args.And == nil {
		for _, link := range links {
			for _, option := range *args.Or {
				if strings.Contains(link.URL, option) {
					processedLinks = append(processedLinks, link)
				} else if strings.Contains(link.Description, option) {
					processedLinks = append(processedLinks, link)
				}
			}
		}
	}
	if args.And != nil {
		containsAll := true
		for _, link := range processedLinks {
			for _, option := range *args.And {
				if containsAll == false {
					break
				}
				if strings.Contains(link.URL, option) {
					containsAll = true
				} else {
					containsAll = false
				}
			}
			if containsAll == true {
				processedLinks = append(processedLinks, link)
			} else {
				for _, option := range *args.And {
					if containsAll == false {
						break
					}
					if strings.Contains(link.Description, option) {
						containsAll = true
					} else {
						containsAll = false
					}
				}
			}
		}
	} else {
		processedLinks = links
	}
	for index := range processedLinks {
		for _, vote := range votes {
			if vote.Link.ID == processedLinks[index].ID {
				processedLinks[index].Votes = append(processedLinks[index].Votes, vote)
			}
		}
	}

	var (
		resolvers = make([]*LinkResolver, 0, len(processedLinks))
		errs      errors.Errors
	)
	for index, link := range processedLinks {
		resolver, err := NewLink(NewLinkArgs{ID: link.ID})
		if err != nil {
			errs = append(errs, errors.WithIndex(err, index))
		}
		resolvers = append(resolvers, resolver)
	}
	if errs != nil {
		return &resolvers, errs.Err()
	}
	return &resolvers, nil
}

func (r *LinkResolver) ID() graphql.ID {
	return r.Link.ID
}

func (r *LinkResolver) CreatedAt() graphql.Time {
	return r.Link.CreatedAt
}

func (r *LinkResolver) Description() string {
	return r.Link.Description
}

func (r *LinkResolver) Url() string {
	return r.Link.URL
}

func (r *LinkResolver) PostedBy() (*UserResolver, error) {
	return NewUser(NewUserArgs{ID: r.Link.PostedBy.ID})
}

func (r *LinkResolver) Votes() (*[]*VoteResolver, error) {
	var (
		resolvers []*VoteResolver
		errs      errors.Errors
	)
	for index, vote := range votes {
		if vote.Link.ID == r.Link.ID {
			resolver, err := NewVote(NewVoteArgs{ID: vote.ID})
			if err != nil {
				errs = append(errs, errors.WithIndex(err, index))
			} else {
				resolvers = append(resolvers, resolver)
			}
		}
	}
	if errs != nil {
		return &resolvers, errs.Err()
	}
	return &resolvers, nil
}
