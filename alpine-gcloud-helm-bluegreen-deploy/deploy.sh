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

OFFLINE_COLOUR=$(kubectl get service/$TARGET_ENV-$APP_NAME-offline -o=jsonpath="{.spec.selector.colour}")
if [[ -z "${OFFLINE_COLOUR}" ]]; then
  OFFLINE_COLOUR="blue"
fi

echo "Deploying $TARGET_VER from $CHART_DIR to: $TARGET_ENV-$OFFLINE_COLOUR-$APP_NAME"
if helm upgrade $TARGET_ENV-$OFFLINE_COLOUR-$APP_NAME $CHART_DIR -f $VALUES --install --force --recreate-pods --wait --timeout=300 --set bluegreen.deployment.colour=$OFFLINE_COLOUR,bluegreen.deployment.version=$TARGET_VER; then
    echo "Successfully upgraded, switching colour to $OFFLINE_COLOUR"
    helm upgrade $TARGET_ENV-service-$APP_NAME $CHART_DIR -f $VALUES --install --force --wait --timeout=300 --set bluegreen.is_service_release=true,bluegreen.service.selector.colour=$OFFLINE_COLOUR
else
    LAST_GOOD_REV=$(helm history ${TARGET_ENV}-${OFFLINE_COLOUR}-${APP_NAME} -o json | jq -r -M --arg DEPLOYED "DEPLOYED" '[.[] | select(.status==$DEPLOYED)] | reverse | .[0] | .revision')
    
    int_reg='^[0-9]+$'
    if ! [[ $LAST_GOOD_REV =~ $int_reg ]] ; then
        echo "Failed upgrade, rolling back to 0"
        helm rollback --force --recreate-pods --wait --timeout=600 $TARGET_ENV-$OFFLINE_COLOUR-$APP_NAME 0
    else
        echo "Failed upgrade, rolling back to $LAST_GOOD_REV"
        helm rollback --force --recreate-pods --wait --timeout=600 $TARGET_ENV-$OFFLINE_COLOUR-$APP_NAME $LAST_GOOD_REV
    fi
    exit 1
fi