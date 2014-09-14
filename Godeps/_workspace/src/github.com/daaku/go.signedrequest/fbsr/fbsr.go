// Package fbsr implements the standard Facebook signed_request structures.
package fbsr

import (
	"errors"
	"flag"
	"time"

	"github.com/daaku/go.signedrequest"
)

var (
	maxAge = flag.Duration(
		"fbsr.max-age",
		time.Hour*24,
		"Max age of signed request to consider it valid.")
	ErrExpired = errors.New("signed_request has expired.")
)

type Timestamp int64

type SignedRequest struct {
	Algorithm   string    `json:"algorithm"`
	IssuedAt    Timestamp `json:"issued_at"`
	ExpiresAt   Timestamp `json:"expires,omitempty"`
	Code        string    `json:"code,omitempty"`
	AccessToken string    `json:"oauth_token,omitempty"`
	UserID      uint64    `json:"user_id,string,omitempty"`
	AppData     string    `json:"app_data,omitempty"`

	User *struct {
		Country string `json:"country,omitempty"`
		Locale  string `json:"locale,omitempty"`
		Age     *struct {
			Min uint `json:"min,omitempty"`
		} `json:"age,omitempty"`
	} `json:"user,omitempty"`

	Page *struct {
		ID    uint64 `json:"id,string"`
		Liked bool   `json:"liked"`
		Admin bool   `json:"admin"`
	} `json:"page,omitempty"`
}

// Unmarshal a Facebook signed request.
func Unmarshal(data []byte, secret []byte) (*SignedRequest, error) {
	sr := &SignedRequest{}
	err := signedrequest.Unmarshal(data, secret, sr)
	if err != nil {
		return nil, err
	}
	if sr.IssuedAt == 0 || time.Now().After(sr.IssuedAt.Time().Add(*maxAge)) {
		return nil, ErrExpired
	}
	return sr, err
}

func (t Timestamp) Time() time.Time {
	return time.Unix(int64(t), 0)
}
