package main

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	c "lambda/common"
	"log"
	"os"
	"time"
)

var googleClient = resty.New().SetBaseURL("https://oauth2.googleapis.com").SetTimeout(5 * time.Second)
var ddbClient = c.NewDDB()

type googleResp struct {
	Token string `json:"id_token"`
}

type googleClaims struct {
	jwt.Claims
	Email    string `json:"email"`
	Verified bool   `json:"email_verified"`
}

func handleGoogle(ctx context.Context, agent, ip, mfaCh, oauthCode, plainPass string) (c.Res, error) {
	respBody := &googleResp{}

	resp, err := googleClient.R().SetFormData(c.MapS{
		"code":          oauthCode,
		"client_id":     os.Getenv("GOOGLE_CLIENT_ID"),
		"client_secret": os.Getenv("GOOGLE_CLIENT_SECRET"),
		"redirect_uri":  os.Getenv("GOOGLE_REDIRECT_URI"),
		"grant_type":    "authorization_code",
	}).SetResult(respBody).Post("/token")
	if os.IsTimeout(err) {
		return c.Text("Request to provider timed out. Please try again later", 503)
	}

	if nil != err {
		log.Println("failed to authenticate with Google: ", err)
		return c.Text("Failed to authenticate. Please try again later", 503)
	}

	if resp.IsError() {
		return c.Text("Invalid authentication details were provided", 401)
	}

	claims := &googleClaims{}
	jwt.ParseWithClaims(respBody.Token, claims, nil)

	if !claims.Verified {
		return c.Text("Email address is not verified", 401)
	}

	user := getUser(ctx, ddbClient, claims.Email)

	if nil == user {
		return handleNewUser(ctx, oauthCode, mfaCh, claims.Email, plainPass)
	}

	if !verifyTotp(mfaCh, user.MFASecret) {
		return c.Status(401)
	}

	return handleExistingUser(ctx, user.PassHash, plainPass, claims.Email, agent, ip, "Google")
}
