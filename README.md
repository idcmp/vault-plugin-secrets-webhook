# vault-plugin-secrets-webhook

## Purpose

HashiCorp Vault provides top notch AAA services. There are times where it would be handy to use the
authentication/authorization services it provides outside of Vault itself. The webhook plugin is a very
simple plugin that just sends HTTP requests to configured targets when a Vault user writes to a particular path
inside Vault. The HTTP request that Vault makes includes a JWS-signed JSON payload describing the request.

The target verifies the signature of the JWS document and then can perform whatever operation it needs to, knowing
that the request comes from an entity that has authenticated to Vault and has policies to write to the specific
destination.

## Example

Sometimes there are users or services which need to perform activities inside Vault, such as enrolling an AMI.
However, as Vault administrators, we don't want these users to enroll just any AMI into just any policy. We have
a trusted "AMIBot" that runs with heightened Vault privileges. It's configured to map certain project names to
specific Vault roles.

Our CI server has been granted write access to the `webhook/enroll-ami` destination inside Vault. The CI server writes
a request including the project name and the AMI ID. This request is then sent to Vault, and forwarded to the AMIBot
which then performs the enrolling of the AMI on behalf of the user. This gives the CI server the ability to enroll 
AMIs, but limits the mapping of which policies are assigned to which projects.

## Developing
Run `source scripts/dev-init` before working on the plugin. Run `bash scripts/live-vault-test.sh` to spin Vault up.

## Configuring

This is still a Work In Progress. The `live-vault-test.sh` script sets up an example `webhook` secrets backend with
a `hello` destination.

When creating a destination, `params` are parameters which are allowed to be forwarded from the user to the target
endpoint and `metadata` is passed to the target endpoint verbatim. 

## Target Client

I have a separate project that is an example of a target client. It's not published yet (but if you want it, let me know).

## TODO

* Support client-side SSL certificates
* RWLock on configuration
* Flesh out example target project.
* Better logging
* Example policies
* ping support
