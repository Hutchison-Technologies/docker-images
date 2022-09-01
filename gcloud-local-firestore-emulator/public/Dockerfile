FROM google/cloud-sdk:alpine as base

#JDK 8 required for emulator
RUN apk add openjdk8-jre

FROM base as withjdk 

# NPM required to load data into database
RUN apk --update add npm
RUN npm install -g node-firestore-import-export

# Install beta commands (emulators are a beta feature)
RUN gcloud components install beta cloud-firestore-emulator

FROM withjdk
ENV PORT 8080
ENV FIRESTORE_EMULATOR_HOST localhost:8080

COPY populatefirestore.bash .
COPY databaseBackupFile.json .
COPY serviceAccountKey.json .

ENTRYPOINT ["./populatefirestore.bash"]