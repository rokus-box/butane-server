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

func handleDeleteSecret(ctx context.Context, uID, vID, sID string) (c.Res, error) {
	key, _ := attributevalue.MarshalMap(c.MapS{
		"PK": "V#" + vID,
		"SK": "U#" + uID + "#SC#" + sID,
	})

	_, err := ddbClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName:           c.TableName,
		Key:                 key,
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
