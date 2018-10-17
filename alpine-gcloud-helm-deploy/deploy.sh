#!/usr/bin/env bash

if [ "$#" -ne 4 ]
then
  echo "Usage: deploy.sh CHART_DIR APP_NAME TARGET_ENV TARGET_VER"
  exit 1
fi

CHART_DIR=$1
APP_NAME=$2
TARGET_ENV=$3
TARGET_VER=$4
VALUES=$CHART_DIR/$TARGET_ENV.yaml


echo "Deploying $TARGET_VER from $CHART_DIR to: $TARGET_ENV-$APP_NAME"
if [[ -z "${ADDITIONAL_HELM_UPGRADE_SET_ARGS}" ]]; then
    echo "With additional set args $ADDITIONAL_HELM_UPGRADE_SET_ARGS"
fi
if helm upgrade $TARGET_ENV-$APP_NAME $CHART_DIR -f $VALUES --install --force --recreate-pods --wait --timeout=120 --set microservice.deployment.version=$TARGET_VER; then
    echo "Successfully upgraded"
else
    LAST_GOOD_REV=$(helm history ${TARGET_ENV}-${APP_NAME} -o json | jq -r -M --arg DEPLOYED "DEPLOYED" '[.[] | select(.status==$DEPLOYED)] | reverse | .[0] | .revision')
    
    int_reg='^[0-9]+$'
    if ! [[ $LAST_GOOD_REV =~ $int_reg ]] ; then
        echo "Failed upgrade, rolling back to 0"
        helm rollback --force --recreate-pods --wait --timeout=600 $TARGET_ENV-$APP_NAME 0
    else
        echo "Failed upgrade, rolling back to $LAST_GOOD_REV"
        helm rollback --force --recreate-pods --wait --timeout=600 $TARGET_ENV-$APP_NAME $LAST_GOOD_REV
    fi
    exit 1
fi