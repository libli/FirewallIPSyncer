# 腾讯云轻量服务器防火墙自动添加IP

将运行本程序的 IP 自动添加到腾讯云轻量服务器防火墙中，即对本机放开所有服务。

本机公网IP获取方式：
http://myexternalip.com/raw
请确保该网站没有被分流到代理访问。

## 腾讯云配置
需要提前准备好腾讯云API的SecretId和SecretKey，以及轻量服务器的实例ID。

腾讯云上创建一个子用户，只给编程访问（不允许登录控制台）。

链接：https://console.cloud.tencent.com/cam/user/create?systemType=SubAccount

该用户只关联一个自定义策略，策略语法如下：
```
{
    "statement": [
        {
            "action": [
                "lighthouse:CreateFirewallRules",
                "lighthouse:DeleteFirewallRules",
                "lighthouse:ModifyFirewallRules",
                "lighthouse:DescribeFirewallRules"
            ],
            "effect": "allow",
            "resource": [
                "*"
            ]
        }
    ],
    "version": "2.0"
}
```

这样即使该secretKey泄露，也只有修改轻量服务器防火墙的权限。

## Docker 使用

给当前运行的机器一个标签，如在家使用就标记为`#home`，在公司使用就标记为`#company`。

```bash
docker run --name=ipsync -d --restart=unless-stopped \
  -e SecretID=AKIDxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  -e SecretKey=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  -e InstanceID=lhins-xxxxxxxx \
  -e Region=ap-guangzhou \
  -e Tag='#home' \
  libli/ipsync:latest
```

## ARM 软路由使用
编译arm版本：
`GOOS=linux GOARCH=arm64 go build`

复制到软路由上`/usr/local/bin`

编写自启动角本：
vi /etc/init.d/FirewallIPSyncer

```bash
#!/bin/sh /etc/rc.common

START=99
STOP=15

BIN=/usr/local/bin/FirewallIPSyncer

start() {
    if [ -x $BIN ]; then
        echo "Sleeping 300 seconds before starting FirewallIPSyncer..."
        sleep 300
        echo "Starting FirewallIPSyncer..."
        export SecretID=AKIDxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
        export SecretKey=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
        export Region=ap-guangzhou
        export InstanceID=lhins-xxxxxxxx
        export Tag='#Home'
        $BIN >> /var/log/FirewallIPSyncer.log 2>&1 &
    else
        echo "FirewallIPSyncer binary not found..."
    fi
}

stop() {
    echo "Stopping FirewallIPSyncer..."
    killall $(basename $BIN)
}
```

添加权限：
`chmod +x /etc/init.d/FirewallIPSyncer`

启动：
`/etc/init.d/FirewallIPSyncer start`