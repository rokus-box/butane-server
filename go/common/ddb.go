package common

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"os"
)

var TableName = aws.String(os.Getenv("TABLE_NAME"))

type resolver struct{}

func (r resolver) ResolveEndpoint(service string, region string, options ...interface{}) (aws.Endpoint, error) {
	return aws.Endpoint{
		URL: "http://192.168.1.12:8000",
	}, nil
}

func NewDDB() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("eu-central-1"),
		config.WithEndpointResolverWithOptions(resolver{}),
	)

	if nil != err {
		panic(err)
	}

	return dynamodb.NewFromConfig(cfg)
}
