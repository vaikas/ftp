package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/kelseyhightower/envconfig"

	"github.com/knative/pkg/cloudevents"
	"github.com/knative/pkg/signals"
)

var (
	sink  string
	query string
)

type EnvConfig struct {
	ConsumerKey       string `split_words:"true",required:"true"`
	ConsumerSecretKey string `split_words:"true",required:"true"`
	AccessToken       string `split_words:"true",required:"true"`
	AccessSecret      string `split_words:"true",required:"true"`
}

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

	var s EnvConfig
	err = envconfig.Process("twitter", &s)
	if err != nil {
		logger.Fatal(err.Error())
	}

	if query == "" {
		logger.Error("Need to specify query string")
		return
	}

	logger.Info("Conf: ", zap.String("consumerKey", s.ConsumerKey), zap.String("consumerSecretKey", s.ConsumerSecretKey), zap.String("accessToken", s.AccessToken), zap.String("accessSecret", s.AccessSecret))
	logger.Info("Starting and publishing to sink", zap.String("sink", sink))
	logger.Info("querying for ", zap.String("query", query))

	ceClient := cloudevents.NewClient(sink, cloudevents.Builder{
		EventType: "con.twitter",
		Source:    "com.twitter",
	})

	publisher := publisher{ceClient: ceClient, logger: logger}

	config := oauth1.NewConfig(s.ConsumerKey, s.ConsumerSecretKey)
	token := oauth1.NewToken(s.AccessToken, s.AccessSecret)
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
	logger   *zap.Logger
	ceClient *cloudevents.Client
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

	return p.ceClient.Send(tweet, cloudevents.V01EventContext{
		EventID:   strconv.FormatInt(tweet.ID, 10),
		EventTime: eventTime,
	})
}
