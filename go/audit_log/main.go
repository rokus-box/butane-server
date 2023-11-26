package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	c "lambda/common"
	i "lambda/common/interceptors"
)

var ddbClient = c.NewDDB()

func handler(ctx context.Context, r c.Req) (c.Res, error) {
	uID := ctx.Value(c.UserIDKey).(string)
	condExpr := aws.String("PK = :pk AND begins_with(SK, :sk)")
	attrValues, _ := attributevalue.MarshalMap(c.MapS{
		":pk": "U#" + uID,
		":sk": "AL#",
	})

	res, err := ddbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:                 c.TableName,
		KeyConditionExpression:    condExpr,
		ExpressionAttributeValues: attrValues,
	})
	if nil != err {
		panic(err)
	}

	var logs []c.AuditLog

	// Remove the "AL#" prefix from the SK
	for _, item := range res.Items {
		item["SK"] = &types.AttributeValueMemberS{
			Value: item["SK"].(*types.AttributeValueMemberS).Value[3:],
		}
	}

	attributevalue.UnmarshalListOfMaps(res.Items, &logs)

	return c.JSON(logs)
}

func main() {
	icl := i.NewInterceptorList(handler)
	icl.Add(i.Recover)
	icl.Add(i.Auth(ddbClient))

	lambda.Start(icl.Intercept())
}
