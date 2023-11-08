package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	c "lambda/common"
	i "lambda/common/interceptors"
	v "lambda/common/validators"
	"net/http"
)

var ddbClient = c.NewDDB()

func handler(ctx context.Context, r c.Req) (c.Res, error) {
	uID := ctx.Value(c.UserIDKey).(string)
	vID := r.PathParameters["id"]

	if http.MethodPatch == r.HTTPMethod || http.MethodDelete == r.HTTPMethod {
		if len(vID) != c.NanoIDLength {
			return c.Text("Invalid resource identifier", 422)
		}
	}

	switch r.HTTPMethod {
	case http.MethodGet:
		return handleGetVaults(ctx, r, uID)
	case http.MethodPost:
		return handleCreateVault(ctx, r, uID)
	case http.MethodPatch:
		return handleUpdateVault(ctx, r, uID, vID)
	case http.MethodDelete:
		return handleDeleteVault(ctx, r, uID, vID)
	default:
		return c.Text("", 405)
	}
}

func main() {
	icl := i.NewInterceptorList(handler)
	icl.Add(i.Recover)
	icl.Add(i.Auth(ddbClient))
	icl.Add(writeVaultInterceptor)

	lambda.Start(icl.Intercept())
}

// writeVaultInterceptor validates the request body for POST and PATCH requests (write operations).
func writeVaultInterceptor(h c.Handler) c.Handler {
	return func(ctx context.Context, r c.Req) (c.Res, error) {
		if http.MethodPost == r.HTTPMethod || http.MethodPatch == r.HTTPMethod {
			m := "Vault display name must be between 3 and 32 characters"
			if msg, ok := v.Length(r.Body, 2, 32, m); !ok {
				return c.Text(msg, 400)
			}
		}

		return h(ctx, r)
	}
}
