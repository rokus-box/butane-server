package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	c "lambda/common"
	"log"
	"strings"
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

	body := strings.SplitN(r.Body, "\n", 2)

	if len(body) != 2 {
		return c.Text("Malformed request body", 400)
	}

	oauthCode := body[0]
	plainPass := body[1]

	if !isValidPass(plainPass) {
		return c.Text("Password must match ^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]{12,73}$ pattern", 400)
	}

	switch provider {
	case "google":
		return handleGoogle(ctx, agent, ip, mfaCh, oauthCode, plainPass)
	case "discord":
		return handleDiscord(ctx, agent, ip, mfaCh, oauthCode, plainPass)
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
func saveUser(ctx context.Context, u *c.User, passHash string) error {
	uItem, _ := attributevalue.MarshalMap(c.MapA{
		"PK":           "U#" + u.Email,
		"SK":           "U#" + u.Email,
		"mfa_secret":   u.MFASecret,
		"pass_hash":    passHash,
		"vault_count":  u.VaultCount,
		"secret_count": u.SecretCount,
	})

	_, err := ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: c.TableName,
		Item:      uItem,
	})

	return err
}

// saveSession saves the session to DynamoDB and returns the token
func saveSession(ctx context.Context, s *c.Session) string {
	item, _ := attributevalue.MarshalMap(c.MapA{
		"PK":           "SS#" + s.Token,
		"SK":           s.UserID,
		"expiry":       s.Expiry,
		"delete_after": time.Now().Add(time.Hour * 24 * 2).Unix(),
	})

	_, err := ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
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

func handleNewUser(ctx context.Context, oauthCode, mfaCh, email, plainPass string) (c.Res, error) {
	otpSecret := genOtp(oauthCode[len(oauthCode)-10:], "#")
	if !verifyTotp(mfaCh, otpSecret) {
		return c.Status(401)
	}

	newUser := c.NewUser(email)
	newUser.MFASecret = otpSecret
	passHash, err := bcrypt.GenerateFromPassword([]byte(plainPass), bcrypt.DefaultCost)

	if nil != err {
		log.Println("failed to hash password: ", err)
		return c.Text("Failed to register user. Please try again later", 503)
	}

	if nil != saveUser(ctx, newUser, string(passHash)) {
		return c.Text("Failed to register user. Please try again later", 503)
	}

	sess := c.NewSession(email)
	return c.Text(saveSession(ctx, sess), 201)
}

func handleExistingUser(ctx context.Context, passHash, plainPass, email, agent, ip, provider string) (c.Res, error) {
	err := bcrypt.CompareHashAndPassword([]byte(passHash), []byte(plainPass))
	if nil != err {
		al := c.NewAuditLog(email, "Failed attempt to login with "+provider, c.ResourceSession, c.ActionCreate, c.MapS{"agent": agent, "ip": ip})

		item, _ := attributevalue.MarshalMap(c.MapA{
			"PK":           "U#" + al.UserID,
			"SK":           "AL#" + al.Timestamp.Format(time.RFC3339Nano),
			"action":       al.Action,
			"resource":     al.Resource,
			"message":      al.Message,
			"data":         al.Data,
			"delete_after": time.Now().Add(time.Hour * 24 * 20).Unix(),
		})

		_, err := ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: c.TableName,
			Item:      item,
		})

		if nil != err {
			panic(err)
		}

		return c.Status(401)
	}

	sess := c.NewSession(email)
	return c.Text(saveSession(ctx, sess), 201)
}
