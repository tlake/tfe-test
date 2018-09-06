package main

import (
	"fmt"
	"net/http"
	"os"
)

var deploymentEnvironment = getEnv("DEPLOYMENT_ENVIRONMENT", "unset/dev")

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return fallback
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world!\nThe URI you requested was '%s'\n\n", r.URL.Path)
		fmt.Fprintf(w, "Deployment environment: '%s'\n\n", deploymentEnvironment)
		fmt.Fprintf(w, "This is some code.")
	})

	http.ListenAndServe(":8080", nil)
}
