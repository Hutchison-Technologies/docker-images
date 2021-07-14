#!/bin/bash

if [ "$#" -ne 6 ]
then
  echo "Usage: gcloud_auth.sh CHART_DIR TARGET_ENV PROJECT_ID REGION CLUSTER_ID CHARTMUSEUM"
  exit 1
fi

CHART_DIR=$1
TARGET_ENV=$2
PROJECT_ID=$3
REGION=$4
CLUSTER_ID=$5
CHARTMUSEUM=$6

if [[ -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]]; then
  GOOGLE_APPLICATION_CREDENTIALS=/tmp/key.json

  if [[ -z "${GOOGLE_SERVICE_KEY_BLOB}" ]]; then
    echo "GOOGLE_SERVICE_KEY_BLOB was not set."
    exit 1
  else
    printf ${GOOGLE_SERVICE_KEY_BLOB} | base64 -d > ${GOOGLE_APPLICATION_CREDENTIALS}
  fi

  gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}
else
  gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}
fi

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

helm init --wait --service-account tiller --client-only

if [[ -z "${CHARTMUSEUM}" ]]; then
  echo "CHARTMUSEUM was not set."
  exit 1
else
  helm repo add chartmuseum ${CHARTMUSEUM}
fi

helm dep update ${CHART_DIR} 
helm dep build ${CHART_DIR}

if ! helm lint -f ${CHART_DIR}/${TARGET_ENV}.yaml ${CHART_DIR}; then
  echo "${CHART_DIR} contains malformed chart."
  exit 1
fi