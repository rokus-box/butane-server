package interceptors

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	c "lambda/common"
	"log"
)

// HTTP interceptors for AWS Lambda functions.
/*
	Usage 1: lambda.Start(SomeInterceptor(handler))

	Usage 2: lambda.Start(AInterceptor(BInterceptor(CInterceptor(handler))))

	Usage 3 interceptorList := NewInterceptorList(handler)
			interceptorList.Add(AInterceptor)
			interceptorList.Add(BInterceptor)
			interceptorList.Add(CInterceptor)
			lambda.Start(interceptorList.Intercept())

	All usages are valid. The chain is syntactic sugar for the second usage.
 	The interceptor chain runs in the order they are added.
	For example 3. AInterceptor runs first, then BInterceptor, then CInterceptor.
*/

type Interceptor func(c.Handler) c.Handler
type InterceptorChain struct {
	interceptors []Interceptor
	handler      c.Handler
	isBuilt      bool
}

// Intercept initializes runs the chain and returns the final handler every time it is invoked.
func (ic *InterceptorChain) Intercept() c.Handler {
	if ic.isBuilt {
		panic("InterceptorChain can only be built once")
	}

	ic.isBuilt = true

	h := ic.handler

	for i := len(ic.interceptors) - 1; i >= 0; i-- {
		h = ic.interceptors[i](h)
	}

	return h
}

// NewInterceptorList returns a new interceptor chain for HTTP v2.
func NewInterceptorList(h c.Handler) *InterceptorChain {
	return &InterceptorChain{
		interceptors: make([]Interceptor, 0),
		handler:      h,
	}
}

// Add adds an interceptor to the chain.
func (ic *InterceptorChain) Add(i Interceptor) {
	ic.interceptors = append(ic.interceptors, i)
}

// Recover recovers from panics and logs them.
func Recover(h c.Handler) c.Handler {
	return func(ctx context.Context, r c.Req) (c.Res, error) {
		defer func() {
			if err := recover(); nil != err {
				log.Println("panic:", err)
			}
		}()

		return h(ctx, r)
	}
}

// Auth checks if the user is authenticated. If not, a 401 is returned.
// If the user is authenticated, the user Email is added to the context.
func Auth(ddb *dynamodb.Client) func(h c.Handler) c.Handler {
	return func(h c.Handler) c.Handler {
		return func(ctx context.Context, r c.Req) (c.Res, error) {
			token := r.Headers["Authorization"]
			uID := getUserID(ctx, ddb, token)

			if "" == uID {
				return c.Status(401)
			}

			ctx = context.WithValue(ctx, c.UserIDKey, uID)

			return h(ctx, r)
		}
	}
}

// getUserID returns the user Email from the given token.
// If the token is invalid for any reason, an empty string is returned.
func getUserID(ctx context.Context, ddb *dynamodb.Client, token string) string {
	exprAttrValues, _ := attributevalue.MarshalMap(c.MapS{
		":pk": "SS#" + token,
	})

	res, err := ddb.Query(ctx, &dynamodb.QueryInput{
		TableName:                 c.TableName,
		KeyConditionExpression:    aws.String("PK = :pk"),
		ProjectionExpression:      aws.String("SK"),
		ExpressionAttributeValues: exprAttrValues,
	})

	if 0 == res.Count {
		return ""
	}

	if nil != err {
		panic(err)
	}

	sess := c.Session{}

	attributevalue.UnmarshalMap(res.Items[0], &sess)

	return sess.UserID
}
