package emailcodes

import (
	"encoding/json"

	"github.com/aisa-it/aiplan-mem/apierror"
	"github.com/aisa-it/aiplan-mem/internal/dao"
	"github.com/aisa-it/aiplan-mem/internal/utils"
	"github.com/dgraph-io/badger/v4"

	"github.com/aisa-it/aiplan-mem/internal/config"
	"github.com/gofrs/uuid/v5"
)

type EmailCodesStore struct {
	db *badger.DB
}

func NewEmailCodesStore(db *badger.DB) *EmailCodesStore {
	return &EmailCodesStore{db: db}
}

func (ecs *EmailCodesStore) GenCode(userID uuid.UUID, newEmail string) (string, error) {
	data := dao.EmailCodeData{
		NewEmail: newEmail,
		Code:     utils.GenCode(),
	}
	return data.Code, ecs.db.Update(func(tx *badger.Txn) error {
		item, err := tx.Get(key(userID))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		// Code exists
		if err != badger.ErrKeyNotFound && item != nil {
			return apierror.ErrEmailCodeTooSoon
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return tx.SetEntry(badger.NewEntry(key(userID), jsonData).WithTTL(config.EmailCodeLimitReq))
	})
}

func (ecs *EmailCodesStore) VerifyCode(userID uuid.UUID, email, code string) (bool, error) {
	var verified bool
	err := ecs.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key(userID))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		return item.Value(func(val []byte) error {
			var codeData *dao.EmailCodeData
			if err := json.Unmarshal(val, &codeData); err != nil {
				return err
			}

			if codeData.NewEmail != email || codeData.Code != code {
				return apierror.ErrVerification
			}

			return txn.Delete(key(userID))
		})
	})
	return verified, err
}

func key(userID uuid.UUID) []byte {
	return []byte(config.EmailCodesBucket + ":" + userID.String())
}
