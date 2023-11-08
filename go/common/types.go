package common

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
)

type (
	MapA        map[string]any
	MapS        map[string]string
	MapL[T any] map[string]T
	// Handler is the type of the lambda handler function.
	Handler func(context.Context, Req) (Res, error)
	Req     events.APIGatewayProxyRequest
	Res     events.APIGatewayV2HTTPResponse
)

// Context keys for the context object passed to the handler.
const (
	UserIDKey uint8 = iota
)
