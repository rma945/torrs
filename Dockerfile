FROM --platform=$BUILDPLATFORM golang:alpine AS build

ARG TARGETOS
ARG TARGETARCH
ARG CGO_ENABLED

ENV CGO_ENABLED=${CGO_ENABLED:-0} \
    TARGETOS=${TARGETOS:-linux} \
    TARGETARCH=${TARGETARCH:-amd64}

WORKDIR /src/

COPY ./ /src/

RUN /src/build.sh

FROM scratch AS final
COPY --from=build /src/dist/torrs /torrs
USER 1000:1000

LABEL org.opencontainers.image.title="Torrs" \
    org.opencontainers.image.description="Torrs"

ENTRYPOINT ["/torrs"]
