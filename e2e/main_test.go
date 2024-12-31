package e2e_test

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/jm96441n/movieswithfriends/e2e/internal/helpers"
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

	flag.BoolVar(&helpers.Headless, "headless", true, "run tests in headless mode")
	flag.Parse()

	m.Run()
}
