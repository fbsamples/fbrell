// Package signedrequest implements generic signed request logic.
package signedrequest

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	padding             = byte('=')
	dot                 = byte('.')
	ErrInvalidSignature = errors.New("Invalid signature")
)

// Unmarshal the data from a signed_request and validate the signature.
func Unmarshal(data []byte, secret []byte, v interface{}) error {
	givenSigBase64, payloadBase64, err := split(data)
	if err != nil {
		return err
	}
	givenSigBytes, err := decode(givenSigBase64)
	if err != nil {
		return fmt.Errorf("Could not decode signature as base64 data: %s", err)
	}
	expectedSigBytes, err := sig(payloadBase64, secret)
	if err != nil {
		return fmt.Errorf("Could not generate expected signature: %s", err)
	}
	if bytes.Compare(expectedSigBytes, givenSigBytes) != 0 {
		return ErrInvalidSignature
	}
	payloadBytes, err := decode(payloadBase64)
	if err != nil {
		return fmt.Errorf("Could not decode payload as base64 data: %s", err)
	}
	err = json.Unmarshal(payloadBytes, v)
	if err != nil {
		return fmt.Errorf("Error in json.Unmarshal of payload bytes: %s", err)
	}
	return nil
}

// Generate signature for the payload.
func sig(payload []byte, secret []byte) ([]byte, error) {
	h := hmac.New(sha256.New, secret)
	_, err := h.Write(payload)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// Split the data into signature and payload on the dot character.
func split(data []byte) ([]byte, []byte, error) {
	pos := bytes.IndexByte(data, dot)
	if pos == -1 {
		return nil, nil, errors.New(
			"Could not find dot separator in signed request.")
	}
	return data[:pos], data[pos+1:], nil
}

// Padding is wasteful, but go wants it.
func fixPadding(b64 []byte) []byte {
	count := 4 - len(b64)%4
	for count > 0 && count < 4 {
		count--
		b64 = append(b64, padding)
	}
	return b64
}

// Decode some base64 URL encoded bytes into bytes.
func decode(b64 []byte) ([]byte, error) {
	b64 = fixPadding(b64)
	dst := make([]byte, base64.URLEncoding.DecodedLen(len(b64)))
	n, err := base64.URLEncoding.Decode(dst, b64)
	if err != nil {
		return nil, err
	}
	return dst[:n], err
}
