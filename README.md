# majsoul

## [majsoul](https://game.maj-soul.com/1) 的客户端通信协议Go实现

使用grpc生成了向majsoul服务器请求的通信协议，但是对于majsoul服务器的消息下发使用了更加原始的处理方式。

> current liqi.proto version v0.10.194.w

### 安装

```
go get -u github.com/constellation39/majsoul
```

示例文件在 [example](https://github.com/constellation39/majsoul/tree/master/example) 文件中