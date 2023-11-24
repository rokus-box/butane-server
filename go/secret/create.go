package main

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	c "lambda/common"
)

var (
	secretLimitReached = errors.New("secret limit reached")
)

func handleCreateSecret(ctx context.Context, p secretPayload, uID, vID string) (c.Res, error) {

	secret := c.NewSecret(p.DisplayName, p.URI, p.Username, p.Password, vID)

	if len(p.Metadata) > 0 {
		for _, md := range p.Metadata {
			secret.Metadata = append(secret.Metadata, c.NewMetadatum(md.Key, md.Value, md.Type))
		}
	}

	if err := saveSecret(ctx, secret, uID); nil != err {
		if errors.Is(err, secretLimitReached) {
			return c.Text("Secret limit reached", 400)
		}

		panic(err)
	}

	return c.Text(secret.ID, 201)
}

func saveSecret(ctx context.Context, s *c.Secret, uID string) error {
	uKey, _ := attributevalue.MarshalMap(c.MapS{
		"PK": "U#" + uID,
		"SK": "U#" + uID,
	})

	res, err := ddbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:            c.TableName,
		Key:                  uKey,
		ProjectionExpression: aws.String("secret_count"),
	})

	if nil != err {
		panic(err)
	}

	sCount := res.Item["secret_count"].(*types.AttributeValueMemberN).Value

	if sCount == SecretLimit {
		return secretLimitReached
	}

	sItem, _ := attributevalue.MarshalMap(c.MapA{
		"PK":           "V#" + s.VaultID,
		"SK":           "U#" + uID + "#SC#" + s.ID,
		"display_name": s.DisplayName,
		"uri":          s.URI,
		"username":     s.Username,
		"password":     s.Password,
	})

	if len(s.Metadata) > 0 {
		list, _ := attributevalue.MarshalList(s.Metadata)
		sItem["metadata"] = &types.AttributeValueMemberL{Value: list}
	}

	_, err = ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: c.TableName,
		Item:      sItem,
	})

	if nil != err {
		panic(err)
	}

	_, err = ddbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 c.TableName,
		Key:                       uKey,
		ExpressionAttributeValues: c.AtomicExpr(1),
		UpdateExpression:          aws.String("ADD secret_count :c"),
	})

	if nil != err {
		panic(err)
	}

	return err
}
