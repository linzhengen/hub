FROM golang:1.25 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY ./server ./server
COPY ./ui/web/dist ./ui/web/dist
COPY ./ui/web/embed.go ./ui/web/embed.go

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /server /src/server/cmd/server

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /hub /src/server/cmd/cli

FROM gcr.io/distroless/base:debug-nonroot
WORKDIR /

COPY --from=build /server /server
COPY --from=build /hub /hub

ENTRYPOINT ["/server"]
