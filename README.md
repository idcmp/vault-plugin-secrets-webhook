# vault-plugin-secrets-webhook

## Lexicon

* A "destination" is the path in Vault which can be written to.
* A "target" is the remote service that Vault does the HTTP POST of the JSON document to.

## Purpose

HashiCorp Vault provides top notch AAA services. There are times where it would be handy to use the
authentication/authorization services it provides outside of Vault itself. The webhook plugin is a very
simple plugin that just sends HTTP requests to configured targets when a Vault user writes to a particular path
inside Vault. The HTTP request that Vault makes includes a JWS-signed JSON payload describing the request.

The target verifies the signature of the JWS document and then can perform whatever operation it needs to, knowing
that the request comes from an entity that has authenticated to Vault and has policies to write to the specific
destination.

## Developing
Run `source scripts/dev-init` before working on the plugin. Run `bash scripts/live-vault-test.sh` to spin Vault up.

This is still a Work In Progress. The `live-vault-test.sh` script sets up an example `webhook` secrets backend with
a `hello` destination.

## Configuring the Plugin

Adding the plugin to Vault is outside the scope of this document. Please see HashiCorp's official docs (or 
have a peek at `scripts/live-vault-test.sh`).

The plugin must be mounted at a certain path, and then needs to be supplied with a public/private key pair for
doing JSON document signing. You can use Vault's PKI backend, or you can use OpenSSL. 

```
vault secrets enable -path=webhook -plugin-name=webhook-plugin plugin
openssl genrsa -out "webhook.priv" 2048
openssl rsa -in "webhook.priv" -pubout >webhook.pub
vault write webhook/config/keys/jws certificate=@webhook.pub private_key=@webhook.priv
```

## Configuring a Destination


Destinations are created by writing to `webhook/config/destination/:name`. Anyone with write privileges to
`webhook/config/destination/*` can create a new destination. 

The following parameters are recognized:

* `target_url` is the URL Vault will POST its document to.

* `params` are parameters which are allowed to be forwarded from the user to the target endpoint. Defaults to empty.

* `metadata` are key value pairs passed to the target endpoint verbatim. Defaults to empty.

* `target_ca` can be set to a PEM-encoded value of the public CA certificate (this is what a call to Vault's `/pki/ca/pem` would return, for example). 
Note: If `target_ca` is set, it is the _only_ recognized CA. Defaults to unused.

* `send_entity_id` can be set to true if the target would find it useful to know identity information about the caller.
Vault  will send the entity ID in the payload. The target needs to request details about that entity ID by calling
Vault's identity API (`/identity/entity/id/:id`) was sufficient access. Defaults to false.

* `follow_redirects` can be set to true if you want Vault to follow redirects before posting its document. Note that
whatever Go's default HTTP client decides is best practices are used when following redirects. Defaults to false.

* `timeout` duration to allow the request to the target to run before bailing out. Defaults to 60s.

## Extra Security

When a target receives the signed JSON document, one of the fields is a nonce. The target can then call back
to Vault at `webhook/verify/:nonce` and it will receive the same payload. This endpoint will only exist as long
as the original HTTP call from Vault to the target is active. You can use this if you do not trust the JSON to be
signed properly or cannot reliably verify the JWS signature.

## Target Client

I have a separate project that is an example of a target client. It's not published yet (but if you want it, let me know).

Targets can request the public certificate by reading the `webhook/keys/jws/certificate` path and using the
`certificate` field. This path is available unauthenticated.

## TODO

* at least 30% test coverage
* ping support
* Support client-side SSL certificates
* Flesh out example target project.
* Better logging
* Example policies
