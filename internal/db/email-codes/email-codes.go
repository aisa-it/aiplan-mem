package emailcodes

import (
	"encoding/json"
	"github.com/aisa-it/aiplan-mem/internal/dao"
	"time"

	"github.com/aisa-it/aiplan-mem/internal/config"
	"github.com/boltdb/bolt"
	"github.com/gofrs/uuid/v5"
)

type EmailCodesStore struct {
	db *bolt.DB
}

func NewEmailCodesStore(db *bolt.DB) *EmailCodesStore {
	return &EmailCodesStore{db: db}
}

func (ecs *EmailCodesStore) SaveCode(userID uuid.UUID, newEmail, code string, expiresIn time.Duration) error {
	return ecs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.EmailCodesBucket))

		data := dao.EmailCodeData{
			NewEmail:  newEmail,
			Code:      code,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(expiresIn),
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return b.Put(userID.Bytes(), jsonData)
	})
}

func (ecs *EmailCodesStore) GetCode(userID uuid.UUID) (*dao.EmailCodeData, error) {
	var codeData *dao.EmailCodeData
	err := ecs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.EmailCodesBucket))

		jsonData := b.Get(userID.Bytes())
		if jsonData == nil {
			return nil
		}

		return json.Unmarshal(jsonData, &codeData)
	})
	return codeData, err
}
