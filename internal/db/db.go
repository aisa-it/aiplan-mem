package db

import (
	"log/slog"
	"os"
	"time"

	"github.com/aisa-it/aiplan-mem/internal/config"
	"github.com/aisa-it/aiplan-mem/internal/db/sessions"

	"github.com/boltdb/bolt"
)

type DataStore struct {
	db *bolt.DB

	Sessions *sessions.SessionsStore
}

func OpenDB(cfg *config.Config) (*DataStore, error) {
	db, err := bolt.Open(cfg.Path, 0644, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(config.SessionsBlaclistBucket))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(config.LastSeenBucket))
		return err
	}); err != nil {
		slog.Error("Create sessions bucket", "err", err)
		os.Exit(1)
	}

	return &DataStore{db: db, Sessions: sessions.NewSessionsStore(db, cfg)}, nil
}

func (ds DataStore) Close() error {
	return ds.db.Close()
}
