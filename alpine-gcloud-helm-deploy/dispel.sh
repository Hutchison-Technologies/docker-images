#!/usr/bin/env bash

if [ "$#" -ne 2 ]
then
  echo "Usage: dispel.sh APP_NAME TARGET_ENV"
  exit 1
fi

APP_NAME=$1
TARGET_ENV=$2

echo "Dispelling $TARGET_ENV-$APP_NAME"
OFFLINE_DISPEL_RESULT=$(helm delete --purge $TARGET_ENV-$APP_NAME 2>&1)
echo "Dispel result: $OFFLINE_DISPEL_RESULT"

if [[ $OFFLINE_DISPEL_RESULT = *"not found"* ]] || [[ $OFFLINE_DISPEL_RESULT = "release \"$TARGET_ENV-$APP_NAME\" deleted" ]]; then
  echo "Successful dispel!"

  echo "Dispelling $TARGET_ENV-service-$APP_NAME"
  SERVICE_DISPEL_RESULT=$(helm delete --purge $TARGET_ENV-service-$APP_NAME 2>&1)
  echo "Dispel result: $SERVICE_DISPEL_RESULT"

  if [[ $SERVICE_DISPEL_RESULT = *"not found"* ]] || [[ $SERVICE_DISPEL_RESULT = "release \"$TARGET_ENV-service-$APP_NAME\" deleted" ]]; then
    echo "Successful dispel!"

    echo "Dispelling $TARGET_ENV-$APP_NAME"
    ONLINE_DISPEL_RESULT=$(helm delete --purge $TARGET_ENV-$APP_NAME 2>&1)
    echo "Dispel result: $ONLINE_DISPEL_RESULT"

    if [[ $ONLINE_DISPEL_RESULT = *"not found"* ]] || [[ $ONLINE_DISPEL_RESULT = "release \"$TARGET_ENV-$APP_NAME\" deleted" ]]; then
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