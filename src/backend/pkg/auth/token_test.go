package auth

import (
	"testing"
	"time"
)

func TestTokenGenerateValidate(t *testing.T) {

	var err error

	token, err := generateSessionToken(options.AccessKey, 15*60)
	if err != nil {
		t.Error("X Failed ")
	} else {
		t.Logf("→ OK ▶ %s", "generateSessionToken")
	}

	if expired, err := validateSessionToken(options.AccessKey, token); err != nil {
		t.Error("X Failed ")
	} else {
		if expired {
			t.Error("X Failed to validate token (is expired)")
		} else {
			t.Logf("→ OK ▶ %s", "validateSessionToken")
		}
	}

	claims, err := getTokenClaims(token)
	if err != nil {
		t.Error("X Failed ")
	} else {
		t.Logf("→ OK ▶ %s", claims)
	}

}

func TestTokenExpired(t *testing.T) {

	var err error

	token, err := generateSessionToken(options.AccessKey, 1) //set expire  1 second
	if err != nil {
		t.Error("X Failed ")
	} else {
		t.Logf("→ OK ▶ %s", token)
	}

	time.Sleep(2 * time.Second) //delay 2 second
	if expired, err := validateSessionToken(options.AccessKey, token); err != nil {
		t.Error("X Failed ")
	} else {
		if expired {
			t.Logf("→ OK ▶ %s", "expired")
		} else {
			t.Error("X Failed (is not expired)")
		}
	}

}
