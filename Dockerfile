FROM golang:1.14 as builder
WORKDIR /usr/local/src/
COPY . ./

RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' -a \
    -v -o ./scorer ./server

FROM gcr.io/distroless/static
COPY --from=builder /usr/local/src/scorer /scorer
CMD ["/scorer"]

EXPOSE 5000
EXPOSE 5001

# docker build -t yourworkspaceacr.azurecr.io/joongrpc:1 -f ./Dockerfile .
# docker push yourworkspaceacr.azurecr.io/joongrpc:1
