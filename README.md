# 服务注册发现
### 内容介绍
- 微服务为什么引入服务注册发现
- 服务注册中心设计原理
- Golang 代码实现服务注册中心

### 项目使用
- 服务构建 
```shell
cd cmd
go build -o discovery main.go
./discovery -c configs.yaml
```

- 服务注册
```shell
curl -XPOST http://127.0.0.1:6666/api/register -H 'Content-Type:application/json' -d'{"env":"dev", "appid":"testapp","hostname":"testhost1","addrs":["rpc:aaa","rpc:bbb"],"status":1,"replication":true}'
```

- 服务发现
```shell
curl -XPOST http://127.0.0.1:6666/api/fetch -H 'Content-Type:application/json' -d'{"env":"dev", "appid":"testapp","status":1}'

curl -XPOST http://127.0.0.1:6666/api/fetchall
```

- 服务续约
```shell
curl -XPOST http://127.0.0.1:8866/api/renew -H 'Content-Type:application/json' -d'{"env":"dev","appid":"testapp","hostname":"testhost","replication":true}'
```

- 服务取消
```shell
curl -XPOST http://127.0.0.1:8866/api/cancel -H 'Content-Type:application/json' -d'{"env":"dev","appid":"testapp","hostname":"testhost","replication":true}'
```

### 代码解读
[服务注册发现](https://mp.weixin.qq.com/s?__biz=MzIyMzMxNjYwNw==&mid=2247484142&idx=1&sn=0844fc63f9463b614afc23f450f266f2&chksm=e8215dfedf56d4e8c11c3e87c4a71de5fe65ad20fdd5a92c6ee58e8c3e0ba358e03f4e4e1f2e&token=2101083059&lang=zh_CN#rd)

扫码关注微信公众号 ***技术岁月*** 支持：

![技术岁月](https://i.loli.net/2021/01/21/orQm9BUkEqKAR6x.jpg)
