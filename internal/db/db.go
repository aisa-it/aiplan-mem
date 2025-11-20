package db

import (
	emailcodes "github.com/aisa-it/aiplan-mem/internal/db/email-codes"
	"github.com/dgraph-io/badger/v4"

	"github.com/aisa-it/aiplan-mem/internal/config"
	"github.com/aisa-it/aiplan-mem/internal/db/sessions"
)

type DataStore struct {
	db *badger.DB

	Sessions   *sessions.SessionsStore
	EmailCodes *emailcodes.EmailCodesStore
}

func OpenDB(cfg *config.Config) (*DataStore, error) {
	db, err := badger.Open(badger.DefaultOptions(cfg.Path).
		//WithInMemory(true).
		WithNumVersionsToKeep(0).
		WithValueThreshold(1024).
		WithNumLevelZeroTables(10),
	)
	if err != nil {
		return nil, err
	}

	return &DataStore{db: db,
		Sessions:   sessions.NewSessionsStore(db, cfg),
		EmailCodes: emailcodes.NewEmailCodesStore(db)}, nil
}

func (ds DataStore) Close() error {
	return ds.db.Close()
}
