package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/rs/cors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

var (
	opts  = []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	links = []Link{
		{
			ID:          "0",
			URL:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
		},
	}
)

type RootResolver struct{}

type Link struct {
	ID          graphql.ID
	Description string
	URL         string
}

func (r *RootResolver) Info() (string, error) {
	return "this is a thing", nil
}

func (r *RootResolver) Feed() ([]Link, error) {
	return links, nil
}

func (r *RootResolver) Link(args struct {
	ID graphql.ID
}) (Link, error) {
	for _, link := range links {
		if link.ID == args.ID {
			return link, nil
		}
	}
	return Link{
		ID:          "",
		Description: "",
		URL:         "",
	}, errors.New("ID not found")
}

func (r *RootResolver) Post(args struct {
	Description string
	URL         string
}) (Link, error) {
	newLink := Link{
		ID:          graphql.ID(fmt.Sprint(len(links))),
		Description: args.Description,
		URL:         args.URL,
	}

	links = append(links, newLink)
	return newLink, nil
}

// Reads and parses the schema from file.
// Associates root resolver. Panics if can't read.
func parseSchema(path string, resolver interface{}) *graphql.Schema {
	bstr, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	schemaString := string(bstr)
	parsedSchema, err := graphql.ParseSchema(
		schemaString,
		resolver,
		opts...,
	)
	if err != nil {
		panic(err)
	}
	return parsedSchema
}

func main() {
	mux := http.NewServeMux()

	// cors.Default() setup the middleware with default options being
	// all origins accepted with simple methods (GET, POST). See
	// documentation below for more options.
	handler := cors.Default().Handler(mux)

	mux.Handle("/graphql", &relay.Handler{
		Schema: parseSchema("./schema.graphql", &RootResolver{}),
	})

	fmt.Println("serving on 8081")
	log.Fatal(http.ListenAndServe(":8081", handler))
}
