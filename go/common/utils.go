package common

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	gonanoid "github.com/matoous/go-nanoid"
)

// NewID returns a new random string with a length of 21 characters.
const NanoIDLength = 21

func NewID() string {
	nanoid, _ := gonanoid.Nanoid()

	return nanoid
}

func HashSHA256Base64(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return base64.StdEncoding.WithPadding(-1).EncodeToString(h.Sum(nil))
}

// Status returns a status code response with no body. If no status code is provided, 200 is used.
func Status(code int) (Res, error) {
	return Res{
		StatusCode: code,
		Headers:    MapS{"Content-Type": "text/plain"},
	}, nil
}

// Text returns a text/plain response with the given body and status code. If no status code is provided, 200 is used.
func Text[T ~string](body T, code ...int) (Res, error) {
	statusCode := 200
	if 0 < len(code) {
		statusCode = code[0]
	}

	return Res{
		Body:       string(body),
		StatusCode: statusCode,
		Headers:    MapS{"Content-Type": "text/plain"},
	}, nil
}

// JSON returns a application/json response with the given body and status code. If no status code is provided, 200 is used.
func JSON(body any, code ...int) (Res, error) {
	statusCode := 200
	if 0 < len(code) {
		statusCode = code[0]
	}

	bytes, err := json.Marshal(body)

	if nil != err {
		panic(err)
	}

	return Res{
		Body:       string(bytes),
		StatusCode: statusCode,
		Headers:    MapS{"Content-Type": "application/json"},
	}, nil
}
