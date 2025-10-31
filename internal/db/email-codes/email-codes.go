package emailcodes

import (
	"encoding/json"
	"github.com/aisa-it/aiplan-mem/apierror"
	"github.com/aisa-it/aiplan-mem/internal/dao"
	"github.com/aisa-it/aiplan-mem/internal/utils"
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

func (ecs *EmailCodesStore) GenCode(userID uuid.UUID, newEmail string) (string, error) {
	var codeData *dao.EmailCodeData
	data := dao.EmailCodeData{
		NewEmail:  newEmail,
		Code:      utils.GenCode(),
		CreatedAt: time.Now(),
	}
	return data.Code, ecs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.EmailCodesBucket))

		jsonDataOld := b.Get(userID.Bytes())
		if jsonDataOld != nil {
			if err := json.Unmarshal(jsonDataOld, &codeData); err != nil {
				return err
			}

			if codeData.CreatedAt.Add(config.EmailCodeLimitReq).After(time.Now()) {
				return apierror.ErrEmailCodeTooSoon
			}
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return b.Put(userID.Bytes(), jsonData)
	})
}

func (ecs *EmailCodesStore) VerifyCode(userID uuid.UUID, email, code string) (bool, error) {
	var verified bool
	var codeData *dao.EmailCodeData

	err := ecs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.EmailCodesBucket))

		jsonData := b.Get(userID.Bytes())
		if jsonData == nil {
			return nil
		}

		if err := json.Unmarshal(jsonData, &codeData); err != nil {
			return err
		}

		if codeData.NewEmail != email ||
			codeData.Code != code ||
			codeData.CreatedAt.Add(config.EmailCodeLifeTime).Before(time.Now()) {
			return apierror.ErrVerification
		}

		if err := b.Delete(userID.Bytes()); err != nil {
			return err
		}

		verified = true
		return nil
	})

	if err != nil {
		return false, err
	}

	return verified, nil
}
