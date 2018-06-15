package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"crypto/x509"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-uuid"
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
			logical.UpdateOperation: b.pathContactDestination,
			logical.ReadOperation:   b.pathPingDestination,
		},

		//HelpSynopsis:    pathFetchHelpSyn,
		//HelpDescription: pathFetchHelpDesc,
	}
}

func pathConfigDestinations(b *backend) *framework.Path {

	return &framework.Path{
		Pattern: `config/destination/`,
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ListOperation: b.pathListDestinations,
		},
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
			logical.CreateOperation: b.pathWriteDestination,
			logical.ReadOperation:   b.pathReadDestination,
			logical.UpdateOperation: b.pathWriteDestination,
			logical.DeleteOperation: b.pathDeleteDestination,
		},
		ExistenceCheck: b.pathDestinationExistenceCheck,
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
	TargetCA        []byte            `yaml:"target_ca"`
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

	b.Logger().Warn("About to read target_ca")
	targetCA, err := getFieldValue("target_ca", data)
	if err != nil {
		return nil, err
	}
	if targetCA != "" {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(targetCA.([]byte)) {
			return nil, fmt.Errorf("could not parse \"target_ca\" certificate as PEM")
		}
	}
	b.Logger().Info("Destination created", "destination", d)
	return &d, nil

}

func (b *backend) pathWriteDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Warn("write destination", "path", req.Path)

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

func (b *backend) pathReadDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Warn("read", "path", req.Path)

	entry, _ := req.Storage.Get(ctx, req.Path)
	d, err := entryToDestination(entry)
	if err != nil {
		return nil, errwrap.Wrapf("failed to unmarshal destination: {{err}}", err)
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
			"target_ca":        d.TargetCA,
		},
	}, nil
}

func (b *backend) pathDeleteDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Warn("delete destination", "path", req.Path)

	err := req.Storage.Delete(ctx, req.Path)
	return nil, err
}

func (b *backend) pathDestinationExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	b.Logger().Warn("destination existence check", "path", req.Path)
	entry, err := req.Storage.Get(ctx, req.Path)
	return entry != nil, err
}

// Sends an empty "ping" document to the destination and expects it to respond with a non-error HTTP return code.
func (b *backend) pathPingDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Warn("hi", "path", req.Path)
	return nil, fmt.Errorf("baby steps")
}

func (b *backend) pathListDestinations(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	elements, err := req.Storage.List(ctx, req.Path)
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(elements), nil
}

func (b *backend) buildDocument(destination *Destination, req *logical.Request, data *framework.FieldData) (*Document, error) {
	// Build Document
	var document Document

	nonce, err := uuid.GenerateUUID()
	if err != nil {
		return nil, errwrap.Wrapf("failed to generate nonce: {{err}}", err)
	}

	document.Nonce = nonce
	document.Path = data.Get("target_name").(string)
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

	return &document, nil
}

func (b *backend) pathContactDestination(ctx context.Context, req *logical.Request, data *framework.FieldData) (response *logical.Response, retErr error) {
	b.Logger().Warn("Request: ", "path", req.Path, "params", data.Raw)

	b.Logger().Warn("contact destination", "path", req.Path)
	entry, _ := req.Storage.Get(ctx, filepath.ToSlash(filepath.Join("config", req.Path)))

	destination, err := entryToDestination(entry)
	if err != nil {
		return nil, errwrap.Wrapf("failed to unmarshal destination: {{err}}", err)
	}

	// TODO Should we cache this?
	storageEntry, err := req.Storage.Get(ctx, "config/keys/jws/private_key")

	if err != nil {
		return nil, errwrap.Wrapf("could not get jws private_key: {{err}}", err)
	}

	if storageEntry == nil {
		return nil, fmt.Errorf("incomplete cryptographic configuration, set jws keys")
	}

	document, err := b.buildDocument(destination, req, data)
	if err != nil {
		return nil, errwrap.Wrapf("could not build document: {{err}}", err)
	}

	bytesOut, err := serializeDocument(*document, storageEntry.Value)
	if err != nil {
		return nil, errwrap.Wrapf("could not marshal document: {{err}}", err)
	}

	verifyNonce := &logical.StorageEntry{
		Key:   "verify/" + document.Nonce,
		Value: bytesOut,
	}
	if err := req.Storage.Put(ctx, verifyNonce); err != nil {
		return nil, errwrap.Wrapf("could not store nonce verification: {{err}}", err)
	}

	b.Logger().Warn("writing to webhook/verify/" + document.Nonce)
	defer req.Storage.Delete(ctx, "verify/"+document.Nonce)

	bytesIn, err := sendRequest(destination.TargetURL, bytesOut, destination.FollowRedirects, destination.Timeout, destination.TargetCA)

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
