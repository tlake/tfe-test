package backend

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type RedisBackend struct {
	address string
}

func NewRedisBackend(address string) *RedisBackend {
	return &RedisBackend{
		address: address,
	}
}

func (r *RedisBackend) connect() (redis.Conn, error) {
	conn, err := redis.Dial("tcp", r.address)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Redis: %v", err)
	}

	return conn, nil
}

func (r *RedisBackend) GetEntries() ([]string, error) {
	conn, err := r.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	entries, err := redis.Strings(conn.Do("LRANGE", "entries", 0, -1))
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (r *RedisBackend) AddEntry(entry string) error {
	conn, err := r.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Do("LPUSH", "entries", entry); err != nil {
		return err
	}

	return nil
}

func (r *RedisBackend) ClearEntries() error {
	conn, err := r.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Do("DEL", "entries"); err != nil {
		return err
	}

	return nil
}
