# Alpine + PGBouncer + c-ares + udns

This image is based on the [alpine:3.8](https://hub.docker.com/_/alpine/) image. On top of this, it has [pgbouncer](https://pgbouncer.github.io/) installed, this version also uses c-ares and udns for support of asynchronous DNS requests, which is required when resolving hostnames that are passed into a container as environment variables.

This image is useful for connection pooling to a PostgresSQL instance.
