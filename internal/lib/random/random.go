package random

import (
	"math/rand"
	"time"
)

func NewRandomString(length int) string {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "abcdefghijklmnopqrstuvwxyz" + "0123456789")

	text := make([]rune, length)
	for i := range text {
		text[i] = chars[rnd.Intn(len(chars))]
	}

	return string(text)
}
