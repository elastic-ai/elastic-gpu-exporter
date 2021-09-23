FROM golang:1.16


ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /go/src/
COPY go.mod .
COPY go.sum .
RUN GO111MODULE=on go mod download
COPY . .

CMD ["/bin/bash", "/go/src/script/build.sh"]






