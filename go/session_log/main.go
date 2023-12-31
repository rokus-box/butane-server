package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	c "lambda/common"
	"time"
)

var ddbClient = c.NewDDB()

func handler(ctx context.Context, ev events.DynamoDBEvent) {
	for _, record := range ev.Records {
		al := c.NewAuditLog(record.Change.Keys["SK"].String(), "")

		switch record.EventName {
		case "INSERT":
			al.Message = "Session created"
		case "REMOVE":
			al.Message = "Session deleted"
		default:
			fmt.Printf("Unknown event")
			jsonBytes, _ := json.Marshal(record)
			fmt.Println(string(jsonBytes))
			continue
		}

		item, _ := attributevalue.MarshalMap(c.MapA{
			"PK":           "U#" + al.UserID,
			"SK":           "AL#" + al.Timestamp.Format(time.RFC3339Nano),
			"message":      al.Message,
			"delete_after": time.Now().Add(time.Hour * 24 * 20).Unix(),
		})
		_, err := ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: c.TableName,
			Item:      item,
		})

		if nil != err {
			panic(err)
		}
	}
}

func main() {
	lambda.Start(handler)
}
