#!/usr/bin/env bash

killall -9 vault
killall -9 vault-plugin-secrets-webhook

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PLUGIN_DIR=$(dirname "$(realpath "${DIR}/../bin/vault-plugin-secrets-webhook")")

# The --dev-plugin-dir flag make break at any time. If it does, you will need to
# create a temporary vault configuration file setting the plugin_directory appropriately.
vault server -dev --dev-root-token-id=root-token --dev-plugin-dir="${PLUGIN_DIR}" &
sleep 1


vault secrets enable pki
vault write pki/root/generate/internal common_name="example.com"
curl http://127.0.0.1:8200/v1/pki/ca/pem >"${DIR}/pki_ca.pem"

SHA=$(shasum -a 256 "${PLUGIN_DIR}/vault-plugin-secrets-webhook" |awk '{print $1}')

vault write sys/plugins/catalog/webhook-plugin command=vault-plugin-secrets-webhook "sha_256=${SHA}"
vault secrets enable -path=webhook -plugin-name=webhook-plugin plugin
openssl genrsa -out "${DIR}/sample_key.priv" 2048
openssl rsa -in "${DIR}/sample_key.priv" -pubout >"${DIR}/sample_key.pub"
vault write webhook/config/keys/jws "certificate=@${DIR}/sample_key.pub" "private_key=@${DIR}/sample_key.priv"
vault read -field=certificate webhook/keys/jws/certificate


vault write webhook/config/destination/hello target_url=http://localhost:8888/ params=foo timeout=5s send_entity_id=true follow_redirects=false metadata=version=1 metadata=test=yes

vault read webhook/config/destination/hello
vault write webhook/destination/hello foo=bar




