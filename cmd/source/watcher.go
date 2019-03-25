package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/secsy/goftp"
)

type watcher struct {
	server    string
	dir       string
	user      string
	password  string
	secure    bool  // use SFTP if true
	frequency int64 // in seconds
	handler   func(os.FileInfo) error
	stop      chan bool
	fetcher   func()

	store *configstore
	// Our copy of the data so we don't have to load the configmap each time, it
	// shouldn't change underneath us.
	data configdata
}

type configdata struct {
	LastFileProcessed string
	LastModTime       time.Time
}

func NewWatcher(server string, dir string, user string, password string, secure bool, frequency int64, handler func(os.FileInfo) error, stop chan bool, store *configstore) *watcher {
	return &watcher{server: server, dir: dir, user: user, password: password, secure: secure, frequency: frequency, handler: handler, stop: stop, store: store}
}

func (s *watcher) run() {
	// Initialize the config store that we'll be using to store our state.
	s.store.Init(&configdata{})

	if s.secure {
		fmt.Println("Using SFTP to fetch files")
		s.fetcher = s.fetchSFTP
	} else {
		fmt.Println("Using FTP to fetch files")
		s.fetcher = s.fetchFTP
	}
	tickChan := time.NewTicker(10 * time.Second).C
	go func() {
		for {
			select {
			case <-tickChan:
				s.fetcher()
			case <-s.stop:
				fmt.Println("Exiting")
				return
			}
		}

	}()
}

func (s *watcher) processFiles(entries []os.FileInfo) {
	var data configdata
	err := s.store.Load(&data)
	if err != nil {
		fmt.Printf("Failed to load configmap: %s\n", err)
		return
	}
	fmt.Printf("Loaded configdata: %+v\n", data)

	// This is super simple way to figure out what we've seen so far, we look
	// at the ModTime and upon finishing the batch of files, we'll stash the
	// highest we've seen.
	// Better way would be to keep list of files we've seen (so that you could
	// send deleted events, etc.). For now, we'll just stash the last file seen
	// and the latest modtime. Not great, but will suffice for now.
	highwaterMark := data.LastModTime
	highwaterFile := data.LastFileProcessed

	for _, e := range entries {
		fmt.Printf("Got file: %s - %s\n", e.Name(), e.ModTime())
		if e.ModTime().Before(data.LastModTime) {
			continue
		}
		if e.Name() == highwaterFile {
			continue
		}
		fmt.Printf("Found new file: %s\n", e.Name())
		handlerErr := s.handler(e)
		if handlerErr != nil {
			fmt.Printf("Failed to post: %s", err)
			break
		}
		if e.ModTime().After(highwaterMark) {
			fmt.Printf("Setting HighwaterMark: %s %s", highwaterFile, highwaterMark)
			highwaterFile = e.Name()
			highwaterMark = e.ModTime()
		}
	}
	if highwaterMark.After(data.LastModTime) {
		fmt.Printf("Saving CONFIGMAP\n")
		newConfigData := configdata{LastFileProcessed: highwaterFile, LastModTime: highwaterMark}
		err = s.store.Save(&newConfigData)
		if err != nil {
			fmt.Printf("Failed to save the configdata: %s", err)
			return
		}
		s.data = newConfigData
	}

}

func (s *watcher) fetchFTP() {
	config := goftp.Config{
		User:        s.user,
		Password:    s.password,
		DisableEPSV: true,
	}

	fmt.Printf("making plain connection to %q\n", s.server)
	client, err := goftp.DialConfig(config, s.server)
	if err != nil {
		fmt.Printf("Failed to dial: %s", err)
		return
	}
	defer client.Close()

	entries, err := client.ReadDir(s.dir)
	if err != nil {
		fmt.Printf("Failed to ReadDir: %s", err)
		return
	}

	s.processFiles(entries)
}

func (s *watcher) fetchSFTP() {
	sshConfig := &ssh.ClientConfig{
		User: s.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	fmt.Printf("making ssh connection to %q\n", s.server)
	conn, err := ssh.Dial("tcp", s.server, sshConfig)
	if err != nil {
		fmt.Printf("Failed to ssh dial: %s", err)
		return
	}

	sftp, err := sftp.NewClient(conn)
	if err != nil {
		fmt.Printf("Failed to create new sftp client: %s", err)
		return
	}
	defer sftp.Close()

	entries, err := sftp.ReadDir(s.dir)
	if err != nil {
		fmt.Printf("Failed to ReadDir: %s", err)
		return
	}

	s.processFiles(entries)
}
