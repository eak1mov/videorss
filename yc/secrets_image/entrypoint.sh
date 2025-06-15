#!/bin/sh

set -e

test -n "${YC_FOLDER_ID}"
test -n "${YC_SECRET_ID}"
test -n "${SECRET_KEYS}"
test -n "${SECRETS_PATH}"
echo "environment: ok"

# test -n "${YC_TOKEN}"
YC_TOKEN=$(curl -sf -H "Metadata-Flavor:Google" "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token" | jq -r ".access_token")
echo "token: ok"

SECRET_VALUES=$(curl -sf -H "Authorization: Bearer ${YC_TOKEN}" "https://payload.lockbox.api.cloud.yandex.net/lockbox/v1/secrets/${YC_SECRET_ID}/payload")
echo "secret payload: ok"

for SECRET_KEY in $SECRET_KEYS; do
    SECRET_VALUE=$(echo -n "${SECRET_VALUES}" | jq -r ".entries[] | select(.key == \"${SECRET_KEY}\") | .textValue")
    echo -n "${SECRET_VALUE}" > "${SECRETS_PATH}/${SECRET_KEY}"
    echo "${SECRET_KEY}: ok"
done
