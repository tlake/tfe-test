package main

import (
	"fmt"
	"github.com/quintilesims/guides/guestbook/backend"
	"github.com/quintilesims/guides/guestbook/controllers"
	"github.com/urfave/cli"
	"github.com/zpatrick/fireball"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	ENVVAR_PORT           = "GUESTBOOK_PORT"
	ENVVAR_BACKEND_TYPE   = "GUESTBOOK_BACKEND_TYPE"
	ENVVAR_BACKEND_CONFIG = "GUESTBOOK_BACKEND_CONFIG"
)

func main() {
	app := cli.NewApp()
	app.Name = "Guestbook"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:   "p, port",
			Value:  80,
			EnvVar: ENVVAR_PORT,
		},
		cli.StringFlag{
			Name:   "backend-type",
			Value:  "memory",
			EnvVar: ENVVAR_BACKEND_TYPE,
		},
		cli.StringFlag{
			Name:   "backend-config",
			EnvVar: ENVVAR_BACKEND_CONFIG,
		},
	}

	app.Action = func(c *cli.Context) error {
		var b backend.Backend

		switch backendType := strings.ToLower(c.String("backend-type")); backendType {
		case "memory":
			log.Println("Using memory backend")
			b = backend.NewMemoryBackend()
		case "redis":
			address := c.String("backend-config")
			if address == "" {
				return fmt.Errorf("Redis backend requires 'backend-config' (EnvVar: %s) to be set!", ENVVAR_BACKEND_CONFIG)
			}

			log.Println("Using redis backend")
			b = backend.NewRedisBackend(address)
		case "consul-redis":
			address := c.String("backend-config")
			if address == "" {
				return fmt.Errorf("Consul-Redis backend requires 'backend-config' (EnvVar: %s) to be set!", ENVVAR_BACKEND_CONFIG)
			}

			log.Println("Using consul-redis backend")
			b = backend.NewConsulRedisBackend(address)
		case "dynamo":
			table := c.String("backend-config")
			if table == "" {
				return fmt.Errorf("Dyamo backend requires 'backend-config' (EnvVar: %s) to be set!", ENVVAR_BACKEND_CONFIG)
			}

			log.Println("Using dynamo backend")
			b = backend.NewDynamoBackend(table)

		default:
			return fmt.Errorf("Unrecognized backend '%s'", backendType)
		}

		entryController := controllers.NewEntryController(b)
		routes := fireball.Decorate(entryController.Routes(), fireball.LogDecorator())
		server := fireball.NewApp(routes)

		addr := fmt.Sprintf(":%d", c.Int("port"))
		log.Printf("Listening on %s", addr)
		return http.ListenAndServe(addr, server)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
