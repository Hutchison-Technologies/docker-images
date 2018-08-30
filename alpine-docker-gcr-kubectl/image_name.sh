#!/usr/bin/env bash

if [ "$#" -ne 2 ]
then
  echo "Usage: image_name.sh TARGET_ENV APP_NAME"
  exit 1
fi

TARGET_ENV=$1
APP_NAME=$2

ONLINE_COLOUR=$(kubectl get service/$TARGET_ENV-$APP_NAME -o=jsonpath="{.spec.selector.colour}")
if [[ -z "${ONLINE_COLOUR}" ]]; then
    echo "kubectl get service/$TARGET_ENV-$APP_NAME -o=jsonpath=\"{.spec.selector.colour}\" - Returned no colour"
    exit 1
fi

ONLINE_IMAGE=$(kubectl get deployment/$TARGET_ENV-$ONLINE_COLOUR-$APP_NAME -o=jsonpath="{.spec.template.spec.containers[0].image}")
if [[ -z "${ONLINE_IMAGE}" ]]; then
    echo "kubectl get deployment/$TARGET_ENV-$ONLINE_COLOUR-$APP_NAME -o=jsonpath=\"{.spec.template.spec.containers[0].image}\" - Returned no image"
    exit 1
fi

echo $ONLINE_IMAGE