package sessions

import (
	"encoding/binary"
	"time"

	"github.com/aisa-it/aiplan-mem/internal/config"
	"github.com/gofrs/uuid/v5"

	"github.com/boltdb/bolt"
)

type SessionsStore struct {
	config *config.Config
	db     *bolt.DB
}

func NewSessionsStore(db *bolt.DB, cfg *config.Config) *SessionsStore {
	return &SessionsStore{db: db, config: cfg}
}

func (ss SessionsStore) BlacklistToken(signature []byte) error {
	return ss.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.SessionsBlaclistBucket))

		tm := make([]byte, 8)
		binary.LittleEndian.PutUint64(tm, uint64(time.Now().Unix()))

		return b.Put(signature, tm)
	})
}

func (ss SessionsStore) IsTokenBlacklisted(signature []byte) (bool, error) {
	var blacklisted bool
	err := ss.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.SessionsBlaclistBucket))

		timeRaw := b.Get(signature)
		if timeRaw == nil {
			return nil
		}
		t := time.Unix(int64(binary.LittleEndian.Uint64(timeRaw)), 0)
		blacklisted = t.After(t.Add(time.Minute)) // Freeze for 1 minute for all async requests
		return nil
	})
	return blacklisted, err
}

func (ss SessionsStore) SaveUserLastSeenTime(userId uuid.UUID) error {
	return ss.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.LastSeenBucket))

		tm := make([]byte, 8)
		binary.LittleEndian.PutUint64(tm, uint64(time.Now().Unix()))

		return b.Put(userId.Bytes(), tm)
	})
}

func (ss SessionsStore) GetUserLastSeenTime(userId uuid.UUID) (time.Time, error) {
	var lastSeen time.Time
	err := ss.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.LastSeenBucket))

		timeRaw := b.Get(userId.Bytes())
		if timeRaw == nil {
			return nil
		}
		lastSeen = time.Unix(int64(binary.LittleEndian.Uint64(timeRaw)), 0)
		return nil
	})
	return lastSeen, err
}
