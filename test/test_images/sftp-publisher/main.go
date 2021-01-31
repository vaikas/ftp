package main

import (
	"net"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func main() {
	sshConfig := &ssh.ClientConfig{
		User: os.Getenv("USER"),
		Auth: []ssh.AuthMethod{
			ssh.Password(os.Getenv("PASS")),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	conn, err := ssh.Dial("tcp", os.Getenv("FTP_URL"), sshConfig)
	if err != nil {
		os.Exit(1)
	}

	sftp, err := sftp.NewClient(conn)
	if err != nil {
		os.Exit(1)
	}
	defer sftp.Close()

	if _, err := sftp.Create(os.Getenv("PATH")); err != nil {
		os.Exit(1)
	}
}
