package common

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	const email = "john_appleseed@example.com"

	user := NewUser(email)

	if email != user.Email {
		t.Error("user Email is not set")
	}

	if "" != user.MFASecret {
		t.Error("MFA secret is not empty")
	}
}

func TestNewVault(t *testing.T) {
	const name = "test vault"
	const uID = "test user Email"

	vault := NewVault(name, uID)

	if "" == vault.ID {
		t.Error("vault id is empty")
	}

	if name != vault.DisplayName {
		t.Error("vault display name is not set")
	}

	if uID != vault.UserID {
		t.Error("vault user id is not set")
	}
}

func TestNewSecret(t *testing.T) {
	const displayname = "test secret"
	const uri = "https://example.com"
	const username = "john_appleseed"
	const password = "password"
	const vID = "test vault ID"

	secret := NewSecret(displayname, uri, username, password, vID)

	if "" == secret.DisplayName {
		t.Error("secret display name is empty")
	}

	if "" == secret.ID {
		t.Error("secret id is empty")
	}

	if uri != secret.URI {
		t.Error("secret uri is not set")
	}

	if username != secret.Username {
		t.Error("secret username is not set")
	}

	if password != secret.Password {
		t.Error("secret password is not set")
	}

	if vID != secret.VaultID {
		t.Error("secret vault id is not set")
	}
}

func TestNewMetadatum(t *testing.T) {
	const key = "test key"
	const value = "test value"

	metadatum := NewMetadatum(key, value, MetadatumTypeText)

	if key != metadatum.Key {
		t.Error("metadatum key is not set")
	}

	if value != metadatum.Value {
		t.Error("metadatum value is not set")
	}

	if MetadatumTypeText != metadatum.Type {
		t.Error("metadatum type is not set")
	}
}

func TestNewSession(t *testing.T) {
	const uID = "test user Email"

	session := NewSession(uID)

	if "" == session.Token {
		t.Error("session token is empty")
	}

	if uID != session.UserID {
		t.Error("session user ID is not set")
	}

	if 0 == session.Expiry {
		t.Error("session Expiry cannot be empty")
	}

	if time.Now().Unix() > session.Expiry {
		t.Error("session Expiry cannot be in the past")
	}
}
