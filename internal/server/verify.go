//
// Adapted from stripe-go: https://github.com/stripe/stripe-go/blob/63c2c151964f81b4f8cad1d8a2d2c773dd9f7aaa/webhook/client.go
//
package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	rt "gitlab.com/clickport/clickport-agent/internal/clickport"
	cfg "gitlab.com/clickport/clickport-agent/internal/config"
)

var (
	errInvalidHeader    = errors.New("request has invalid Clickport-Signature header")
	errNoValidSignature = errors.New("request had no valid signature")
	errNotSigned        = errors.New("request has no Clickport-Signature header")
	errTooOld           = errors.New("timestamp wasn't within tolerance")
)

const (
	// DefaultTolerance indicates that signatures older than this will be rejected.
	DefaultTolerance time.Duration = 300 * time.Second
	// signingVersion represents the version of the signature we currently use.
	signingVersion string = "v1"
)

type signedHeader struct {
	timestamp  time.Time
	signatures [][]byte
}

func parseSignatureHeader(header string) (*signedHeader, error) {
	sh := &signedHeader{}

	if header == "" {
		return sh, errNotSigned
	}

	// Signed header looks like "t=1495999758,v1=ABC,v1=DEF,v0=GHI"
	pairs := strings.Split(header, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return sh, errInvalidHeader
		}

		switch parts[0] {
		case "t":
			timestamp, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return sh, errInvalidHeader
			}
			sh.timestamp = time.Unix(timestamp, 0)

		case signingVersion:
			sig, err := hex.DecodeString(parts[1])
			if err != nil {
				continue // Ignore invalid signatures
			}

			sh.signatures = append(sh.signatures, sig)

		default:
			continue // Ignore unknown parts of the header
		}
	}

	if len(sh.signatures) == 0 {
		return sh, errNoValidSignature
	}

	return sh, nil
}

func computeSignature(t time.Time, payload []byte, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d", t.Unix())))
	mac.Write([]byte("."))
	mac.Write(payload)
	return mac.Sum(nil)
}

func validatePayload(payload []byte, sigHeader string, secret string, tolerance time.Duration) error {
	header, err := parseSignatureHeader(sigHeader)
	if err != nil {
		return err
	}

	expectedSignature := computeSignature(header.timestamp, payload, secret)
	expiredTimestamp := time.Since(header.timestamp) > tolerance
	if expiredTimestamp {
		return errTooOld
	}

	// Check all given v1 signatures, multiple signatures will be sent temporarily in the case of a rolled signature secret
	for _, sig := range header.signatures {
		if hmac.Equal(expectedSignature, sig) {
			return nil
		}
	}

	return errNoValidSignature
}

func getPayload(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	return ioutil.ReadAll(r.Body)
}

func VerifyRequestSignature(config *cfg.Config, w http.ResponseWriter, r *http.Request) error {
	payload, err := getPayload(w, r)
	if err != nil {
		return fmt.Errorf("error reading request body: %v", err)
	}

	return validatePayload(payload, r.Header.Get("Clickport-Signature"), config.SigningSecret, DefaultTolerance)
}

func ConstructExecutionRequest(config *cfg.Config, w http.ResponseWriter, r *http.Request) (*rt.ExecutionRequest, error) {
	payload, err := getPayload(w, r)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	err = validatePayload(payload, r.Header.Get("Clickport-Signature"), config.SigningSecret, DefaultTolerance)
	if err != nil {
		return nil, err
	}

	executionRequest := rt.ExecutionRequest{}
	if err := json.Unmarshal(payload, &executionRequest); err != nil {
		return nil, fmt.Errorf("failed to parse request body json")
	}

	return &executionRequest, nil
}
