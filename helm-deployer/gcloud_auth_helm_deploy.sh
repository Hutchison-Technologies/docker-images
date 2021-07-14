#!/bin/bash

if [ "$#" -ne 5 ]
then
  echo "Usage: gcloud_auth_helm_deploy.sh DEPLOY_TYPE CHART_DIR APP_NAME TARGET_ENV TARGET_VER"
  exit 1
fi

DEPLOY_TYPE=$1
CHART_DIR=$2
APP_NAME=$3
TARGET_ENV=$4
TARGET_VER=$5

if [[ ! -d ${CHART_DIR} ]]; then
    echo "${CHART_DIR} is not a directory."
    exit 1
fi

if [ "${TARGET_ENV}" != "staging" ] && [ "${TARGET_ENV}" != "prod" ] ; then
    echo "${TARGET_ENV} is not a valid target env."
    exit 1
fi

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

helm-deployer ${DEPLOY_TYPE} -chart-dir ${CHART_DIR} -app-name ${APP_NAME} -target-env ${TARGET_ENV} -app-version ${TARGET_VER}