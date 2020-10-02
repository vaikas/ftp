package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"go.uber.org/zap"

	"knative.dev/eventing/pkg/adapter/v2"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/source"
	"knative.dev/pkg/system"
)

const (
	log_config = `{
					        "level": "info",
					        "development": false,
					        "outputPaths": ["stdout"],
					        "errorOutputPaths": ["stderr"],
					        "encoding": "json",
					        "encoderConfig": {
					          "timeKey": "ts",
					          "levelKey": "level",
					          "nameKey": "logger",
					          "callerKey": "caller",
					          "messageKey": "msg",
					          "stacktraceKey": "stacktrace",
					          "lineEnding": "",
					          "levelEncoder": "",
					          "timeEncoder": "iso8601",
					          "durationEncoder": "",
					          "callerEncoder": ""
					        }
      					}`

	event_type = "org.aikas.ftp.fileadded"
)

var (
	dir            string
	sftpServer     string
	secure         bool
	storename      string
	probeFrequency int
)

type EnvConfig struct {
	adapter.EnvConfig

	User     string `envconfig:"FTP_USER" required:"false"`
	Password string `envconfig:"FTP_PASSWORD" required:"false"`
}

func init() {
	flag.StringVar(&dir, "dir", ".", "directory to watch files in")
	flag.StringVar(&sftpServer, "sftpServer", "", "server to connect to")
	flag.StringVar(&storename, "storename", "", "ConfigMap name to use for storing state")
	flag.IntVar(&probeFrequency, "probeFrequency", 10, "interval in seconds between two probes")
	flag.BoolVar(&secure, "secure", true, "if set to true, use sftp")
}

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	logger, _ := logging.NewLogger(log_config, "")
	ctx := logging.WithLogger(context.Background(), logger)

	e := adapter.ConstructEnvOrDie(NewEnvConfig)
	env := e.(*EnvConfig)

	if sftpServer == "" {
		logger.Error("Need to specify server string")
		return
	}

	logger.Info("Storing state in  ", zap.String("ConfigMap", storename))
	logger.Info("Using sftp", zap.Bool("sftp", secure))
	logger.Info("watching ", zap.String("server", sftpServer), zap.String("dir", dir))
	logger.Info("Probing frequency  ", zap.Int("probeFrequency", probeFrequency))

	ctx, _ = injection.Default.SetupInformers(ctx, sharedmain.ParseAndGetConfigOrDie())

	configStore := NewConfigStore(storename, env.Namespace, kubeclient.Get(ctx).CoreV1().ConfigMaps(system.Namespace()))

	ceOverrides, err := env.GetCloudEventOverrides()
	if err != nil {
		logger.Error("Error loading cloudevents overrides", zap.Error(err))
	}

	reporter, err := source.NewStatsReporter()
	if err != nil {
		logger.Error("error building statsreporter", zap.Error(err))
	}

	ceClient, err := adapter.NewCloudEventsClientCRStatus(env.GetSink(), ceOverrides, reporter, nil)

	if err != nil {
		logger.Info("failed to create client", zap.Error(err))
		return
	}

	publisher := publisher{ceClient: ceClient, logger: logger.Desugar(), sourceServer: fmt.Sprintf("ftp://%s%s/", sftpServer, dir)}

	done := make(chan bool)

	searcher := NewWatcher(sftpServer, dir, env.User, env.Password, secure, time.Duration(probeFrequency)*time.Second, publisher.postMessage, done, configStore)
	searcher.run(ctx)
	<-stopCh
	done <- true
}

func NewEnvConfig() adapter.EnvConfigAccessor {
	return &EnvConfig{}
}
