#!/usr/bin/env bash

if [ "$#" -ne 1 ]
then
  echo "Usage: enable_private_cluster_access.sh CIDR"
  exit 1
fi

valid_cidr_network() {
  local ip="${1%/*}"    # strip bits to leave ip address
  local bits="${1#*/}"  # strip ip address to leave bits
  local IFS=.; local -a a=($ip)

  # Sanity checks (only simple regexes)
  [[ $ip =~ ^[0-9]+(\.[0-9]+){3}$ ]] || return 1
  [[ $bits =~ ^[0-9]+$ ]] || return 1
  [[ $bits -gt 32 ]] || return 1

  # Create an array of 8-digit binary numbers from 0 to 255
  local -a binary=({0..1}{0..1}{0..1}{0..1}{0..1}{0..1}{0..1}{0..1})
  local binip=""

  # Test and append values of quads
  for quad in {0..3}; do
    [[ "${a[$quad]}" -gt 255 ]] && return 1
    printf -v binip '%s%s' "$binip" "${binary[${a[$quad]}]}"
  done

  # Fail if any bits are set in the host portion
  [[ ${binip:$bits} = *1* ]] && return 1

  return 0
}

CIDR=$1

if ! valid_cidr_network ${CIDR};
then
  echo "$CIDR Not a valid CIDR"
  exit 1
fi

GOOGLE_APPLICATION_CREDENTIALS=/tmp/key.json

if [[ -z "${GOOGLE_SERVICE_KEY_BLOB}" ]]; then
  echo "GOOGLE_SERVICE_KEY_BLOB was not set."
  exit 1
else
  printf ${GOOGLE_SERVICE_KEY_BLOB} | base64 -d > ${GOOGLE_APPLICATION_CREDENTIALS}
fi

gcloud auth activate-service-account $(cat ${GOOGLE_APPLICATION_CREDENTIALS} | jq -r -M \".client_email\") --key-file=${GOOGLE_APPLICATION_CREDENTIALS}

if [[ -z "${PROJECT_ID}" ]]; then
  echo "PROJECT_ID was not set."
  exit 1
else
  gcloud config set project ${PROJECT_ID}
fi

if [[ -z "${REGION}" ]]; then
  echo "REGION was not set."
  exit 1
else
  gcloud config set compute/region ${REGION}
fi

if [[ -z "${CLUSTER_ID}" ]]; then
  echo "CLUSTER_ID was not set."
  exit 1
else
  gcloud beta container clusters get-credentials --region ${REGION} ${CLUSTER_ID}
fi

AUTHORIZED_NETWORKS=$(gcloud container clusters describe ${CLUSTER_ID} --region=${REGION} --format=json | jq -r -M ".masterAuthorizedNetworksConfig.cidrBlocks" | grep "cidrBlock" | cut -d"\"" -f4 | paste -sd "," -)

if printf ${AUTHORIZED_NETWORKS} | grep -q ${CIDR};
then
    echo "Already Authorised!"
    exit 0
fi

gcloud container clusters update ${CLUSTER_ID} --region=${REGION} --enable-master-authorized-networks --master-authorized-networks="${AUTHORIZED_NETWORKS},${CIDR}"

gcloud auth revoke --all
rm ${GOOGLE_APPLICATION_CREDENTIALS}
exit 0