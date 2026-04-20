package main

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var openapiSpec []byte

func openapiHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(openapiSpec)
}
