package relay

import (
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/hashicorp/errwrap"
)

// Document is serialized to JSON, signed using JWS and then POSTed to the target
// server where the signature must be verified.
type Document struct {
	Nonce      string            `json:"nonce"`
	Path       string            `json:"path"`
	Timestamp  int64             `json:"timestamp"`
	RequestID  string            `json:"request_id"`
	EntityID   string            `json:"entity_id,omitempty"`
	Parameters map[string]string `json:"params,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

func serializeDocument(doc Document, privKeyBytes []byte) ([]byte, error) {

	jws := jws.New(doc, crypto.SigningMethodRS512)

	privKey, err := crypto.ParseRSAPrivateKeyFromPEM(privKeyBytes)

	if err != nil {
		return nil, errwrap.Wrapf("cryptography issue: {{err}}", err)
	}

	jwsBytes, err := jws.General(privKey)
	if err != nil {
		return nil, errwrap.Wrapf("jws issue: {{err}}", err)
	}

	return jwsBytes, nil

}
