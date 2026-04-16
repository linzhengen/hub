FROM golang:1.25 as build

WORKDIR /src
COPY go.mod ./
COPY go.sum ./
COPY ./server ./server
COPY ./ui/web/dist ./ui/web/dist
COPY ./ui/web/embed.go ./ui/web/embed.go
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server /src/server/cmd/server
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /hub /src/server/cmd/hub

FROM gcr.io/distroless/base:debug-nonroot
WORKDIR /

COPY --from=build /server /server
COPY --from=build /hub /hub

ENTRYPOINT ["/server"]
