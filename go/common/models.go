package common

import (
	"time"
)

type (
	User struct {
		Email       string `json:"-" dynamodbav:"PK"`
		MFASecret   string `json:"-" dynamodbav:"mfa_secret"`
		PassHash    string `json:"-" dynamodbav:"pass_hash"`
		VaultCount  int    `json:"-" dynamodbav:"vault_count"`
		SecretCount int    `json:"-" dynamodbav:"secret_count"`
	}

	Session struct {
		Expiry int64  `json:"-" dynamodbav:"expiry"`
		Token  string `json:"-" dynamodbav:"PK"`
		UserID string `json:"-" dynamodbav:"SK"`
	}

	Vault struct {
		ID          string `json:"id" dynamodbav:"SK"`
		DisplayName string `json:"display_name" dynamodbav:"display_name"`
		UserID      string `json:"-"`
	}

	Secret struct {
		ID          string       `json:"id" dynamodbav:"Sk"`
		DisplayName string       `json:"display_name" dynamodbav:"display_name"`
		URI         string       `json:"uri,omitempty" dynamodbav:"uri"`
		Username    string       `json:"username,omitempty" dynamodbav:"username"`
		Password    string       `json:"password,omitempty" dynamodbav:"password"`
		Metadata    []*Metadatum `json:"metadata,omitempty" dynamodbav:"metadata"`
		VaultID     string       `json:"-"`
	}

	Metadatum struct {
		Key   string `json:"key" dynamodbav:"key"`
		Value string `json:"value" dynamodbav:"value"`
		Type  uint8  `json:"type" dynamodbav:"type"`
	}

	AuditLog struct {
		Data      any       `json:"data,omitempty" dynamodbav:"data"`
		Action    uint8     `json:"action" dynamodbav:"action"`
		Resource  uint8     `json:"resource" dynamodbav:"resource"`
		Message   string    `json:"message" dynamodbav:"message"`
		Timestamp time.Time `json:"timestamp" dynamodbav:"SK"`
		UserID    string    `json:"-"`
	}
)

const (
	MetadatumTypeText uint8 = iota
	MetadatumTypeMFA
	MetadatumTypeConfidential
)

const (
	ActionCreate uint8 = iota
	ActionUpdate
	ActionDelete
)

const (
	ResourceSession uint8 = iota
	ResourceVault
	ResourceSecret
	ResourceMetadatum
)

// NewUser creates a new user with a sample vault, secret and metadata
func NewUser(email string) *User {
	return &User{
		Email: email,
	}
}

// NewVault creates a new vault
func NewVault(name, uID string) *Vault {
	return &Vault{
		ID:          NewID(),
		DisplayName: name,
		UserID:      uID,
	}
}

// NewSecret creates a new secret
func NewSecret(name, uri, username, password, vID string) *Secret {
	return &Secret{
		ID:          NewID(),
		DisplayName: name,
		URI:         uri,
		Username:    username,
		Password:    password,
		VaultID:     vID,
	}
}

// NewMetadatum creates a new metadatum
func NewMetadatum(key, value string, t uint8) *Metadatum {
	return &Metadatum{
		Key:   key,
		Value: value,
		Type:  t,
	}
}

// NewSession creates a new session
func NewSession(uID string) *Session {
	return &Session{
		Token:  HashSHA256Base64(NewID()),
		Expiry: time.Now().Add(1 * time.Minute).Unix(),
		UserID: uID,
	}
}

// Extend updates the expiry of the session
func (ss *Session) Extend() {
	ss.Expiry = time.Now().Add(1 * time.Minute).Unix()
}
