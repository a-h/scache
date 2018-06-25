package main

import (
	"log"
	"os"
	"time"

	"github.com/a-h/scache/expiry"

	"github.com/a-h/scache"

	"github.com/a-h/scache/example/dynamo"
	"github.com/a-h/scache/example/user"
	"github.com/akrylysov/algnhsa"
)

func main() {
	// Create our user handler.
	region := os.Getenv("DYNAMODB_REGION")
	userTable := os.Getenv("DYNAMODB_USER_TABLE")
	db, err := dynamo.NewUserStore(region, userTable)
	if err != nil {
		log.Fatalf("unable to start up, error creating user store: %v", err)
	}
	uh := user.Handler{
		GetUser: db.Get,
		PutUser: db.Put,
	}

	// Wrap the handlers in scache middleware which handles stream processing.
	streamName := os.Getenv("KINESIS_STREAM_NAME")
	if streamName == "" {
		log.Fatal("unable to start up, environment variable KINESIS_STREAM_NAME is missing")
	}
	s := expiry.NewStream(streamName)
	// Cache for between 30 seconds and 1 minute.
	minCacheDuration, maxCacheDuration := time.Second*30, time.Minute
	h := scache.AddMiddleware(uh, s, minCacheDuration, maxCacheDuration)

	// Use the output from AddMiddleware to handle HTTP requests.
	algnhsa.ListenAndServe(h, nil)
}
