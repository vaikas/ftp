package main

import (
	"context"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"

	"go.uber.org/zap"

	"knative.dev/pkg/logging"

	"github.com/pkg/sftp"
	"github.com/secsy/goftp"
)

type watcher struct {
	server    string
	dir       string
	user      string
	password  string
	secure    bool          // use SFTP if true
	frequency time.Duration // in seconds
	handler   func(context.Context, os.FileInfo) error
	stop      chan bool
	fetcher   func(context.Context)

	store *configstore
	// Our copy of the data so we don't have to load the configmap each time, it
	// shouldn't change underneath us.
	data configdata
}

type configdata struct {
	LastFileProcessed string
	LastModTime       time.Time
}

func NewWatcher(server string, dir string, user string, password string, secure bool, frequency time.Duration, handler func(context.Context, os.FileInfo) error, stop chan bool, store *configstore) *watcher {
	return &watcher{server: server, dir: dir, user: user, password: password, secure: secure, frequency: frequency, handler: handler, stop: stop, store: store}
}

func (s *watcher) run(ctx context.Context) {
	logger := logging.FromContext(ctx)
	// Initialize the config store that we'll be using to store our state.
	s.store.Init(ctx, &configdata{})

	if s.secure {
		logger.Info("Using SFTP to fetch files")
		s.fetcher = s.fetchSFTP
	} else {
		logger.Info("Using FTP to fetch files")
		s.fetcher = s.fetchFTP
	}
	tickChan := time.NewTicker(s.frequency).C
	go func() {
		for {
			select {
			case <-tickChan:
				s.fetcher(ctx)
			case <-s.stop:
				logger.Info("Exiting")
				return
			}
		}

	}()
}

func (s *watcher) processFiles(ctx context.Context, entries []os.FileInfo) {
	logger := logging.FromContext(ctx)
	var data configdata
	err := s.store.Load(ctx, &data)
	if err != nil {
		logger.Error("Failed to load configmap:", zap.Error(err))
		return
	}
	logger.Info("Loaded configdata:", zap.Any("data", data))

	// This is super simple way to figure out what we've seen so far, we look
	// at the ModTime and upon finishing the batch of files, we'll stash the
	// highest we've seen.
	// Better way would be to keep list of files we've seen (so that you could
	// send deleted events, etc.). For now, we'll just stash the last file seen
	// and the latest modtime. Not great, but will suffice for now.
	highwaterMark := data.LastModTime
	highwaterFile := data.LastFileProcessed

	for _, e := range entries {
		logger.Info("Got file:", zap.String("file", e.Name()), zap.Time("time", e.ModTime()))
		if e.ModTime().Before(data.LastModTime) {
			continue
		}
		if e.Name() == highwaterFile {
			continue
		}
		logger.Info("Found new file:", zap.String("file", e.Name()))
		handlerErr := s.handler(ctx, e)
		if handlerErr != nil {
			logger.Error("Failed to post:", zap.Error(err))
			break
		}
		if e.ModTime().After(highwaterMark) {
			logger.Info("Found new file:", zap.String("highwaterFile", highwaterFile), zap.Time("highwaterMark", highwaterMark))
			highwaterFile = e.Name()
			highwaterMark = e.ModTime()
		}
	}
	if highwaterMark.After(data.LastModTime) {
		logger.Info("Saving CONFIGMAP")
		newConfigData := configdata{LastFileProcessed: highwaterFile, LastModTime: highwaterMark}
		err = s.store.Save(ctx, &newConfigData)
		if err != nil {
			logger.Error("Failed to save the configdata:", zap.Error(err))
			return
		}
		s.data = newConfigData
	}

}

func (s *watcher) fetchFTP(ctx context.Context) {
	logger := logging.FromContext(ctx)
	config := goftp.Config{
		User:        s.user,
		Password:    s.password,
		DisableEPSV: true,
	}

	logger.Info("making plain connection to", zap.String("server", s.server))
	client, err := goftp.DialConfig(config, s.server)
	if err != nil {
		logger.Error("Failed to dial:", zap.Error(err))
		return
	}
	defer client.Close()
	entries, err := client.ReadDir(s.dir)
	if err != nil {
		logger.Error("Failed to ReadDir:", zap.Error(err))
		return
	}

	s.processFiles(ctx, entries)
}

func (s *watcher) fetchSFTP(ctx context.Context) {
	logger := logging.FromContext(ctx)
	sshConfig := &ssh.ClientConfig{
		User: s.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	logger.Info("making ssh connection to", zap.String("server", s.server))
	conn, err := ssh.Dial("tcp", s.server, sshConfig)
	if err != nil {
		logger.Error("Failed to ssh dial:", zap.Error(err))
		return
	}

	sftp, err := sftp.NewClient(conn)
	if err != nil {
		logger.Error("Failed to create new sftp client:", zap.Error(err))
		return
	}
	defer sftp.Close()

	entries, err := sftp.ReadDir(s.dir)
	if err != nil {
		logger.Error("Failed to ReadDir:", zap.Error(err))
		return
	}

	s.processFiles(ctx, entries)
}
