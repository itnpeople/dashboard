package auth

import (
	"fmt"
	"os/exec"
	"testing"
)

var options = &AuthenticatorOptions{
	Strategy:   StrategyCookie,
	AccessKey:  "whdmstkddk",
	RefreshKey: "hsthvmxm",
	Username:   "admin",
	Password:   "t0p-secret",
	Token:      "kore3lab",
	Location:   "/var/tmp",
}
var oUser = map[string]string{"username": options.Username, "password": options.Password}
var oInvalidUser = map[string]string{"username": options.Username, "password": options.Password + "_modified"}
var oToken = map[string]string{"token": options.Token}
var oInvalidToken = map[string]string{"token": options.Token + "_modified"}

// https://github.com/kore3lab/dashboard/blob/master/docs/user/config-sign-in.md
func TestNewAuthenticator(t *testing.T) {

	opts := fmt.Sprintf("strategy=cookie,secret=static-token,token=%s", options.Token)

	if authenticator, err := NewAuthenticator(opts); err != nil {
		t.Error(err)
	} else {
		if err := authenticator.Validate(oToken); err != nil {
			t.Errorf("X Failed (options=%s, cause=%v)", opts, err)
		} else {
			if err := authenticator.Validate(oInvalidToken); err != nil {
				t.Logf("→ OK ▶ %s", opts)
			} else {
				t.Errorf("X Failed (options=%s, cause=%v)", opts, err)
			}
		}
	}

	opts = fmt.Sprintf("strategy=cookie,secret=static-user,username=%s,password=%s", options.Username, options.Password)

	if authenticator, err := NewAuthenticator(opts); err != nil {
		t.Error(err)
	} else {
		if err := authenticator.Validate(oUser); err != nil {
			t.Errorf("X Failed (options=%s, cause=%v)", opts, err)
		} else {
			if err := authenticator.Validate(oInvalidUser); err != nil {
				t.Logf("→ OK ▶ %s", opts)
			} else {
				t.Errorf("X Failed (options=%s, cause=%v)", opts, err)
			}
		}
	}

	opts = fmt.Sprintf("strategy=cookie,secret=basic-user,location=%s", options.Location)

	createBasicUserFile(t)
	if authenticator, err := NewAuthenticator(opts); err != nil {
		t.Error(err)
	} else {
		if err := authenticator.Validate(oUser); err != nil {
			t.Errorf("X Failed (options=%s, cause=%v)", opts, err)
		} else {
			if err := authenticator.Validate(oInvalidUser); err != nil {
				t.Logf("→ OK ▶ %s", opts)
			} else {
				t.Errorf("X Failed (options=%s, cause=%v)", opts, err)
			}
		}
	}

	opts = fmt.Sprintf("strategy=cookie,secret=opaque,location=%s", options.Location)

	createOpaqueFile(t)
	if authenticator, err := NewAuthenticator(opts); err != nil {
		t.Error(err)
	} else {
		if err := authenticator.Validate(oUser); err != nil {
			t.Errorf("X Failed (options=%s, cause=%v)", opts, err)
		} else {
			if err := authenticator.Validate(oInvalidUser); err != nil {
				t.Logf("→ OK ▶ %s", opts)
			} else {
				t.Errorf("X Failed (options=%s, cause=%v)", opts, err)
			}
		}
	}

}

func createBasicUserFile(t *testing.T) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo -n '%s' > %s/username", options.Username, options.Location))
	if err := cmd.Run(); err != nil {
		t.Errorf("X Failed to create a username file (cause=%v)", err)
	} else {
		cmd := exec.Command("bash", "-c", fmt.Sprintf("echo -n '%s' > %s/password", options.Password, options.Location))
		if err := cmd.Run(); err != nil {
			t.Errorf("X Failed to create a password file (cause=%v)", err)
		}
	}

}
func createOpaqueFile(t *testing.T) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo -n '%s' | base64 > %s/%s", options.Password, options.Location, options.Username))
	if err := cmd.Run(); err != nil {
		t.Errorf("X Failed to create a qpaque file (cause=%v)", err)
	}
}
