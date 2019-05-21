package utils

import (
	"math/rand"
	"log"
	"encoding/base64"
	"strings"
)

func RandomString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		log.Fatalln(err)
	}
	return strings.Replace(base64.URLEncoding.EncodeToString(b), "-", "_", -1)
}