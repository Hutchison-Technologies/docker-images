# Alpine Gcloud Firestore Emulator Docker File

This image is to be used as a base for running a firestore instance locally. Import this image as a docker base, transfer a service account key and backup.json file to your new image. This will allow the firestore import tool to pre-populate your firestore database with the given backup image. To generate a backup.json file, see below. An example of a docker image using this as a base can be found at [Link to test service].

# Automatically populating via backup

This image includes a `firestore-start.bash` script which will initialise the firestore emulator instance, and populate it using the import tool if you set it as the entrypoint point for your docker image e.g. `ENTRYPOINT [ "./populatefirestore.bash" ]`. If you do this, you'll need to provide a service account key file and a backup.json file to populate the database with.

If you're using this entry point, you'll need to copy a `databaseFile.json` and a `serviceAccountKey.json` into your docker file.

eg:

```
COPY serviceAccountKey.json .
COPY databaseFile.json .
ENTRYPOINT ["./populatefirestore.bash"]
```

To obtain these files, follow the steps below.

# Create a backup.json of your live Firestore

If you have NPM installed locally, you can install the import/export package via `npm install -g node-firestore-import-export`

Once this is complete, you'll need to have get hold of a service account key from your gcp instance. You'll find this under IAM -> Service Account -> Keys.
Select the Default Compute Engine service account. Generate a new one, and download it as json.

Rename it to `serviceAccountKey.json` and place it in the root of this folder.

You can now run `firestore-export --accountCredentials serviceAccountKey.json --backupFile databaseBackupFile.json`. A `databaseBackupFile.json` file will be generated and can be used by `firestore-import`.

Don't run firestore import on your local system, this should only be done in the firestore. The system knows which GCP project to pull from automagically, as it utilises the information in the service account key.

# Running

Once you've followed the steps above, you can run the docker image. Please ensure that you pass the `-p 8080:8080` parameter into the run parameters to map the emulator port to your local 0.0.0.0:8080 port.

Additionally, you may find you can't kill the docker image when it's running. The work around is to create an interactive terminal using `-it`.

Thus, the full command set:

- `docker build . -t firestore-emulator:latest` (to tag it as firestore-emulator:latest)
- `docker run -it -p 8080:8080 firestore-emulator:latest` (assuming you tagged it with firestore-emulator:latest)

To connect test against this firestore instance, we need to set an environment variable `FIRESTORE_EMULATOR_HOST`.

The docker image will be on `0.0.0.0:8080` if you've followed the above steps. Set the `FIRESTORE_EMULATOR_HOST` variable to this value:

`FIRESTORE_EMULATOR_HOST=localhost:8080`.

You can also set this in your integration test suite as a os.SetEnv("FIRESTORE_EMULATOR_HOST", "localhost:8080") call if you're interacting with the database using GoLang.

eg:

```
  ctx := context.Background()

  newFirestoreClient = firestore.NewClient
  os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")

  // Run
  client, err := FirestoreInit(ctx, "acs-ent-staging-[PROJ-ID-NUM]")
  if err != nil {
    fmt.Printf("\nERR: %s", err)
  }

  defer client.Close()
```
