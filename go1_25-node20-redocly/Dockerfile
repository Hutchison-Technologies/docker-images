# Install node base image with specific version (for redocly)
FROM node:20.19.4-alpine AS node

# Install Go base image for build machine
FROM golang:1.25.7-alpine3.23 AS build

# Copy Node + npm from the node stage to Go image
COPY --from=node /usr/local /usr/local

# Install base packages
RUN apk add --no-cache curl tar xz git libstdc++ libgcc

# Verify node and go versions
RUN node -v && npm -v && go version