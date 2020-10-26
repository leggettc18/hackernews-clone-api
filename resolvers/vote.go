package resolvers

import "github.com/graph-gophers/graphql-go"

type Vote struct {
	ID   graphql.ID
	User *User
	Link *Link
}
