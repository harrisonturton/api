package auth

import (
	"github.com/harrisonturton/api/config"
	"testing"
)

/*
Expected behaviour:
- Login creates access_token that contains refresh_token
- access_token can be used to get another access_token even if it
  is expired, so long as the contained refresh_token is invalid
- if the refresh_token is invalid, cannot refresh

cases:
- access_token is expired but refresh_token is valid
- access_token is expired and refresh_token is expired
- access_token is valid but refresh token is expired
- access_token is valid and refresh_token is valid
*/
