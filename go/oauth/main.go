package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pquerna/otp/totp"
	c "lambda/common"
	"time"
)

func handler(ctx context.Context, r c.Req) (c.Res, error) {
	provider := r.PathParameters["provider"]

	switch provider {
	case "google":
		return handleGoogle(ctx, r)
	default:
		return c.Text("Invalid provider", 400)
	}
}

func main() {
	lambda.Start(handler)
}

func getMfaSecret(ctx context.Context, ddb *dynamodb.Client, id string) string {
	key, _ := attributevalue.MarshalMap(c.MapS{
		"PK": "U#" + id,
		"SK": "U#" + id,
	})

	resp, err := ddb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:            c.TableName,
		Key:                  key,
		ProjectionExpression: aws.String("mfa_secret"),
	})

	if nil != err {
		panic(err)
	}

	u := &c.User{}

	attributevalue.UnmarshalMap(resp.Item, u)

	return u.MFASecret
}

// registerUserWithSeed registers the user in DynamoDB with a seed vault and secret
func registerUserWithSeed(ctx context.Context, ddb *dynamodb.Client, u *c.User) error {
	uItem, _ := attributevalue.MarshalMap(c.MapS{
		"PK":         "U#" + u.Email,
		"SK":         "U#" + u.Email,
		"mfa_secret": u.MFASecret,
	})

	v := c.NewVault("My First Vault", u.Email)
	vItem, _ := attributevalue.MarshalMap(c.MapS{
		"PK":           "U#" + u.Email,
		"SK":           "V#" + v.ID,
		"display_name": v.DisplayName,
	})

	s := c.NewSecret("My Google Account", "https://google.com", "example@gmail.com", "s3cr3t_p455w0rd!", v.ID)

	sItem, _ := attributevalue.MarshalMap(c.MapA{
		"PK":           "V#" + v.ID,
		"SK":           "U#" + u.Email + "#SC#" + s.ID,
		"display_name": s.DisplayName,
		"uri":          s.URI,
		"username":     s.Username,
		"password":     s.Password,
		"metadata": []*c.Metadatum{
			c.NewMetadatum("My MFA Secret", "JBSWY3DPEHPK3PXP", c.MetadatumTypeMFA),
			c.NewMetadatum("Example Note", "This is a plain-text note", c.MetadatumTypeText),
			c.NewMetadatum("My Confidential Note", "This should not appear on UI", c.MetadatumTypeConfidential),
		},
	})

	batch := []types.WriteRequest{
		{PutRequest: &types.PutRequest{Item: uItem}},
		{PutRequest: &types.PutRequest{Item: vItem}},
		{PutRequest: &types.PutRequest{Item: sItem}},
	}

	_, err := ddb.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: c.MapL[[]types.WriteRequest]{
			*c.TableName: batch,
		},
	})

	return err
}

// saveSession saves the session to DynamoDB and returns the token
func saveSession(ctx context.Context, ddb *dynamodb.Client, s *c.Session) string {
	item, _ := attributevalue.MarshalMap(c.MapA{
		"PK":         "SS#" + s.Token,
		"SK":         s.UserID,
		"TTL":        s.TTL,
		"user_agent": s.UserAgent,
		"ip_address": s.IPAddress,
		"timestamp":  s.Timestamp.Unix(),
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
