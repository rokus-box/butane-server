package common

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"os"
)

var TableName = aws.String(os.Getenv("TABLE_NAME"))

func NewDDB() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-central-1"))

	if nil != err {
		panic(err)
	}

	return dynamodb.NewFromConfig(cfg)
}

func AtomicExpr(a int) map[string]types.AttributeValue {
	if a == 0 {
		return nil
	}

	expr, _ := attributevalue.MarshalMap(MapA{
		":c": a,
	})

	return expr
}
