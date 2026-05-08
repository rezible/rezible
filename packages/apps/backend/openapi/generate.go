//go:build ignore

package main

import (
	"log"
	"os"

	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

func main() {
	spec, specErr := oapiv1.GetSpec(false)
	if specErr != nil {
		log.Fatalf("failed to get spec: %v", specErr)
	}
	if writeErr := os.WriteFile("./v1/openapi.yaml", []byte(spec), 0644); writeErr != nil {
		log.Fatalf("failed to write spec: %v", writeErr)
	}
}
