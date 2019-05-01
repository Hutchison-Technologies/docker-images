#!/usr/bin/env bash

if [ "$#" -ne 2 ]
then
  echo "Usage: dispel.sh APP_NAME TARGET_ENV"
  exit 1
fi

APP_NAME=$1
TARGET_ENV=$2

OFFLINE_COLOUR=$(kubectl get service/$TARGET_ENV-$APP_NAME-offline -o=jsonpath="{.spec.selector.colour}")
if [[ -z "${OFFLINE_COLOUR}" ]]; then
  OFFLINE_COLOUR="blue"
fi

ONLINE_COLOUR=$(kubectl get service/$TARGET_ENV-$APP_NAME -o=jsonpath="{.spec.selector.colour}")
if [[ -z "${ONLINE_COLOUR}" ]]; then
  ONLINE_COLOUR="green"
fi

echo "Dispelling $TARGET_ENV-$OFFLINE_COLOUR-$APP_NAME"
OFFLINE_DISPEL_RESULT=$(helm delete --purge $TARGET_ENV-$OFFLINE_COLOUR-$APP_NAME 2>&1)
echo "Dispel result: $OFFLINE_DISPEL_RESULT"

if [[ $OFFLINE_DISPEL_RESULT = *"not found"* ]] || [[ $OFFLINE_DISPEL_RESULT = "release \"$TARGET_ENV-$OFFLINE_COLOUR-$APP_NAME\" deleted" ]]; then
  echo "Successful dispel!"

  echo "Dispelling $TARGET_ENV-service-$APP_NAME"
  SERVICE_DISPEL_RESULT=$(helm delete --purge $TARGET_ENV-service-$APP_NAME 2>&1)
  echo "Dispel result: $SERVICE_DISPEL_RESULT"

  if [[ $SERVICE_DISPEL_RESULT = *"not found"* ]] || [[ $SERVICE_DISPEL_RESULT = "release \"$TARGET_ENV-service-$APP_NAME\" deleted" ]]; then
    echo "Successful dispel!"

    echo "Dispelling $TARGET_ENV-$ONLINE_COLOUR-$APP_NAME"
    ONLINE_DISPEL_RESULT=$(helm delete --purge $TARGET_ENV-$ONLINE_COLOUR-$APP_NAME 2>&1)
    echo "Dispel result: $ONLINE_DISPEL_RESULT"

    if [[ $ONLINE_DISPEL_RESULT = *"not found"* ]] || [[ $ONLINE_DISPEL_RESULT = "release \"$TARGET_ENV-$ONLINE_COLOUR-$APP_NAME\" deleted" ]]; then
      echo "Successful dispel!"
    else
      echo "Failed to dispel."
      exit 1
    fi
  else
    echo "Failed to dispel."
    exit 1
  fi
else
  echo "Failed to dispel."
  exit 1
fi