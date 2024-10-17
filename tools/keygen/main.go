package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
)

const length = 32

func main() {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		log.Fatalf("could not generate secure key: %v", err)
		return
	}
	fmt.Printf("export SESSION_KEY=%q\n", base64.URLEncoding.EncodeToString(k))
}
