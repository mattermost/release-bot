# This dockerfile is used to build Mattermost Release Bot.
# A multi stage build, with golang used as a builder
# and gcr.io/distroless/static-debian11 as runner
ARG GO_IMAGE=golang:1.18@sha256:90c06f42c1aa2b6b96441c0e6192aff48815cf5e7950cd661ed316fdbfb06ed4
# hadolint ignore=DL3006
FROM ${GO_IMAGE} as builder

#GO_BUILD_PLATFORMS holds the platforms that we will build the docker image against
ARG GO_BUILD_PLATFORMS=linux-amd64

# Setup directories structure and compile
COPY . /src
WORKDIR /src
RUN make go-build

# Cause of bash plugins we need to execute bash and shell commands.
FROM gcr.io/distroless/static-debian11@sha256:35817ff4e3441ebfa430cfca98121e43cbf4a134412285c2460a9464c97ac2ca as runner
COPY --from=builder /src/dist/release_bot-linux-amd64 /opt/release-bot/bin/release-bot

# We should refrain from running as privileged user
# Run as UID for nobody
USER 65534

WORKDIR /opt/release-bot
ENTRYPOINT ["./bin/release-bot"]
