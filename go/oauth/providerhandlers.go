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

func handleGoogle(ctx context.Context, r c.Req) (c.Res, error) {
	mfaCh := r.Headers["X-Mfa-Challenge"]
	if "" == mfaCh {
		return c.Text("X-Mfa-Challenge header is required", 401)
	}

	agent, ip := r.RequestContext.Identity.UserAgent, r.RequestContext.Identity.SourceIP
	if "" == agent {
		return c.Text("User-Agent header is required", 401)
	}

	respBody := &googleResp{}
	oauthCode := r.Body
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

	userMfa := getMfaSecret(ctx, ddbClient, claims.Email)

	if "" == userMfa {
		otpSecret := genOtp(oauthCode[len(oauthCode)-10:], "#")
		if !verifyTotp(mfaCh, otpSecret) {
			return c.Text("Invalid MFA code", 401)
		}

		user := c.NewUser(claims.Email)
		user.MFASecret = otpSecret
		if nil != registerUserWithSeed(ctx, ddbClient, user) {
			return c.Text("Failed to register user. Please try again later", 503)
		}

		sess := c.NewSession(agent, ip, claims.Email)
		return c.Text(saveSession(ctx, ddbClient, sess), 201)
	}

	if !verifyTotp(mfaCh, userMfa) {
		return c.Text("Invalid MFA code", 401)
	}

	sess := c.NewSession(agent, ip, claims.Email)
	return c.Text(saveSession(ctx, ddbClient, sess), 201)
}
