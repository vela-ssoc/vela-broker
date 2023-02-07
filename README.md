# 安全平台代理节点

## 配置文件示例

```json5
{
  // broker 节点的 ID
  "id": 123,
  // broker 节点的密钥
  "secret": "50e7e0f37c6147f9868457a7b97d077179a3f411309e4175a1634243957d0f2d",
  // broker 节点的版本
  "semver": "0.0.1-dev",
  // 要连接的中心端地址
  "servers": [
    {
      // 中心端是否开启了 TLS
      "tls": true,
      // 中心端地址，如果没有注明端口号，则开启 TLS 默认为 443，不开启默认为 80
      "addr": "127.0.0.1",
      // 开始 TLS 的 servername，没有配置则默认 addr 去除端口号
      "name": "local.eastmoney.com"
    },
    {
      "tls": true,
      "addr": "local.eastmoney.com"
    },
    {
      "addr": "127.0.0.1:8899"
    },
    {
      "addr": "guba.eastmoney.com"
    }
  ]
}

```
