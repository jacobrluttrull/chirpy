package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetApiKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no Authorization header")
	}
	split := strings.Split(authHeader, " ")
	if len(split) != 2 {
		return "", fmt.Errorf("invalid Authorization header")
	}
	return split[1], nil
}
