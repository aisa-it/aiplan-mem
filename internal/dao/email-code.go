package dao

import "time"

type EmailCodeData struct {
	NewEmail  string    `json:"new_email"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
