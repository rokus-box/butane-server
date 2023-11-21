package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	c "lambda/common"
)

func handleCreateSecret(ctx context.Context, p secretPayload, uID, vID string) (c.Res, error) {

	secret := c.NewSecret(p.DisplayName, p.URI, p.Username, p.Password, vID)

	if len(p.Metadata) > 0 {
		for _, md := range p.Metadata {
			secret.Metadata = append(secret.Metadata, c.NewMetadatum(md.Key, md.Value, md.Type))
		}
	}

	if err := saveSecret(ctx, secret, uID); nil != err {
		panic(err)
	}

	return c.Text(secret.ID, 201)
}

func saveSecret(ctx context.Context, s *c.Secret, uID string) error {
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

	_, err := ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: c.TableName,
		Item:      sItem,
	})

	return err
}
