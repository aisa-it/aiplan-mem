package dao

type EmailCodeData struct {
	NewEmail string `json:"new_email"`
	Code     string `json:"code"`
}
