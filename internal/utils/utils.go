package utils

import "github.com/sethvargo/go-password/password"

func GenCode() string {
	return password.MustGenerate(6, 6, 0, false, true)
}
