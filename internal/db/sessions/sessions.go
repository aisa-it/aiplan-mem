package sessions

import (
	"encoding/binary"
	"time"

	"github.com/aisa-it/aiplan-mem/internal/config"
	"github.com/dgraph-io/badger/v4"
	"github.com/gofrs/uuid/v5"
)

type SessionsStore struct {
	config *config.Config
	db     *badger.DB
}

const (
	refreshTokenTTL = time.Hour * 24 * 30
)

func NewSessionsStore(db *badger.DB, cfg *config.Config) *SessionsStore {
	return &SessionsStore{db: db, config: cfg}
}

func (ss SessionsStore) BlacklistToken(signature []byte) error {
	return ss.db.Update(func(tx *badger.Txn) error {
		tm := make([]byte, 8)
		binary.LittleEndian.PutUint64(tm, uint64(time.Now().Unix()))

		e := badger.NewEntry(key(signature), tm).WithTTL(refreshTokenTTL)
		return tx.SetEntry(e)
	})
}

func (ss SessionsStore) IsTokenBlacklisted(signature []byte) (bool, error) {
	var blacklisted bool
	err := ss.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(key(signature))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if item == nil || err == badger.ErrKeyNotFound {
			return nil
		}
		return item.Value(func(timeRaw []byte) error {
			t := time.Unix(int64(binary.LittleEndian.Uint64(timeRaw)), 0)
			blacklisted = time.Now().After(t.Add(time.Second * 15)) // Freeze for all async requests
			return nil
		})
	})
	return blacklisted, err
}

func (ss SessionsStore) SaveUserLastSeenTime(userId uuid.UUID) error {
	/*return ss.db.Update(func(tx *bolt.Tx) error {
	b := tx.Bucket([]byte(config.LastSeenBucket))

	tm := make([]byte, 8)
	binary.LittleEndian.PutUint64(tm, uint64(time.Now().Unix()))

	return b.Put(userId.Bytes(), tm)
	})*/
	return nil
}

func (ss SessionsStore) GetUserLastSeenTime(userId uuid.UUID) (time.Time, error) {
	/*var lastSeen time.Time
	err := ss.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.LastSeenBucket))

		timeRaw := b.Get(userId.Bytes())
		if timeRaw == nil {
			return nil
		}
		lastSeen = time.Unix(int64(binary.LittleEndian.Uint64(timeRaw)), 0)
		return nil
	})
	return lastSeen, err*/
	return time.Time{}, nil
}

func key(key []byte) []byte {
	return []byte(config.SessionsBlaclistBucket + ":" + string(key))
}
