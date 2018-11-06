#!/usr/bin/env bash

if [ "$#" -ne 1 ]
then
  echo "Usage: restrict_private_cluster_access.sh CIDR"
  exit 1
fi

CIDR=$1

if ! printf $CIDR | grep -Po '(\d+\.){3}\d+/\d+';
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

gcloud container clusters update ${CLUSTER_ID} --region=${REGION} --enable-master-authorized-networks --master-authorized-networks="${CIDR}"

gcloud auth revoke --all
rm ${GOOGLE_APPLICATION_CREDENTIALS}
exit 0