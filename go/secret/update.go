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

func handleUpdateSecret(ctx context.Context, p secretPayload, uID, vID, sID string) (c.Res, error) {
	sItem, _ := attributevalue.MarshalMap(c.MapA{
		"PK":           "V#" + vID,
		"SK":           "U#" + uID + "#SC#" + sID,
		"display_name": p.DisplayName,
		"uri":          p.URI,
		"username":     p.Username,
		"password":     p.Password,
	})

	if len(p.Metadata) > 0 {
		list, _ := attributevalue.MarshalList(p.Metadata)
		sItem["metadata"] = &types.AttributeValueMemberL{Value: list}
	}

	_, err := ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           c.TableName,
		Item:                sItem,
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
	})

	if nil != err {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return c.Status(404)
		}

		panic(err)
	}

	return c.Status(204)

}
