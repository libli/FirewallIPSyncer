# 腾讯云轻量服务器防火墙自动添加IP

## 介绍
如果你想使用腾讯云轻量服务器开放一些服务，仅自己能使用，可以使用本程序。

比如可以直接在公网上使用一些局域网服务，如samba、私人网盘、http proxy、家庭内网穿透等。避免开放到公网上有漏洞被利用。

## 原理

server端运行在腾讯云轻量服务器上，用来解决动态IP的问题。如你手机在外使用，可SSH连接一次服务器，即可将手机的IP添加到防火墙中。

client端运行在固定地方，如家里路由器、公司电脑等。client端可以设置crontab定时运行，将本机的公网IP添加到防火墙中。

client端公网IP获取方式：
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

在服务器上运行的时候，Tag可以设置为`#SSH`。

以下是服务端运行：

编写shell角本：
```bash
vi /usr/local/bin/run_ipsync.sh

```

添加执行权限：`chmod +x /usr/local/bin/run_ipsync.sh`

然后修改
```bash
vi ~/.bashrc
run_ipsync.sh
```

即每次登录服务器都会运行一次docker来自动更新防火墙。

## 家里软路由使用

```bash
docker run --name=ipsync -d --restart=unless-stopped \
  -e SecretID=AKIDxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  -e SecretKey=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  -e InstanceID=lhins-xxxxxxxx \
  -e Region=ap-guangzhou \
  -e TYPE=client \
  -e Tag='#home' \
  libli/ipsync:latest
```

配置crontab，每天凌晨7点运行一次（一般运营商每天凌晨5点重置公网IP）：
```bash
0 7 * * * docker restart ipsync
```