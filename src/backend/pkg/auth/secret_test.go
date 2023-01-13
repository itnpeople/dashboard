package auth

import (
	"testing"
)

// https://github.com/kore3lab/dashboard/blob/master/docs/user/config-sign-in.md
func TestSecretProvider(t *testing.T) {

	options.Secret = SecretStaticToken
	if provider, err := getStaticTokenSecretProvider(options); err != nil {
		t.Error(err)
	} else {
		if provider(options.Token, "") == options.Token {
			t.Logf("→ OK ▶ %s", options.Secret)
		} else {
			t.Error("X Failed ")
		}
	}

	options.Secret = SecretStaticUser
	if provider, err := getStaticUserSecretProvider(options); err != nil {
		t.Error(err)
	} else {
		if provider(options.Username, "") == options.Password {
			t.Logf("→ OK ▶ %s", options.Secret)
		} else {
			t.Error("X Failed ")
		}
	}

	createBasicUserFile(t)
	options.Secret = SecretBasicUser
	if provider, err := getBasicUserSecretProvider(options); err != nil {
		t.Error(err)
	} else {
		if provider(options.Username, "") == options.Password {
			t.Logf("→ OK ▶ %s", options.Secret)
		} else {
			t.Error("X Failed ")
		}
	}

	createOpaqueFile(t)
	options.Secret = SecretOpaque
	if provider, err := getOpaqueSecretProvider(options); err != nil {
		t.Error(err)
	} else {
		if provider(options.Username, "") == options.Password {
			t.Logf("→ OK ▶ %s", options.Secret)
		} else {
			t.Error("X Failed ")
		}
	}

}
