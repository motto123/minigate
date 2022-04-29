---
    title: minigame document
    date: Fri 08 Apr 2022
---

[toc]

## 项目说明

#### 1. 介绍

​ 多服务版的游戏框架，主要有gate服务，mq，多个业务服务，缓存，数据库，etcd

#### 2. 项目文件目录说明

```go
├── **br.sh**        编译和启动server的脚本
├── **build.sh**    编译server的脚本
├── **clear.sh**    清理脚本
├── **debug.sh**    dlv启动脚本
├── **kill.sh**    杀死server的脚本
├── **component**    公共组件
│  ├── **base**    server基础
│  ├── **conf**    conf配置
│  ├── **db**        数据和缓存
│  ├── go.mod
│  ├── go.sum
│  └── **log**        log通用日志，可以打印到屏幕和文件中，日志文件按天切割
├── **log**        日志文件
├── **proto**        proto文件
│  ├── **auth**
│  ├── **email**
│  ├── **gate**
│  ├── go.mod
│  ├── go.sum
├── **proto.sh**    自动编译proto文件的脚本
├── Readme.md
├── **restart.sh**    重启server的脚本
├── **server**        所有的server存放目录
│  ├── **email**
│  ├── **gate**
│  ├── **hub**
│  └── **mq**
├── **start.sh**    启动server的脚本
└── **utils**        工具
├── go.mod
└── go.sum
```

#### 2. 项目架构

#### 4. 报文协议

![报文协议](https://mottopicturecloud.oss-cn-chengdu.aliyuncs.com/typora/202204071724745.png)

- 报文分为2部分,packet和message

- packet

    - head

      5bytes

        - type

          1byte

            - Handshake = 0x01 客户端握手请求
            - HandshakeAck = 0x02 服务器握手ack
            - Heartbeat = 0x03 心跳
            - Data = 0x04 数据包
            - Kick = 0x05 客户端下线

        - len

          3bytes,body的长度

        - subpackage

          1byte，分包，但是没实现

    - body

      range 0-64kb

      二进制数据

    - message

      就是body

        - head

          range 1- xx bytes

            - flag

                - empty

                  2bit

                - message type

                  3bits

                  ```go
                  // Message types
                  // s: server
                  // c: client
                  // who can send message type
                  Request  Type = 0x00 // c
                  Notify        = 0x01 // c
                  Response      = 0x02 // s
                  Push          = 0x03 // s
                  Ack           = 0x04 // s,c
                  ```

                - route

                  1bit,是否压缩路由

                - date type

                  2bits

        - len

          1byte,message id的长度

        - message id

          range 0-8bytes,uint64类型

        - len

          1byte,route的长度

        - route

          路由的内容，如果是字符串，它就是没有压缩的路由，如果是数字它就是压缩路由

        - len

          1byte,obj name的长度

        - obj name

          range 0-256bytes

          后面data里存的对象的名字

    - data

      range 0-约60kb

​

#### 5. 服务器与客户端交互流程

#### 6. 注意事项

