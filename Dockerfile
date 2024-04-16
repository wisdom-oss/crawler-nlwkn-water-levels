FROM docker.io/golang:alpine AS build-service
COPY . /tmp/src
WORKDIR /tmp/src
RUN mkdir -p /tmp/build
RUN go mod download
RUN go build -o /tmp/build/app

FROM docker.io/alpine:latest
COPY --from=build-service /tmp/build/app /service
COPY resources/* /
ENTRYPOINT ["/service"]
ARG GH_REPO=unset
ARG GH_VERSION=unset
LABEL org.opencontainers.image.source=https://github.com/$GH_REPO
LABEL org.opencontainers.image.version=$GH_VERSION
EXPOSE 8000
HEALTHCHECK --interval=30s --timeout=15s CMD /service -healthcheck