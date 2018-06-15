package webhook

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathVerify(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: `verify/(?P<nonce>.+)`,
		Fields: map[string]*framework.FieldSchema{
			"nonce": {
				Type:        framework.TypeString,
				Description: `Nonce from active call to a destination.`,
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathVerifyNonce,
		},
		// TODO -v
		//HelpSynopsis:    pathFetchHelpSyn,
		//HelpDescription: pathFetchHelpDesc,
	}
}

func (b *backend) pathVerifyNonce(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {

	b.Lock.RLock()
	defer b.Lock.RUnlock()

	nonce, ok, err := data.GetOkErr("nonce")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("unspecified nonce")
	}

	entry, err := req.Storage.Get(ctx, "verify/"+nonce.(string))
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	return &logical.Response{
		Data: map[string]interface{}{
			"payload": entry.Value,
		},
	}, nil
}
