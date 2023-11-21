package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	c "lambda/common"
	"strings"
)

func handleGetSecrets(ctx context.Context, uID, vID string) (c.Res, error) {
	secretList := getSecrets(ctx, uID, vID)

	return c.JSON(secretList, 200)
}

func getSecrets(ctx context.Context, uID, vID string) []c.Secret {
	exprAttrValues, _ := attributevalue.MarshalMap(c.MapS{
		":pk": "V#" + vID,
		":sk": "U#" + uID + "#SC#",
	})

	res, err := ddbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:                 c.TableName,
		KeyConditionExpression:    aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: exprAttrValues,
		ProjectionExpression:      aws.String("PK, SK , display_name, uri, username, password, metadata"),
	})
	if err != nil {
		panic(err)
	}

	secretList := make([]c.Secret, len(res.Items))
	attributevalue.UnmarshalListOfMaps(res.Items, &secretList)

	for i := 0; i < len(secretList); i++ {
		split := strings.SplitAfterN(secretList[i].ID, "#SC#", 2)
		secretList[i].ID = split[1]
	}

	return secretList
}
