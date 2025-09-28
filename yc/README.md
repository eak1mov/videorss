# YC

## folder
```
# https://yandex.cloud/ru/docs/resource-manager/cli-ref/folder/create

yc resource-manager folder create \
  --name myapp2-folder

export YC_FOLDER_ID=`yc resource-manager folder get --name myapp2-folder --jq ".id"`

```


## service account
```
# https://yandex.cloud/ru/docs/cli/cli-ref/iam/cli-ref/service-account/create
# https://yandex.cloud/ru/docs/cli/cli-ref/iam/cli-ref/service-account/get

yc iam service-account create \
  --folder-name myapp2-folder \
  --name myapp2-account

yc iam service-account get \
  --folder-name myapp2-folder \
  --name myapp2-account

export YC_SERVICE_ACCOUNT_ID=`yc iam service-account get --folder-name myapp2-folder --name myapp2-account --jq ".id"`

```


## logging
```
# https://yandex.cloud/ru/docs/logging/operations/create-group
# https://yandex.cloud/ru/docs/logging/security/
# https://yandex.cloud/ru/docs/logging/cli-ref/group/create
# https://yandex.cloud/ru/docs/logging/cli-ref/group/add-access-binding

yc logging group create \
  --folder-name myapp2-folder \
  --name myapp2-logging \
  --retention-period 168h

yc logging group add-access-binding \
  --folder-name myapp2-folder \
  --name myapp2-logging \
  --role logging.writer \
  --service-account-name myapp2-account

export YC_LOGGING_GROUP_ID=`yc logging group get --folder-name myapp2-folder --name myapp2-logging --jq ".id"`

```


## monitoring
```
# https://yandex.cloud/ru/docs/monitoring/security/
# https://yandex.cloud/ru/docs/iam/operations/sa/assign-role-for-sa
# https://yandex.cloud/ru/docs/resource-manager/cli-ref/folder/add-access-binding

yc resource-manager folder add-access-binding \
  --folder-name myapp2-folder \
  --name myapp2-folder \
  --role monitoring.editor \
  --service-account-name myapp2-account

```


## storage
```
# https://yandex.cloud/ru/docs/storage/operations/buckets/create
# https://yandex.cloud/ru/docs/storage/operations/buckets/edit-acl
# https://yandex.cloud/ru/docs/storage/concepts/acl
# https://yandex.cloud/ru/docs/storage/security/
# https://yandex.cloud/ru/docs/storage/security/acl
# https://yandex.cloud/ru/docs/iam/operations/authentication/manage-access-keys
# https://yandex.cloud/ru/docs/cli/cli-ref/storage/cli-ref/bucket/create
# https://yandex.cloud/ru/docs/cli/cli-ref/storage/cli-ref/bucket/update
# https://yandex.cloud/ru/docs/cli/cli-ref/iam/cli-ref/access-key/create

yc storage bucket create \
  --folder-name myapp2-folder \
  --name myapp2-bucket \
  --max-size 1 \
  --grants grant-type=grant-type-account,grantee-id=${YC_SERVICE_ACCOUNT_ID},permission=permission-read \
  --grants grant-type=grant-type-account,grantee-id=${YC_SERVICE_ACCOUNT_ID},permission=permission-write

yc storage bucket update \
  --folder-name myapp2-folder \
  --name myapp2-bucket \
  --grants grant-type=grant-type-account,grantee-id=${YC_SERVICE_ACCOUNT_ID},permission=permission-read \
  --grants grant-type=grant-type-account,grantee-id=${YC_SERVICE_ACCOUNT_ID},permission=permission-write

yc storage bucket get \
  --folder-name myapp2-folder \
  --name myapp2-bucket \
  --with-acl

yc iam access-key create \
  --folder-name myapp2-folder \
  --service-account-name myapp2-account

export YC_STORAGE_ACCESS_KEY=`yc iam access-key create --folder-name myapp2-folder --service-account-name myapp2-account --format json`

export AWS_ACCESS_KEY_ID=`echo "${YC_STORAGE_ACCESS_KEY}" | jq -r ".access_key.key_id"`
export AWS_SECRET_ACCESS_KEY=`echo "${YC_STORAGE_ACCESS_KEY}" | jq -r ".secret"`

```


