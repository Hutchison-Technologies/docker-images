#!/usr/bin/env bash

# Enable job ctrl so we can pull gcloud to foreground once we've populated it with import tool
set -m

# Config gcloud project
gcloud config set project acs-ent-staging-083511f1

# Start emulator and push to bg
gcloud --quiet beta emulators firestore start --host-port=0.0.0.0:8080 &

# Wait 10 seconds for emulator to spin up
sleep 10

# Import database backup to database
firestore-import -y --accountCredentials ./serviceAccountKey.json --backupFile ./databaseBackupFile.json

# Pull gcloud to foreground
fg %1