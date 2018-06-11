package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"path/filepath"
	"strings"
	"time"
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
			logical.UpdateOperation: b.pathContactDestination,
			logical.ReadOperation:   b.pathPingDestination,
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
				Type:        framework.TypeCommaStringSlice,
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
			logical.CreateOperation: b.pathCreateDestination,
			logical.ReadOperation:   b.readDestination,
			logical.UpdateOperation: b.updateDestination,
			logical.DeleteOperation: b.deleteDestination,
		},
		ExistenceCheck: b.destinationExistenceCheck,
		//HelpSynopsis:    pathFetchHelpSyn,
		//HelpDescription: pathFetchHelpDesc,
	}
}

// Destination contains all the operator specified configuration.
type Destination struct {
	TargetURL       string            `json:"target_url"`
	SendEntityID    bool              `json:"send_entity_id"`
	Timeout         time.Duration     `json:"timeout"`
	FollowRedirects bool              `json:"follow_redirects"`
	Parameters      []string          `json:"params"`
	Metadata        map[string]string `json:"metadata"`
}

// Parse fields from the user and create a Destination with sanitized input.
func (b *backend) createDestination(data *framework.FieldData) (*Destination, error) {
	var d Destination

	b.Logger().Info("Creating destination", "data", data)

	target, ok, err := data.GetOkErr("target_url")
	if !ok {
		return nil, fmt.Errorf("target_url is required")
	}
	if err != nil {
		return nil, errwrap.Wrapf("could not parse target_url: {{err}}", err)
	}
	d.TargetURL = target.(string)

	sendEntity, err := getFieldValue("send_entity_id", data)
	if err != nil {
		return nil, err
	}
	d.SendEntityID = sendEntity.(bool)

	timeout, err := getFieldValue("timeout", data)
	if err != nil {
		return nil, err
	}

	d.Timeout = time.Duration(timeout.(int)) * time.Second

	followRedirects, err := getFieldValue("follow_redirects", data)
	if err != nil {
		return nil, err
	}
	d.FollowRedirects = followRedirects.(bool)

	params, err := getFieldValue("params", data)
	if err != nil {
		return nil, err
	}

	for _, param := range params.([]string) {
		key := strings.ToLower(param)
		if !StrListContains(d.Parameters, key) {
			d.Parameters = append(d.Parameters, key)
		}
	}

	metadata, err := getFieldValue("metadata", data)
	if err != nil {
		return nil, err
	}
	d.Metadata = metadata.(map[string]string)

	b.Logger().Info("Destination created", "destination", d)
	return &d, nil

}

func (b *backend) pathCreateDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Warn("create", "path", req.Path)

	d, err := b.createDestination(data)
	if err != nil {
		return nil, errwrap.Wrapf("failed to create destination: {{err}}", err)
	}

	buf, err := json.Marshal(d)
	if err != nil {
		return nil, errwrap.Wrapf("failed to create destination: {{err}}", err)
	}
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
	b.Logger().Warn("read", "path", req.Path)

	entry, _ := req.Storage.Get(ctx, req.Path)
	d, err := entryToDestination(entry)
	if err != nil {
		return nil, errwrap.Wrapf("failed to unmarshal destinati0on: {{err}}", err)
	}

	timeout := fmt.Sprintf("%v", d.Timeout)

	return &logical.Response{
		Data: map[string]interface{}{
			"target_url":       d.TargetURL,
			"send_entity_id":   d.SendEntityID,
			"timeout":          timeout,
			"follow_redirects": d.FollowRedirects,
			"params":           d.Parameters,
			"metadata":         d.Metadata,
		},
	}, nil
}

func (b *backend) updateDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Warn("hi", "path", req.Path)
	return nil, fmt.Errorf("baby steps")

}
func (b *backend) deleteDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Warn("hi", "path", req.Path)
	return nil, fmt.Errorf("baby steps")

}
func (b *backend) destinationExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	b.Logger().Warn("hi", "path", req.Path)
	return false, nil
}

// Sends an empty "ping" document to the destination and expects it to respond with a non-error HTTP return code.
func (b *backend) pathPingDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Warn("hi", "path", req.Path)
	return nil, fmt.Errorf("baby steps")
}

type Document struct {
	Nonce      string            `json:"nonce"`
	Path       string            `json:"path"`
	Timestamp  int64             `json:"timestamp"`
	RequestID  string            `json:"request_id"`
	EntityID   string            `json:"entity_id,omitempty"`
	Parameters map[string]string `json:"params,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

func (b *backend) pathContactDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Warn("Request: ", "path", req.Path, "params", data.Raw)

	b.Logger().Warn("contact destination", "path", req.Path)
	entry, _ := req.Storage.Get(ctx, filepath.ToSlash(filepath.Join("config", req.Path)))

	destination, err := entryToDestination(entry)
	if err != nil {
		return nil, errwrap.Wrapf("failed to unmarshal destination: {{err}}", err)
	}

	nonce, err := uuid.GenerateUUID()
	if err != nil {
		return nil, errwrap.Wrapf("failed to generate nonce: {{err}}", err)
	}

	// Build Document
	var document Document

	document.Nonce = nonce
	document.Path = data.Raw["target_name"].(string)
	if destination.SendEntityID {
		document.EntityID = req.EntityID
	}
	document.Timestamp = time.Now().Unix()
	document.RequestID = req.ID
	if destination.Metadata != nil {
		document.Metadata = destination.Metadata
	}
	document.Parameters = make(map[string]string)

	for k, v := range data.Raw {
		lowKey := strings.ToLower(k)
		if StrListContains(destination.Parameters, lowKey) {
			document.Parameters[lowKey] = v.(string)
		}
	}

	bytesOut, err := json.Marshal(document)

	if err != nil {
		return nil, errwrap.Wrapf("could not marshal document: {{err}}", err)
	}

	bytesIn, err := sendRequest(destination.TargetURL, bytesOut, destination.FollowRedirects, destination.Timeout)

	if err != nil {
		return nil, errwrap.Wrapf("could not process request: {{err}}", err)
	}

	b.Logger().Warn("response", "body", string(bytesIn))

	return &logical.Response{
		Data: map[string]interface{}{
			"response": string(bytesIn),
		},
	}, nil
}
