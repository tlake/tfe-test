package backend

import (
	"fmt"
	"github.com/hashicorp/consul/api"
)

type ConsulRedisBackend struct {
	redisServiceName string
}

func NewConsulRedisBackend(redisServiceName string) *ConsulRedisBackend {
	return &ConsulRedisBackend{
		redisServiceName: redisServiceName,
	}
}

func (c *ConsulRedisBackend) getRedisAddress() (string, error) {
	config := api.DefaultConfig()
	config.Address = "consul-agent:8500"
	client, err := api.NewClient(config)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to consul: %v", err)
	}

	services, _, err := client.Catalog().Service(c.redisServiceName, "", nil)
	if err != nil {
		return "", fmt.Errorf("Failed to list consul services: %v", err)
	}

	if len(services) == 0 {
		return "", fmt.Errorf("Service '%s' was not registered in consul", c.redisServiceName)
	}

	return fmt.Sprintf("%s:%d", services[0].ServiceAddress, services[0].ServicePort), nil
}

func (c *ConsulRedisBackend) getRedisBackend() (*RedisBackend, error) {
	address, err := c.getRedisAddress()
	if err != nil {
		return nil, err
	}

	return NewRedisBackend(address), nil
}

func (c *ConsulRedisBackend) GetEntries() ([]string, error) {
	redis, err := c.getRedisBackend()
	if err != nil {
		return nil, err
	}

	return redis.GetEntries()
}

func (c *ConsulRedisBackend) AddEntry(message string) error {
	redis, err := c.getRedisBackend()
	if err != nil {
		return err
	}

	return redis.AddEntry(message)
}

func (c *ConsulRedisBackend) ClearEntries() error {
	redis, err := c.getRedisBackend()
	if err != nil {
		return err
	}

	return redis.ClearEntries()
}
