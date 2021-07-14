#!/bin/bash

if [ "$#" -ne 4 ]
then
  echo "Usage: gcloud_auth_helm_dispel.sh CHART_DIR APP_NAME TARGET_ENV TARGET_VER"
  exit 1
fi

CHART_DIR=$1
APP_NAME=$2
TARGET_ENV=$3
TARGET_VER=$4

if [[ ! -d ${CHART_DIR} ]]; then
    echo "${CHART_DIR} is not a directory."
    exit 1
fi

if [ "${TARGET_ENV}" != "staging" ] && [ "${TARGET_ENV}" != "prod" ] ; then
    echo "${TARGET_ENV} is not a valid target env."
    exit 1
fi

GOOGLE_APPLICATION_CREDENTIALS=/tmp/key.json

if [[ -z "${GOOGLE_SERVICE_KEY_BLOB}" ]]; then
  echo "GOOGLE_SERVICE_KEY_BLOB was not set."
  exit 1
else
  printf ${GOOGLE_SERVICE_KEY_BLOB} | base64 -d > ${GOOGLE_APPLICATION_CREDENTIALS}
fi

gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}

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

/dispel.sh ${APP_NAME} ${TARGET_ENV}

gcloud container images list-tags eu.gcr.io/${PROJECT_ID}/${APP_NAME} --filter='tags:*' --format='get(tags)' | while read -r tag ; do gcloud container images delete eu.gcr.io/${PROJECT_ID}/${APP_NAME}:$tag --force-delete-tags --quiet ; done
gcloud auth revoke --all
rm $GOOGLE_APPLICATION_CREDENTIALS