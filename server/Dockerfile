FROM golang:1.26 as build

WORKDIR /src
COPY go.mod ./
COPY go.sum ./
COPY .. ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server /src/cmd/server
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /hub /src/cmd/hub

FROM gcr.io/distroless/base:debug-nonroot
WORKDIR /

COPY --from=build /server /server
COPY --from=build /hub /hub

ENTRYPOINT ["/server"]