## secrets
```
# https://yandex.cloud/ru/docs/lockbox/operations/secret-create
# https://yandex.cloud/ru/docs/lockbox/cli-ref/secret/create
# https://yandex.cloud/ru/docs/lockbox/cli-ref/secret/add-access-binding
# https://yandex.cloud/ru/docs/lockbox/cli-ref/secret/add-version
# https://yandex.cloud/ru/docs/lockbox/cli-ref/secret/schedule-version-destruction

yc lockbox secret create \
  --folder-name myapp2-folder \
  --name myapp2-secret \
  --payload "[{'key': 'TEST_KEY', 'text_value': 'TEST_VALUE'}]"

yc lockbox secret add-access-binding \
  --folder-name myapp2-folder \
  --name myapp2-secret \
  --role lockbox.payloadViewer \
  --service-account-name myapp2-account

export YC_SECRET_PAYLOAD="[\
{'key': 'VK_API_TOKEN', 'text_value': '$VK_API_TOKEN'},\
{'key': 'AWS_ACCESS_KEY_ID', 'text_value': '$AWS_ACCESS_KEY_ID'},\
{'key': 'AWS_SECRET_ACCESS_KEY', 'text_value': '$AWS_SECRET_ACCESS_KEY'},\
{'key': 'SETTINGS_PASSWORD', 'text_value': '$SETTINGS_PASSWORD'},\
]"

yc lockbox secret add-version \
  --folder-name myapp2-folder \
  --name myapp2-secret \
  --payload "${YC_SECRET_PAYLOAD}"

export YC_SECRET_ID=`yc lockbox secret get --folder-name myapp2-folder --name myapp2-secret --jq ".id"`

```


## network
```
# https://yandex.cloud/ru/docs/vpc/concepts/network
# https://yandex.cloud/ru/docs/vpc/concepts/security-groups
# https://yandex.cloud/ru/docs/vpc/operations/subnet-create
# https://yandex.cloud/ru/docs/vpc/operations/security-group-create
# https://yandex.cloud/ru/docs/vpc/cli-ref/network/create
# https://yandex.cloud/ru/docs/vpc/cli-ref/subnet/create
# https://yandex.cloud/ru/docs/vpc/cli-ref/security-group/create

yc vpc network create \
  --folder-name myapp2-folder \
  --name myapp2-network

yc vpc subnet create \
  --folder-name myapp2-folder \
  --name myapp2-subnet-ru-central1-a \
  --network-name myapp2-network \
  --zone ru-central1-a \
  --range 10.128.0.0/24

yc vpc security-group create \
  --folder-name myapp2-folder \
  --name myapp2-security-group \
  --network-name myapp2-network \
  --rule "description=ssh,direction=ingress,port=22,protocol=tcp,v4-cidrs=[0.0.0.0/0]" \
  --rule "description=http,direction=ingress,port=80,protocol=tcp,v4-cidrs=[0.0.0.0/0]" \
  --rule "description=https,direction=ingress,port=443,protocol=tcp,v4-cidrs=[0.0.0.0/0]" \
  --rule "description=any,direction=egress,from-port=0,to-port=65535,protocol=tcp,v4-cidrs=[0.0.0.0/0]"

export YC_SECURITY_GROUP_ID=`yc vpc security-group get --folder-name myapp2-folder --name myapp2-security-group --jq ".id"`

```


