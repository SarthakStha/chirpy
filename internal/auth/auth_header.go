package auth

import (
	"fmt"
	"net/http"
	"strings"
)

// Typical Header structure Authorization: <scheme> <credentials>
// Assuming only one white space between scheme and credentials
func GetSchemeToken(header http.Header, scheme string) (string, error) {
	authHeader := header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("unable to get Token, 'Authorization' header not set")
	}

	tokenPrefix := fmt.Sprintf("%s ", scheme)
	token, found := strings.CutPrefix(authHeader, tokenPrefix)
	if !found {
		return "", fmt.Errorf("invalid 'Authorization' header, missing/misformatted scheme prefix")
	}
	return token, nil
}

// Get the credentials from the Authentication header for the
// scheme 'Bearer'
func GetBearerToken(header http.Header) (string, error) {
	scheme := "Bearer"
	return GetSchemeToken(header, scheme)
}

// Get the credentials from the Authentication header for the
// scheme 'ApiKey'
func GetAPIKey(header http.Header) (string, error) {
	scheme := "ApiKey"
	return GetSchemeToken(header, scheme)
}
