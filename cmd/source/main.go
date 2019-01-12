package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/knative/pkg/signals"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"

	"github.com/knative/pkg/cloudevents"
)

const (
	consumerKeyKey    = "TWITTER_CONSUMER_KEY"
	consumerSecretKey = "TWITTER_CONSUMER_SECRET_KEY"
	accessTokenKey    = "TWITTER_ACCESS_TOKEN"
	accessSecretKey   = "TWITTER_ACCESS_SECRET"
)

var (
	sink  string
	query string
)

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")
	flag.StringVar(&query, "query", "", "query string to look for")
}

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	logConfig := zap.NewProductionConfig()
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := logConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %s", err))
	}

	consumerKey := os.Getenv(consumerKeyKey)
	consumerSecret := os.Getenv(consumerSecretKey)
	accessToken := os.Getenv(accessTokenKey)
	accessSecret := os.Getenv(accessSecretKey)

	if consumerKey == "" || consumerSecret == "" || accessToken == "" || accessSecret == "" {
		logger.Error("need to specify all of : consumerKey, consumerSecret, accessToken and accessSecret or no twitter for you")
		return
	}

	if query == "" {
		logger.Error("Need to specify query string")
		return
	}

	logger.Info("Starting and publishing to sink", zap.String("sink", sink))
	logger.Info("querying for ", zap.String("query", query))

	publisher := publisher{sinkURI: sink, logger: logger}

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	done := make(chan bool)

	searcher := NewSearcher(client, query, 5, publisher.postMessage, done)
	searcher.run()
	<-stopCh
	done <- true
}

type publisher struct {
	logger  *zap.Logger
	sinkURI string
}

type simpleTweet struct {
	user string `json:"user"`
	text string `json:"text"`
}

func (p *publisher) postMessage(tweet *twitter.Tweet) error {
	eventTime, err := time.Parse(time.RubyDate, tweet.CreatedAt)

	if err != nil {
		p.logger.Info("Failed to parse created at: ", zap.Error(err))
		eventTime = time.Now()
	}

	eventCtx := cloudevents.EventContext{
		CloudEventsVersion: cloudevents.CloudEventsVersion,
		EventType:          "com.twitter",
		EventID:            strconv.FormatInt(tweet.ID, 10),
		EventTime:          eventTime,
		ContentType:        "application/json",
		Source:             "com.twitter",
	}

	req, err := cloudevents.Binary.NewRequest(p.sinkURI, &tweet, eventCtx)
	if err != nil {
		p.logger.Info("Failed to MARSHAL: ", zap.Error(err))
		return err
	}

	p.logger.Info("Posting tweet: ", zap.String("sinkURI", p.sinkURI), zap.String("user", tweet.User.Name), zap.String("tweet", tweet.Text))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// TODO: in general, receive adapters may have to be able to retry for error cases.
		p.logger.Info("Response Status ", zap.String("response status", resp.Status))
		body, _ := ioutil.ReadAll(resp.Body)
		p.logger.Info("response Body:", zap.String("body", string(body)))
	}
	return nil
}
