package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-playground/validator/v10"
	c "lambda/common"
	i "lambda/common/interceptors"
	"net/http"
)

// SecretLimit is string because DynamoDB stores numbers as strings.
const SecretLimit = "180"

var (
	ddbClient = c.NewDDB()
	validate  = validator.New(validator.WithRequiredStructEnabled())
)

type (
	secretPayload struct {
		DisplayName string             `json:"display_name" validate:"min=3,max=255"`
		URI         string             `json:"uri" validate:"omitempty,min=3,max=255"`
		Username    string             `json:"username,omitempty" validate:"omitempty,min=3,max=1024"`
		Password    string             `json:"password" validate:"omitempty,min=3,max=1024"`
		Metadata    []metadatumPayload `json:"metadata" validate:"omitempty,gte=1,lte=5,dive"`
	}
	metadatumPayload struct {
		Key   string `json:"key" validate:"min=1,max=255"`
		Value string `json:"value" validate:"min=1,max=255"`
		Type  uint8  `json:"type" validate:"min=0,max=2"`
	}
)

func handler(ctx context.Context, r c.Req) (c.Res, error) {
	uID := ctx.Value(c.UserIDKey).(string)
	vID := r.PathParameters["vaultId"]
	sID := r.PathParameters["id"]

	if len(vID) != c.NanoIDLength {
		return c.Text("Invalid resource identifier", 422)
	}

	if http.MethodPatch == r.HTTPMethod || http.MethodDelete == r.HTTPMethod {
		if len(sID) != c.NanoIDLength {
			return c.Text("Invalid resource identifier", 422)
		}
	}

	var payload secretPayload
	if http.MethodPost == r.HTTPMethod || http.MethodPatch == r.HTTPMethod {
		if err := json.Unmarshal([]byte(r.Body), &payload); nil != err {
			return c.Text("Invalid request body", 422)
		}

		if err := validate.Struct(payload); nil != err {
			return c.JSON(getErrors(err.(validator.ValidationErrors)), 422)
		}
	}

	switch r.HTTPMethod {
	case http.MethodGet:
		return handleGetSecrets(ctx, uID, vID)
	case http.MethodPost:
		return handleCreateSecret(ctx, payload, uID, vID)
	case http.MethodPatch:
		return handleUpdateSecret(ctx, payload, uID, vID, sID)
	case http.MethodDelete:
		return handleDeleteSecret(ctx, uID, vID, sID)
	default:
		return c.Status(405)
	}
}

func main() {
	icl := i.NewInterceptorList(handler)
	icl.Add(i.Recover)
	icl.Add(i.Auth(ddbClient))

	lambda.Start(icl.Intercept())
}

func getErrors(errs validator.ValidationErrors) []string {
	errorList := make([]string, len(errs))

	for j := 0; j < len(errs); j++ {
		switch errs[j].Field() {
		case "DisplayName":
			errorList[j] = "Display name must be between 3 and 255 characters and required"
		case "URI":
			errorList[j] = "URI must be between 3 and 255 characters"
		case "Username":
			errorList[j] = "Username must be between 3 and 1024 characters"
		case "Password":
			errorList[j] = "Password must be between 3 and 1024 characters"
		case "Metadata":
			errorList[j] = "Metadata must be between 1 and 5 items"
		default:
			md := errs[j].Namespace()[20:]
			field := errs[j].Field()

			switch field {
			case "Key":
				errorList[j] = md + " must be between 1 and 255 characters"
			case "Value":
				errorList[j] = md + " must be between 1 and 255 characters"
			case "Type":
				errorList[j] = md + " must be one of 0, 1, 2. If not specified, 0 is assumed"
			}
		}
	}

	return errorList
}
