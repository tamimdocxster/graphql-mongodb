package main

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"graphql-mongodb/graph"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultPort = "8080"

var client *mongo.Client

func main() {
	// MongoDB connection setup
	var err error
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb+srv://tamimdocxster:test1234@test-db.tnqyz.mongodb.net/?retryWrites=true&w=majority&appName=test-db"))
	if err != nil {
		log.Fatal(err)
	}

	// Fiber app setup
	app := fiber.New()

	// GraphQL setup
	schema := graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: client}})
	srv := handler.NewDefaultServer(schema)

	// Middleware to inject fiber.Ctx into context
	injectFiberCtx := func(c *fiber.Ctx) error {
		ctx := context.WithValue(c.Context(), "fiberCtx", c)
		c.Locals("ctx", ctx)
		return c.Next()
	}

	// GraphQL handler
	app.All("/graphql", injectFiberCtx, func(c *fiber.Ctx) error {
		ctx := c.Locals("ctx").(context.Context)
		h := adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(ctx)
			srv.ServeHTTP(w, r)
		})
		return h(c)
	})

	// Playground handler
	app.Get("/playground", injectFiberCtx, func(c *fiber.Ctx) error {
		ctx := c.Locals("ctx").(context.Context)
		h := adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(ctx)
			playground.Handler("GraphQL Playground", "/graphql").ServeHTTP(w, r)
		})
		return h(c)
	})

	port := defaultPort
	fmt.Printf("Connect to http://localhost:%s/ for GraphQL Playground\n", port)
	log.Fatal(app.Listen(":" + port))
}
