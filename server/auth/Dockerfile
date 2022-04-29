
FROM golang:1.16-stretch as golang-builder

ARG dirname

COPY server/$dirname /workspace/server/app

COPY component /workspace/component
COPY proto /workspace/proto
COPY utils /workspace/utils

WORKDIR /workspace/server/app

RUN go env -w GO111MODULE=on
#RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service main.go

# 采用alpine作为部署镜像的基础环境
FROM alpine:3.15

# 使用国内镜像，加快打包速度
# RUN echo "http://mirrors.ustc.edu.cn/alpine/v3.15/main" > /etc/apk/repositories
# RUN echo "http://mirrors.ustc.edu.cn/alpine/v3.15/community" >> /etc/apk/repositories
RUN apk --no-cache add ca-certificates

ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /root/

COPY --from=golang-builder /workspace/server/app/server .

ENTRYPOINT ["./server"]