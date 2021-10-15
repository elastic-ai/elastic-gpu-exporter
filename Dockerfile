FROM golang:1.16-stretch as build


ENV GO111MODULE=on
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
# config
WORKDIR /go/src/nano-gpu-exporter
COPY . .

RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN GO111MODULE=on go mod download
RUN go get github.com/prometheus/client_golang/prometheus@v1.0.0
#RUN go mod download github.com/alex337/go-nvml
RUN go build -o /go/bin/nano-gpu-exporter cmd/main.go

# runtime image
FROM nvidia/cuda:10.0-base
COPY --from=build /go/bin/nano-gpu-exporter /usr/bin/nano-gpu-exporter
RUN ln -s /usr/lib/x86_64-linux-gnu/libnvidia-ml.so.1 /usr/lib/x86_64-linux-gnu/libnvidia-ml.so

CMD ["nano-gpu-exporter"]