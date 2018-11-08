#!/usr/bin/env bash

if [ "$#" -lt 3 ] || [ "$#" -gt 4 ]
then
  echo "Usage: gcloud_auth_helm_deploy.sh CHART_DIR APP_NAME TARGET_ENV {OPTIONAL:TARGET_VER}"
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

if [[ -z "${TARGET_VER}" ]]; then
  echo "NON-VERSIONED DEPLOY!"
  echo "Deploying from ${CHART_DIR} to: ${TARGET_ENV}-${APP_NAME}"
  if helm upgrade ${TARGET_ENV}-${APP_NAME} ${CHART_DIR} -f ${CHART_DIR}/${TARGET_ENV}.yaml --install --force --recreate-pods --wait --timeout=120; then
      echo "Successfully upgraded"
  else
      LAST_GOOD_REV=$(helm history ${TARGET_ENV}-${APP_NAME} -o json | jq -r -M --arg DEPLOYED "DEPLOYED" '[.[] | select(.status==$DEPLOYED)] | reverse | .[0] | .revision')
      
      int_reg='^[0-9]+$'
      if ! [[ ${LAST_GOOD_REV} =~ $int_reg ]] ; then
          echo "Failed upgrade, rolling back to 0"
          helm rollback --force --recreate-pods --wait --timeout=600 ${TARGET_ENV}-${APP_NAME} 0
      else
          echo "Failed upgrade, rolling back to ${LAST_GOOD_REV}"
          helm rollback --force --recreate-pods --wait --timeout=600 ${TARGET_ENV}-${APP_NAME} ${LAST_GOOD_REV}
      fi
      exit 1
  fi
else
  /deploy.sh ${CHART_DIR} ${APP_NAME} ${TARGET_ENV} ${TARGET_VER}
fi

gcloud auth revoke --all
rm ${GOOGLE_APPLICATION_CREDENTIALS}