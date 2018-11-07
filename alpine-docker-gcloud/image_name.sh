#!/usr/bin/env bash

if [ "$#" -ne 2 ]
then
  echo "Usage: image_name.sh TARGET_ENV APP_NAME"
  exit 1
fi

TARGET_ENV=$1
APP_NAME=$2

SETUP_SUFFIX="-setup"
APP_NAME_SANS_SUFFIX=${APP_NAME%"$SETUP_SUFFIX"}

ONLINE_COLOUR=$(kubectl get service/${TARGET_ENV}-${APP_NAME_SANS_SUFFIX} -o=jsonpath="{.spec.selector.colour}")
if [[ -z "${ONLINE_COLOUR}" ]]; then
    echo "kubectl get service/${TARGET_ENV}-${APP_NAME_SANS_SUFFIX} -o=jsonpath=\"{.spec.selector.colour}\" - Returned no colour"
    exit 1
fi

ONLINE_DEPLOYMENT_IMAGE=$(kubectl get --ignore-not-found deployment/${TARGET_ENV}-${ONLINE_COLOUR}-${APP_NAME} -o=jsonpath="{.spec.template.spec.containers[0].image}")
ONLINE_JOB_IMAGE=$(kubectl get --ignore-not-found job/${TARGET_ENV}-${ONLINE_COLOUR}-${APP_NAME} -o=jsonpath="{.spec.template.spec.containers[0].image}")

if [[ -z "${ONLINE_DEPLOYMENT_IMAGE}" ]] && [[ -z "${ONLINE_JOB_IMAGE}" ]]; then
    echo "Cannot find live image for ${TARGET_ENV} ${APP_NAME}."
    exit 1
fi

FOUND_IMAGE="${ONLINE_DEPLOYMENT_IMAGE}" && [[ -z "${ONLINE_DEPLOYMENT_IMAGE}" ]] && FOUND_IMAGE="${ONLINE_JOB_IMAGE}"
echo $FOUND_IMAGE