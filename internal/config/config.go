package config

import (
	"os"
	"time"
)

const (
	SessionsBlaclistBucket = "sessionsBlacklist"
	LastSeenBucket         = "userLastSeen"
	EmailCodesBucket       = "emailCodes"

	EmailCodeLifeTime time.Duration = time.Minute * 5
	EmailCodeLimitReq               = time.Minute
)

type Config struct {
	Path       string
	ListenAddr string
}

func FromENV() *Config {
	c := &Config{
		Path:       os.Getenv("AIPLAN_MEM_PATH"),
		ListenAddr: os.Getenv("AIPLAN_MEM_LISTEN_ADDR"),
	}

	if c.Path == "" {
		c.Path = "aiplan_mem.db"
	}

	if c.ListenAddr == "" {
		c.ListenAddr = ":8080"
	}

	return c
}
