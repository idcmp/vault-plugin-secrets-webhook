package relay

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/logical"
)

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {

	return nil, fmt.Errorf("not yet")
}
