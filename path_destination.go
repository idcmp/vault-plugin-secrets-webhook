package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathDestination(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: `destination/(?P<target_name>.+)`,
		Fields: map[string]*framework.FieldSchema{
			"target_name": {
				Type:        framework.TypeString,
				Description: `Unique name representing a specific target.`,
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.contactDestination,
			logical.ReadOperation:   b.pingDestination,
		},

		//HelpSynopsis:    pathFetchHelpSyn,
		//HelpDescription: pathFetchHelpDesc,
	}
}

func pathConfigDestination(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: `config/destination/(?P<target_name>.+)`,
		Fields: map[string]*framework.FieldSchema{
			"target_name": {
				Type:        framework.TypeString, // TODO: is this right? lowercase? stringname?
				Description: `Unique name representing a specific target.`,
			},
			"target_url": {
				Type:        framework.TypeString,
				Description: "", // TODO
			},
			"params": {
				Type:        framework.TypeKVPairs,
				Description: "", // TODO
			},
			"send_entity_id": {
				Type:        framework.TypeBool,
				Description: "", // TODO
				Default:     true,
			},
			"timeout": {
				Type:        framework.TypeDurationSecond,
				Description: "", // TODO
				Default:     60,
			},
			"target_ca": {
				Type:        framework.TypeString,
				Description: "", // TODO
			},
			"metadata": {
				Type:        framework.TypeKVPairs,
				Description: "", // TODO
			},
			"follow_redirects": {
				Type:        framework.TypeBool,
				Description: "", // TODO
				Default:     false,
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.CreateOperation: b.createDestination,
			logical.ReadOperation:   b.readDestination,
			logical.UpdateOperation: b.updateDestination,
			logical.DeleteOperation: b.deleteDestination,
		},
		ExistenceCheck: b.destinationExistenceCheck,
		//HelpSynopsis:    pathFetchHelpSyn,
		//HelpDescription: pathFetchHelpDesc,
	}
}

type Destination struct {
	TargetUrl       string            `json:"target_url"`
	SendEntityId    bool              `json:"send_entity_id"`
	Timeout         int               `json:"timeout"`
	FollowRedirects bool              `json:"follow_redirects"`
	Parameters      []string          `json:"params"`
	Metadata        map[string]string `json:"metadata"`
}

func (b *backend) createDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {

	var d Destination

	d.TargetUrl = data.Get("target_url").(string)
	d.SendEntityId = data.Get("send_entity_id").(bool)
	d.Timeout = data.Get("timeout").(int)
	d.FollowRedirects = data.Get("follow_redirects").(bool)

	// TODO : mandatory target url

	buf, _ := json.Marshal(d)

	entry := &logical.StorageEntry{
		Key:   req.Path,
		Value: buf,
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, errwrap.Wrapf("failed to write: {{err}}", err)
	}

	return &logical.Response{}, nil
}

func (b *backend) readDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {

	entry, _ := req.Storage.Get(ctx, req.Path)

	var d Destination

	if err := json.Unmarshal(entry.Value, &d); err != nil {
		return nil, errwrap.Wrapf("failed to unmarshal destinati0on: {{err}}", err)
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"target_url":       d.TargetUrl,
			"send_entity_id":   d.SendEntityId,
			"timeout":          d.Timeout,
			"follow_redirects": d.FollowRedirects,
			"params":           d.Parameters,
			"metadata":         d.Metadata,
		},
	}, nil
}

func (b *backend) updateDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	return nil, fmt.Errorf("baby steps")

}
func (b *backend) deleteDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	return nil, fmt.Errorf("baby steps")

}
func (b *backend) destinationExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	return false, nil
}

// Sends an empty "ping" document to the destination and expects it to respond with a non-error HTTP return code.
func (b *backend) pingDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	return nil, fmt.Errorf("baby steps")
}

func (b *backend) contactDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Warn("Request: ", "path", req.Path, "params", data.Raw)

	sendRequest("http://localhost:8888/yes/", "{}", false)

	return nil, fmt.Errorf("baby steps")
}
