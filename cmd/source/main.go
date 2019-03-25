package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/knative/pkg/signals"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	sink      string
	dir       string
	server    string
	secure    bool
	storename string
)

type EnvConfig struct {
	User     string `split_words:"true",required:"true"`
	Password string `split_words:"true",required:"true"`
}

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")
	flag.StringVar(&dir, "dir", ".", "directory to watch files in")
	flag.StringVar(&server, "server", "", "server to connect to")
	flag.StringVar(&storename, "storename", "", "ConfigMap name to use for storing state")
	flag.BoolVar(&secure, "secure", true, "if set to true, use sftp")
}

type FTPFileEvent struct {
	Name    string
	Size    int64
	ModTime time.Time
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
	err = envconfig.Process("ftp", &s)
	if err != nil {
		logger.Fatal(err.Error())
	}

	if server == "" {
		logger.Error("Need to specify server string")
		return
	}

	logger.Info("Starting and publishing to sink", zap.String("sink", sink))
	logger.Info("Storing state in  ", zap.String("ConfigMap", storename))
	logger.Info("Using sftp", zap.Bool("sftp", secure))
	logger.Info("watching ", zap.String("server", server), zap.String("dir", dir))

	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		logger.Info("Failed to create k8s config ", zap.Error(err))
		return
	}
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Info("Failed to create k8s client ", zap.Error(err))
		return
	}

	configStore := NewConfigStore(storename, "default", clientSet.CoreV1().ConfigMaps("default"))

	t, err := cloudeventshttp.New(
		cloudeventshttp.WithTarget(sink),
		cloudeventshttp.WithEncoding(cloudeventshttp.BinaryV02),
	)

	if err != nil {
		logger.Info("Failed to create the cloudevents http transport", zap.Error(err))
		return
	}

	ceClient, err := client.New(t, client.WithTimeNow())
	if err != nil {
		logger.Info("failed to create client", zap.Error(err))
		return
	}

	publisher := publisher{ceClient: ceClient, logger: logger, sourceServer: fmt.Sprintf("ftp://%s%s/", server, dir)}

	done := make(chan bool)

	searcher := NewWatcher(server, dir, s.User, s.Password, secure, 5, publisher.postMessage, done, configStore)
	searcher.run()
	<-stopCh
	done <- true
}

type publisher struct {
	logger       *zap.Logger
	ceClient     client.Client
	sourceServer string
}

func (p *publisher) postMessage(fileEntry os.FileInfo) error {
	source := fmt.Sprintf("%s%s", p.sourceServer, fileEntry.Name())

	e := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			ID:     uuid.New().String(),
			Type:   "org.aikas.ftp.fileadded",
			Source: *types.ParseURLRef(source),
			Time:   &types.Timestamp{fileEntry.ModTime()},
		}.AsV02(),
		Data: FTPFileEvent{Name: fileEntry.Name(), Size: fileEntry.Size(), ModTime: fileEntry.ModTime()},
	}

	if _, err := p.ceClient.Send(context.Background(), e); err != nil {
		p.logger.Info("Failed to send event:", zap.Error(err))
		return err
	}
	return nil
}
