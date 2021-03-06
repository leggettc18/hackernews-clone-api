package main

import (
	"context"
	"github.com/leggettc18/hackernews-clone-api/db"
	"github.com/leggettc18/hackernews-clone-api/resolvers"
	"github.com/rs/cors"

	//"errors"
	//"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	//"golang.org/x/crypto/bcrypt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/graph-gophers/graphql-transport-ws/graphqlws"
)

var (
	opts = []graphql.SchemaOpt{graphql.UseStringDescriptions()}
)

var (
	addr              = ":8081"
	readHeaderTimeout = 1 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 90 * time.Second
	maxHeaderBytes    = http.DefaultMaxHeaderBytes
)

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

	database, err := db.NewDB("./db.sqlite")

	if err != nil {
		panic(err)
	}

	rootResolver, err := resolvers.NewRoot(database)

	if err != nil {
		panic(err)
	}

	schema := parseSchema("./schema.graphql", rootResolver)
	wsHandler := graphqlws.NewHandlerFunc(
		schema,
		&relay.Handler{
			Schema: schema,
		},
	)

	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		token := strings.ReplaceAll(r.Header.Get("Authorization"), "Bearer ", "")
		ctx := context.WithValue(context.Background(), "token", token)
		wsHandler.ServeHTTP(w, r.WithContext(ctx))
	})

	// necessary CORS options. Should not be used in production
	// AllowedOrigins should be more specific than * and the
	// AllowOriginFunc should not be present. This code is not
	// written for production or for a CORS tutorial so this is fine
	// for its purpose.
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowOriginFunc:  func(origin string) bool { return true },
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	}).Handler(mux)

	s := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}

	// Begin listeing for requests.
	log.Printf("Listening for requests on %s", s.Addr)

	if err := s.ListenAndServe(); err != nil {
		log.Println("server.ListenAndServe:", err)
	}
}
