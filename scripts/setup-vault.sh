#!/usr/bin/env bash

killall -9 vault

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PLUGIN_DIR=$(dirname $(realpath "${DIR}/../bin/vault-plugin-secrets-relay"))

cat <<EOF >/tmp/vault.hcl
plugin_directory = "${PLUGIN_DIR}"
EOF

vault server -dev --dev-root-token-id=root-token --config=/tmp/vault.hcl &
sleep 1

SHA=$(shasum -a 256 "${PLUGIN_DIR}/vault-plugin-secrets-relay" |awk '{print $1}')

vault write sys/plugins/catalog/relay-plugin command=vault-plugin-secrets-relay sha_256=$SHA

vault secrets enable -path=relay -plugin-name=relay-plugin plugin

vault write relay/destination/hello foo=bar