## container registry
```
# https://yandex.cloud/ru/docs/container-registry/quickstart/
# https://yandex.cloud/ru/docs/container-registry/operations/authentication
# https://yandex.cloud/ru/docs/container-registry/cli-ref/registry/create
# https://yandex.cloud/ru/docs/container-registry/cli-ref/registry/add-access-binding

yc container registry create \
  --folder-name myapp2-folder \
  --name myapp2-registry \
  --secure

yc container registry add-access-binding \
  --folder-name myapp2-folder \
  --name myapp2-registry \
  --role container-registry.images.puller \
  --service-account-name myapp2-account

export YC_REGISTRY_ID=`yc container registry get --folder-name myapp2-folder --name myapp2-registry --jq ".id"`

yc iam create-token | docker login --username iam --password-stdin cr.yandex

docker build . -t cr.yandex/$YC_REGISTRY_ID/myapp_image:v06
docker push cr.yandex/$YC_REGISTRY_ID/myapp_image:v06

docker build . -t cr.yandex/$YC_REGISTRY_ID/secrets_image:v02
docker push cr.yandex/$YC_REGISTRY_ID/secrets_image:v02

docker build . -t cr.yandex/$YC_REGISTRY_ID/fluentbit_healthcheck_image:v02
docker push cr.yandex/$YC_REGISTRY_ID/fluentbit_healthcheck_image:v02

docker build . -t cr.yandex/$YC_REGISTRY_ID/fluentbit_image:v01
docker push cr.yandex/$YC_REGISTRY_ID/fluentbit_image:v01

docker build . -t cr.yandex/$YC_REGISTRY_ID/unified_agent_image:v03
docker push cr.yandex/$YC_REGISTRY_ID/unified_agent_image:v03

docker build . -t cr.yandex/$YC_REGISTRY_ID/nginx_image:v02
docker push cr.yandex/$YC_REGISTRY_ID/nginx_image:v02

```


## compute
```
# https://yandex.cloud/ru/docs/compute/concepts/vm
# https://yandex.cloud/ru/docs/compute/concepts/network
# https://yandex.cloud/ru/docs/compute/concepts/vm-platforms
# https://yandex.cloud/ru/docs/compute/concepts/performance-levels
# https://yandex.cloud/ru/docs/compute/pricing
# https://yandex.cloud/ru/docs/cli/cli-ref/compute/cli-ref/image/get-latest-from-family
# https://yandex.cloud/ru/docs/cli/cli-ref/compute/cli-ref/instance/create-with-container
# https://yandex.cloud/ru/docs/cli/cli-ref/compute/cli-ref/instance/create

ssh-keygen -t ed25519 -f ~/.ssh/id_myapp2 -C myapp2

export YC_REGISTRY_ID=`yc container registry get --folder-name myapp2-folder --name myapp2-registry --jq ".id"`
export YC_FOLDER_ID=`yc resource-manager folder get --name myapp2-folder --jq ".id"`
export YC_SECRET_ID=`yc lockbox secret get --folder-name myapp2-folder --name myapp2-secret --jq ".id"`
export YC_LOGGING_GROUP_ID=`yc logging group get --folder-name myapp2-folder --name myapp2-logging --jq ".id"`
export YC_SECURITY_GROUP_ID=`yc vpc security-group get --folder-name myapp2-folder --name myapp2-security-group --jq ".id"`

echo $YC_REGISTRY_ID
echo $YC_FOLDER_ID
echo $YC_SECRET_ID
echo $YC_LOGGING_GROUP_ID
echo $YC_SECURITY_GROUP_ID

yc compute instance create \
  --folder-name myapp2-folder \
  --name myapp6-vm \
  --zone=ru-central1-a \
  --metadata-from-file docker-compose=compose.yaml \
  --ssh-key ~/.ssh/id_myapp2.pub \
  --create-boot-disk type=network-hdd,size=15G,image-family=container-optimized-image,image-folder-id=standard-images \
  --network-interface subnet-name=myapp2-subnet-ru-central1-a,nat-ip-version=ipv4,security-group-ids=$YC_SECURITY_GROUP_ID \
  --memory 1G \
  --cores 2 \
  --core-fraction 20 \
  --platform standard-v3 \
  --maintenance-policy restart \
  --service-account-name myapp2-account

```


## local tests
```
# https://yandex.cloud/ru/docs/compute/cli-ref/ssh/

yc compute ssh --folder-name myapp2-folder --name myapp6-vm --identity-file ~/.ssh/id_myapp2 --login yc-user

docker build --tag myapp20 .

curl -v "http://localhost:8080/vk/test1"

export YC_TOKEN=`yc iam create-token`

```
