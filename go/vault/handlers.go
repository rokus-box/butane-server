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

func handleGetVaults(ctx context.Context, uID string) (c.Res, error) {
	condExpr := aws.String("PK = :pk AND begins_with(SK, :sk)")
	attrValues, _ := attributevalue.MarshalMap(c.MapS{
		":pk": "U#" + uID,
		":sk": "V#",
	})

	res, err := ddbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:                 c.TableName,
		KeyConditionExpression:    condExpr,
		ExpressionAttributeValues: attrValues,
	})
	if nil != err {
		panic(err)
	}

	var vaults []c.Vault

	attributevalue.UnmarshalListOfMaps(res.Items, &vaults)

	for i := 0; i < len(vaults); i++ {
		vaults[i].ID = vaults[i].ID[2:]
	}

	return c.JSON(vaults)
}

func handleCreateVault(ctx context.Context, name, uID string) (c.Res, error) {
	vault := c.NewVault(name, uID)
	item, _ := attributevalue.MarshalMap(c.MapS{
		"PK":           "U#" + vault.UserID,
		"SK":           "V#" + vault.ID,
		"display_name": vault.DisplayName,
	})

	_, err := ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: c.TableName,
		Item:      item,
	})

	if nil != err {
		panic(err)
	}

	return c.Text(vault.ID, 201)
}

func handleUpdateVault(ctx context.Context, name, uID, vID string) (c.Res, error) {
	key, _ := attributevalue.MarshalMap(c.MapS{
		"PK": "U#" + uID,
		"SK": "V#" + vID,
	})

	expr, _ := attributevalue.MarshalMap(c.MapS{
		// r.Body is the new display name
		":d": name,
	})

	_, err := ddbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 c.TableName,
		Key:                       key,
		ExpressionAttributeValues: expr,
		ConditionExpression:       aws.String("attribute_exists(PK)"),
		UpdateExpression:          aws.String("SET display_name = :d"),
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

func handleDeleteVault(ctx context.Context, uID, vID string) (c.Res, error) {
	vaultKey, _ := attributevalue.MarshalMap(c.MapS{
		"PK": "U#" + uID,
		"SK": "V#" + vID,
	})

	_, err := ddbClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		Key:                 vaultKey,
		TableName:           c.TableName,
		ConditionExpression: aws.String("attribute_exists(PK)"),
	})

	if nil != err {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return c.Status(404)
		}
		panic(err)
	}

	secretsKey, _ := attributevalue.MarshalMap(c.MapS{
		":pk": "V#" + vID,
		":sk": "U#" + uID + "#SC",
	})

	res, err := ddbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:                 c.TableName,
		KeyConditionExpression:    aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ProjectionExpression:      aws.String("PK, SK"),
		ExpressionAttributeValues: secretsKey,
	})

	if nil != err {
		panic(err)
	}

	if len(res.Items) > 0 {
		var batch []types.WriteRequest

		for _, item := range res.Items {
			batch = append(batch, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: item,
				},
			})
		}

		_, err = ddbClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: c.MapL[[]types.WriteRequest]{
				*c.TableName: batch,
			},
		})

		if nil != err {
			panic(err)
		}
	}

	return c.Status(204)
}
