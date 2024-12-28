package e2e_test

import (
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	cmd := exec.Command("docker", "build", "--target=prod", "-t", "movieswithfriends:test", ".")
	cmd.Dir = ".."

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		log.Fatalf("could not build docker image: %v", err)
	}

	m.Run()
}
