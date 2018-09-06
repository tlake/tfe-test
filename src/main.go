package main

import (
	"fmt"
	"net/http"
	"os"
)

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return fallback
}

func main() {
	var deploymentEnvironment = getEnv("DEPLOYMENT_ENVIRONMENT", "unset/dev")
	var servicePort = getEnv("SERVICE_PORT", "8080")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world!\nThe URI you requested was '%s'\n\n", r.URL.Path)
		fmt.Fprintf(w, "Service port: '%s'\nDeployment environment: '%s'\n\n", servicePort, deploymentEnvironment)
		fmt.Fprintf(w, "Some staging work")
	})

	http.ListenAndServe(":"+servicePort, nil)
}
