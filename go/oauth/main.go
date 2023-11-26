package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pquerna/otp/totp"
	c "lambda/common"
	"time"
	"unicode"
)

func handler(ctx context.Context, r c.Req) (c.Res, error) {
	provider := r.PathParameters["provider"]

	agent, ip := r.RequestContext.Identity.UserAgent, r.RequestContext.Identity.SourceIP
	if "" == agent {
		return c.Text("User-Agent header is required", 401)
	}

	mfaCh := r.Headers["x-mfa-challenge"]
	if "" == mfaCh {
		return c.Text("X-Mfa-Challenge header is required", 401)
	}

	switch provider {
	case "google":
		return handleGoogle(ctx, r, agent, ip, mfaCh)
	default:
		return c.Text("Invalid provider", 400)
	}
}

func main() {
	lambda.Start(handler)
}

func getUser(ctx context.Context, ddb *dynamodb.Client, id string) *c.User {
	key, _ := attributevalue.MarshalMap(c.MapS{
		"PK": "U#" + id,
		"SK": "U#" + id,
	})

	resp, err := ddb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:            c.TableName,
		Key:                  key,
		ProjectionExpression: aws.String("mfa_secret, pass_hash"),
	})

	if nil != err {
		panic(err)
	}

	if nil == resp.Item {
		return nil
	}

	u := &c.User{}

	attributevalue.UnmarshalMap(resp.Item, u)

	return u
}

// saveUser registers the user in DynamoDB with a seed vault and secret
func saveUser(ctx context.Context, ddb *dynamodb.Client, u *c.User, passHash string) error {
	uItem, _ := attributevalue.MarshalMap(c.MapA{
		"PK":           "U#" + u.Email,
		"SK":           "U#" + u.Email,
		"mfa_secret":   u.MFASecret,
		"pass_hash":    passHash,
		"vault_count":  u.VaultCount,
		"secret_count": u.SecretCount,
	})

	_, err := ddb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: c.TableName,
		Item:      uItem,
	})

	return err
}

// saveSession saves the session to DynamoDB and returns the token
func saveSession(ctx context.Context, ddb *dynamodb.Client, s *c.Session) string {
	item, _ := attributevalue.MarshalMap(c.MapA{
		"PK":           "SS#" + s.Token,
		"SK":           s.UserID,
		"expiry":       s.Expiry,
		"delete_after": time.Now().Add(time.Hour * 24 * 2).Unix(),
	})

	_, err := ddb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: c.TableName,
		Item:      item,
	})

	if nil != err {
		panic(err)
	}

	return s.Token
}

func genOtp(secret, acc string) string {
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      "Butane",
		Secret:      []byte(secret),
		AccountName: acc,
	})

	return key.Secret()
}

func verifyTotp(pass, secret string) bool {
	valid, _ := totp.ValidateCustom(pass, secret, time.Now(), totp.ValidateOpts{
		Digits: 6,
	})

	return valid
}

func isValidPass(s string) bool {
	var (
		correctLength = false
		hasUpper      = false
		hasLower      = false
		hasNumber     = false
		hasSpecial    = false
	)

	if len(s) > 11 && len(s) < 73 {
		correctLength = true
	}
	for _, char := range s {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return correctLength && hasUpper && hasLower && hasNumber && hasSpecial
}
