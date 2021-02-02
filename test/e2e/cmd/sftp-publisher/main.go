package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type envConfig struct {
	User   string `envconfig:"FTP_USER" required:"true"`
	Count  int    `envconfig:"COUNT" default:"1"`
	Pass   string `envconfig:"FTP_PASSWORD" required:"true"`
	FTPUrl string `envconfig:"FTP_URL" required:"true"`
	Path   string `envconfig:"PATH" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Print("[ERROR] Failed to process env var: ", err)
		os.Exit(1)
	}

	sshConfig := &ssh.ClientConfig{
		User: env.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(env.Pass),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	conn, err := ssh.Dial("tcp", env.FTPUrl, sshConfig)
	if err != nil {
		log.Fatalf("Unable to dial %v", err)
	}

	sftp, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatalf("Unable to create sftp connection %v", err)
	}
	defer sftp.Close()

	for i := 0; i < env.Count; i++ {
		if _, err := sftp.Create(fmt.Sprintf("%s/test-%d", env.Path, i)); err != nil {
			log.Fatalf("Unable to create file %v", err)
		}
	}
}
