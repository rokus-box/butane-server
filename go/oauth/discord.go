package main

import (
	"context"
	"github.com/go-resty/resty/v2"
	c "lambda/common"
	"log"
	"os"
)

type discordResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type meResponse struct {
	Id                   string      `json:"id"`
	Username             string      `json:"username"`
	Avatar               string      `json:"avatar"`
	Discriminator        string      `json:"discriminator"`
	PublicFlags          int         `json:"public_flags"`
	PremiumType          int         `json:"premium_type"`
	Flags                int         `json:"flags"`
	Banner               string      `json:"banner"`
	AccentColor          int         `json:"accent_color"`
	GlobalName           string      `json:"global_name"`
	AvatarDecorationData interface{} `json:"avatar_decoration_data"`
	BannerColor          string      `json:"banner_color"`
	MfaEnabled           bool        `json:"mfa_enabled"`
	Locale               string      `json:"locale"`
	Email                string      `json:"email"`
	Verified             bool        `json:"verified"`
}

var discordClient = resty.New().SetBaseURL("https://discord.com/api")

func handleDiscord(ctx context.Context, agent, ip, mfaCh, oauthCode, plainPass string) (c.Res, error) {
	respBody := &discordResponse{}

	resp, err := discordClient.R().
		SetFormData(map[string]string{
			"grant_type":   "authorization_code",
			"code":         oauthCode,
			"redirect_uri": os.Getenv("DISCORD_REDIRECT_URI"),
		}).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBasicAuth(os.Getenv("DISCORD_CLIENT_ID"), os.Getenv("DISCORD_CLIENT_SECRET")).
		SetResult(respBody).
		Post("/oauth2/token")

	if os.IsTimeout(err) {
		return c.Text("Request to provider timed out. Please try again later", 503)
	}

	if nil != err {
		log.Println("failed to authenticate with Discord: ", err)
		return c.Text("Failed to authenticate. Please try again later", 503)
	}

	if resp.IsError() {
		return c.Text("Invalid authentication details were provided", 401)
	}

	meResp := &meResponse{}

	resp, err = discordClient.R().
		SetHeader("Authorization", "Bearer "+respBody.AccessToken).
		SetResult(meResp).
		Get("/users/@me")

	if os.IsTimeout(err) {
		return c.Text("Request to provider timed out. Please try again later", 503)
	}

	if nil != err {
		log.Println("failed to authenticate with Discord: ", err)
		return c.Text("Failed to authenticate. Please try again later", 503)
	}

	if resp.IsError() {
		return c.Text("Invalid authentication details were provided", 401)
	}

	if !meResp.Verified {
		return c.Text("Email address is not verified", 401)
	}

	user := getUser(ctx, ddbClient, meResp.Email)

	if nil == user {
		return handleNewUser(ctx, oauthCode, mfaCh, meResp.Email, plainPass)
	}

	return handleExistingUser(ctx, user.PassHash, plainPass, meResp.Email, agent, ip, "Discord")
}
