---
title: mini gate framework
author: motto123
date: May 1th, 2022
---

[toc]

## 项目说明

#### 1. 介绍

- mini gate是一个轻量级的长连接服务框架,可用于im、游戏、物联网等需要使用长连接的场景,框架自身包含一套完整的路由体系和websock,tcp连接

- 框架中封装了完整的gate服,业务服务基类(BusinessSerer),redis,mysql,rabbitMQ,日志,panic捕捉逻辑,docker,等一系列工具
- 开发者只用关心业务服的实现,和gate服路由的注册,可以节省对公共逻辑开发的时间
- 如果想让业务费A,业务服B...,gate进行一些内部联动,可使用grpc
- example里有一个完整的demo实现login,register,chat功能
- 主要使用的技术grpc,protobuf,rabbitMQ,rides,mysql,websocket,tpc

#### 2. 项目文件目录说明

```shell
├── br.sh        	编译和启动server的脚本
├── build.sh     	编译server的脚本
├── clear.sh     	清理脚本
├── component    	公共组件
│   ├── amqp   	 	队列
│   ├── base     	server基础
│   ├── codec     数据包解码和编译	
│   ├── db      	数据和缓存
│   └── log     	log通用日志，可以打印到屏幕和文件中，日志文件按天切割
├── debug.sh     	dlv启动脚本
├── doc             文档                                                  
│   ├── doc.html                                                          
│   └── doc.md  	
├── docker_build.日志文件                                       
├── doc.sh       	文档html生成工具
├── kill.sh      	杀死server的脚本
├── log           
│   ├── gate    
│   └── main                                                              
├── proto           proto文件                                                     
│   ├── auth    	
│   ├── email                                                             
│   ├── gate 	
├── proto.sh     	自动编译proto文件的脚本
├── README.md     
├── resource     
│   ├── bin                                                               
│   └── doc_css                                                           
├── restart.sh   	启动server的脚本
├── server       	所有的server存放目录
│   ├── email
│   ├── gate
│   └── hub
├── start.sh		重启server的脚本
└── utils			工具
└── user_table.sql	mysql建库建表语句

```



#### 2. 项目架构

![](https://mottopicturecloud.oss-cn-chengdu.aliyuncs.com/typora/202204151651415.png)

#### 3. 报文协议

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
       - Kick = 0x05  客户端下线
   
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
   
           ```shell
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

#### 4. 服务器与客户端交互流程

#### 5. 使用说明

- protobuf注解

  ```protobuf
  // file: test.proto
  syntax = "proto3";
  
  package pb;
  option go_package = "/pb";
  
  message IP {
    // @gotags: valid:"ip"
    string Address = 1;
    
    // Or:
    
    // @gotags: bson:"name" from:"name"
    string Name = 2;
  
    // Or:
    string MAC = 3; // @gotags: validate:"omitempty"
  }
  ```

  

- error使用

  使用`err = errors.WithStack(err)`包裹普通的error,日志格式化输入时，使用`%+`。这样会打印出stack信息。

- 文档生成

  修改doc/doc.md,`.doc.sh`自动把doc.md转换成doc.html

- id使用雪花算法生成,分布式、唯一、有序的id，单机模式绝对有序，多服务相对有序。

#### 6. 代码详解
- 如果想看我的编码思路和更多的源码解释可以去csdn看
  https://blog.csdn.net/qq_27507377/article/details/124505124

#### 7. 后续开发计划
- [ ] 增加etcd服务发现
- [ ] 修复项目内的小bug
- [ ] 增加http连接

#### 8. 感谢

- 参考了大佬[nano](https://github.com/lonng/nano)的报文协议
- 学习了大佬[zinx](https://github.com/aceld/zinx),[cellnet](https://github.com/davyxu/cellnet)思路
