package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

const length = 32

func main() {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		log.Fatalf("could not generate secure key: %v", err)
		return
	}

	sessionStore := sessions.NewCookieStore([]byte(k))
	req := &http.Request{}
	_, err := sessionStore.New(req, "name")
	if err != nil {
		log.Fatalf("could not create session: %v", err)
	}
	fmt.Printf("export SESSION_KEY=%q\n", base64.URLEncoding.EncodeToString(k))
}
