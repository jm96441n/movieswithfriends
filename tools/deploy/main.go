package main

import (
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {
	err := copyMigationFiles()
	if err != nil {
		log.Fatalf("could not copy migration files: %v", err)
	}
	err = runMigrations()
	if err != nil {
		log.Fatalf("could not run migrations: %v", err)
	}

	err = buildApplication()
	if err != nil {
		log.Fatalf("could not build application: %v", err)
	}

	err = deploy()
	if err != nil {
		log.Fatalf("could not deploy app: %v", err)
	}
}

func copyMigationFiles() error {
	key, err := os.ReadFile("/home/user/.ssh/do")
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(signer.PublicKey()),
	}

	// Connect to the remote server and perform the SSH handshake.
	client, err := ssh.Dial("tcp", "162.243.240.76:22", config)
	return nil
}

func runMigrations() error {
	return nil
}

func buildApplication() error {
	return nil
}

func deploy() error {
	return nil
}
