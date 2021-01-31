package main

import (
	"log"
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
		log.Fatalf("Unable to dial %v", err)
	}

	sftp, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatalf("Unable to create sftp connection %v", err)
	}
	defer sftp.Close()

	if _, err := sftp.Create(os.Getenv("PATH")); err != nil {
		log.Fatalf("Unable to create file %s, %v", os.Getenv("PATH"), err)
	}
}
