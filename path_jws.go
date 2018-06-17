package webhook

import (
	"context"
	"fmt"

	"github.com/SermoDigital/jose/crypto"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathFetchJwsCertificate(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: `keys/jws/certificate`,
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathReadJwsCertificate,
		},
	}
}

func pathConfigJws(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: `config/keys/jws`,
		Fields: map[string]*framework.FieldSchema{
			"certificate": {
				Type:        framework.TypeString,
				Description: `PEM encoded public certificate`,
			},
			"private_key": {
				Type:        framework.TypeString,
				Description: `PEM encoded private key`,
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathWriteJwsKeys,
			logical.CreateOperation: b.pathWriteJwsKeys,
		},
		//HelpSynopsis:    pathFetchHelpSyn,
		//HelpDescription: pathFetchHelpDesc,
	}
}

func (b *backend) pathReadJwsCertificate(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {

	b.Logger().Debug("pathReadJwsCertificate", "ctx", ctx, "req", req, "data", data)
	b.Lock.RLock()
	defer b.Lock.RUnlock()

	entry, err := req.Storage.Get(ctx, "config/keys/jws/certificate")
	if err != nil {
		return nil, errwrap.Wrapf("could not get public certificate: {{err}}", err)
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"certificate": entry.Value,
		},
	}, nil
}

func (b *backend) pathWriteJwsKeys(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Debug("pathWriteJwsKeys", "ctx", ctx, "req", req, "data", data)
	b.Lock.Lock()
	defer b.Lock.Unlock()

	certificate, ok := data.GetOk("certificate")
	if !ok {
		return nil, fmt.Errorf("certificate is required")
	}

	privKey, ok := data.GetOk("private_key")
	if !ok {
		return nil, fmt.Errorf("private_key is required")
	}

	certBytes := []byte(certificate.(string))
	privKeyBytes := []byte(privKey.(string))

	if _, err := crypto.ParseRSAPublicKeyFromPEM(certBytes); err != nil {
		return nil, errwrap.Wrapf("could not parse certificate: {{err}}", err)
	}

	if _, err := crypto.ParseRSAPrivateKeyFromPEM(privKeyBytes); err != nil {
		return nil, errwrap.Wrapf("could not parse private_key: {{err}}", err)
	}

	publicEntry := &logical.StorageEntry{
		Key:   "config/keys/jws/certificate",
		Value: certBytes,
	}

	privateEntry := &logical.StorageEntry{
		Key:   "config/keys/jws/private_key",
		Value: privKeyBytes,
	}

	if err := req.Storage.Put(ctx, publicEntry); err != nil {
		return nil, errwrap.Wrapf("could not store certificate: {{err}}", err)
	}
	if err := req.Storage.Put(ctx, privateEntry); err != nil {
		return nil, errwrap.Wrapf("could not store private_key: {{err}}", err)
	}
	return &logical.Response{}, nil
}
