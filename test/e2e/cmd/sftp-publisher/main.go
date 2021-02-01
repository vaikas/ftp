package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func main() {
	fileCount, _ := strconv.Atoi(os.Getenv("COUNT"))

	sshConfig := &ssh.ClientConfig{
		User: os.Getenv("FTP_USER"),
		Auth: []ssh.AuthMethod{
			ssh.Password(os.Getenv("FTP_PASSWORD")),
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

	for i := 0; i < fileCount; i++ {
		if _, err := sftp.Create(fmt.Sprintf("test-%d", i)); err != nil {
			log.Fatalf("Unable to create file %v", err)
		}
	}
}
