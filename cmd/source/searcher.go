package main

import (
	"fmt"
	"time"

	"github.com/dghubble/go-twitter/twitter"
)

type searcher struct {
	client    *twitter.Client
	query     string
	frequency int64 // in seconds
	handler   func(*twitter.Tweet) error
	stop      chan bool
	sinceID   int64 // keep track of what tweets we've seen so far
}

func NewSearcher(client *twitter.Client, query string, frequency int64, handler func(*twitter.Tweet) error, stop chan bool) *searcher {
	return &searcher{client: client, query: query, frequency: frequency, handler: handler, stop: stop}
}

func (s *searcher) run() {
	//	tickChan := time.NewTicker(s.frequency * time.Second).C
	tickChan := time.NewTicker(10 * time.Second).C
	go func() {
		for {
			select {
			case <-tickChan:
				s.search()
				fmt.Println("Fetching tweets")
			case <-s.stop:
				fmt.Println("Exiting")
				return
			}
		}

	}()
}

func (s *searcher) search() {
	search, resp, err := s.client.Search.Tweets(&twitter.SearchTweetParams{
		Query:           s.query,
		Lang:            "en",
		Count:           100,
		SinceID:         s.sinceID,
		IncludeEntities: twitter.Bool(true),
	})

	if err != nil {
		fmt.Printf("Error executing search %s - %v", resp.Status, err)
	}
	for _, t := range search.Statuses {
		handlerErr := s.handler(&t)
		if handlerErr != nil {
			fmt.Printf("Failed to post: %s", err)
			break
		}
		if t.ID > s.sinceID {
			s.sinceID = t.ID
		}
	}
}
