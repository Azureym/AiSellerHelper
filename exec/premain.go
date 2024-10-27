package main

import (
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
)

var (
	authorization string
)

// auth=***** ./script.sh
func init() {
	auth, exists := os.LookupEnv("auth")
	if !exists {
		log.Fatalf("can not get auth key.")
	}
	log.Println(spew.Sprintf("get auth from env. auth=%s", auth))
	authorization = auth
}
