package common

import (
	"testing"
)

func TestNewID(t *testing.T) {
	id := NewID()
	if NanoIDLength != len(id) {
		t.Errorf("Expected Email length of %d, but got %d", NanoIDLength, len(id))
	}
}

func TestHashSHA256Base64(t *testing.T) {
	input := "test string"
	expectedHash := "1VecRt/MfxggcBPmW0Tky04sIpj0rEV7qPgnQ/Mekws"
	hash := HashSHA256Base64(input)

	if hash != expectedHash {
		t.Errorf("HashSHA256Base64(%s) = %s; want %s", input, hash, expectedHash)
	}
}

func TestStatus(t *testing.T) {
	statusCode := 400
	resp, err := Status(statusCode)
	if nil != err {
		t.Errorf("Status(%d) returned an error: %v", statusCode, err)
	}
	if statusCode != resp.StatusCode {
		t.Errorf("Status(%d) returned StatusCode %d; want %d", statusCode, resp.StatusCode, statusCode)
	}

	resp, err = Status(200)
	if nil != err {
		t.Errorf("Status() returned an error: %v", err)
	}

	if 200 != resp.StatusCode {
		t.Errorf("Status() returned StatusCode %d; want %d", resp.StatusCode, 200)
	}
}

func TestText(t *testing.T) {
	statusCode := 201
	body := "Hello, World!"
	resp, err := Text(body, statusCode)
	if nil != err {
		t.Errorf("Text(%s, %d) returned an error: %v", body, statusCode, err)
	}
	if statusCode != resp.StatusCode {
		t.Errorf("Text(%s, %d) returned StatusCode %d; want %d", body, statusCode, resp.StatusCode, statusCode)
	}
	if body != resp.Body {
		t.Errorf("Text(%s, %d) returned Body %s; want %s", body, statusCode, resp.Body, body)
	}

	resp, err = Text(body)
	if nil != err {
		t.Errorf("Text(%s) returned an error: %v", body, err)
	}

	if 200 != resp.StatusCode {
		t.Errorf("Text(%s) returned StatusCode %d; want %d", body, resp.StatusCode, 200)
	}
}

func TestJSON(t *testing.T) {
	statusCode := 201
	body := MapA{"key": "value"}
	resp, err := JSON(body, statusCode)
	if nil != err {
		t.Errorf("JSON(%v, %d) returned an error: %v", body, statusCode, err)
	}
	if resp.StatusCode != statusCode {
		t.Errorf("JSON(%v, %d) returned StatusCode %d; want %d", body, statusCode, resp.StatusCode, statusCode)
	}
	expectedJSON := `{"key":"value"}`
	if resp.Body != expectedJSON {
		t.Errorf("JSON(%v, %d) returned Body %s; want %s", body, statusCode, resp.Body, expectedJSON)
	}

	resp, err = JSON(body)
	if nil != err {
		t.Errorf("JSON(%v) returned an error: %v", body, err)
	}

	if 200 != resp.StatusCode {
		t.Errorf("JSON(%v) returned StatusCode %d; want %d", body, resp.StatusCode, 200)
	}
}
