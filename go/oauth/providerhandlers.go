package main

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	c "lambda/common"
	"log"
	"os"
	"strings"
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

func handleGoogle(ctx context.Context, r c.Req, agent, ip, mfaCh string) (c.Res, error) {
	respBody := &googleResp{}

	body := strings.SplitN(r.Body, "\n", 2)

	if len(body) != 2 {
		return c.Text("Malformed request body", 400)
	}

	oauthCode := body[0]
	plainPass := body[1]

	if !isValidPass(plainPass) {
		return c.Text("Password must match "+
			"^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]{12,73}$ pattern", 400)
	}

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
		otpSecret := genOtp(oauthCode[len(oauthCode)-10:], "#")
		if !verifyTotp(mfaCh, otpSecret) {
			return c.Status(401)
		}

		newUser := c.NewUser(claims.Email)
		newUser.MFASecret = otpSecret
		passHash, err := bcrypt.GenerateFromPassword([]byte(plainPass), bcrypt.DefaultCost)

		if nil != err {
			log.Println("failed to hash password: ", err)
			return c.Text("Failed to register user. Please try again later", 503)
		}

		if nil != saveUser(ctx, ddbClient, newUser, string(passHash)) {
			return c.Text("Failed to register user. Please try again later", 503)
		}

		sess := c.NewSession(agent, ip, claims.Email)
		return c.Text(saveSession(ctx, ddbClient, sess), 201)
	}

	if !verifyTotp(mfaCh, user.MFASecret) {
		return c.Status(401)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(plainPass))
	if nil != err {
		return c.Status(401)
	}

	sess := c.NewSession(agent, ip, claims.Email)
	return c.Text(saveSession(ctx, ddbClient, sess), 201)
}
