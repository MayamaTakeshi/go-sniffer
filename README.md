
# go-sniffer

> Capture mysql, mssql, redis, http, mongodb etc protocol...
> 抓包截取项目中的数据库请求并解析成相应的语句，如 mysql 协议会解析为 sql 语句，便于调试。
> 不需要修改代码，直接嗅探项目中的数据请求。

[![GitHub license](https://img.shields.io/github/license/40t/go-sniffer.svg?style=popout-square)](https://github.com/40t/go-sniffer/blob/master/LICENSE)

[中文使用说明](#中文使用说明)

## Support List

- [mysql](#mysql)
- [redis](#redis)
- [http](#http)
- [mongodb](#mongodb)
- [mssql](#mssql)
- [kafka](#kafka)

## Demo

``` bash
go-sniffer en0 mysql
```

![image](https://github.com/40t/go-sniffer/raw/master/images/demo.gif)

## Setup

- support : `MacOS`, `Linux`
- not support : `Windows`
- If you encounter problems in the `go get` process, try upgrading the go version.

### Centos

``` bash
yum -y install libpcap-devel
```

### Ubuntu

``` bash
apt-get install libpcap-dev
```

### Mac OS

``` bash
brew install libpcap
```

### RUN

``` bash
go get -v -u github.com/40t/go-sniffer
cp -rf $(go env GOPATH)/bin/go-sniffer /usr/local/bin
go-sniffer
```

## Usage

``` bash
==================================================================================
[Usage]

    go-sniffer [device] [plug] [plug's params(optional)]

    [Example]
          go-sniffer lo redis          Capture redis packet
          go-sniffer lo mysql -p 3306  Capture mysql packet

    go-sniffer [command]
               --help "this page"
               --env  "environment variable"
               --list "plugin list"
               --ver  "version"
               --dev  "device"
    [Example]
          go-sniffer --list "show all plugin"

==================================================================================
[device] : lo0 :   127.0.0.1
[device] : en0 : xx:xx:xx:xx:xx:xx  192.168.199.221
==================================================================================
```

## 中文使用说明

``` bash
=======================================================================
[使用说明]

    go-sniffer [设备名] [插件名] [插件参数(可选)]

    [例子]
          go-sniffer en0 redis          抓取redis数据包
          go-sniffer en0 mysql -p 3306  抓取mysql数据包,端口3306

    go-sniffer --[命令]
               --help 帮助信息
               --env  环境变量
               --list 插件列表
               --ver  版本信息
               --dev  设备列表
    [例子]
          go-sniffer --list 查看可抓取的协议

=======================================================================
[设备名] : lo0 :   127.0.0.1
[设备名] : en0 : x:x:x:x:x5:x  192.168.1.3
[设备名] : utun2 :   1.1.11.1
=======================================================================
```

### Example

``` bash
go-sniffer lo0 mysql
go-sniffer en0 redis
go-sniffer eth0 http -p 8080
go-sniffer eth1 mongodb
go-sniffer eth0 mssql
```

## License

[MIT](http://opensource.org/licenses/MIT)

## FAQ

### 找不到 libpcap 动态链接库

报错：

```text
error while loading shared libraries: libpcap.so.0.8: cannot open shared object file: No such file or directory
```

解决办法：

首先参考 [#Setup] 安装 libpcap 依赖，安装完成后如发现版本不一致仍找不到对应的 so 文件，需要手工添加软链。

```bash
# 软链的 so 文件版本视不同系统而定，请根据需要修改
cd /lib64/ && ln -s libpcap.so.1.5.3 libpcap.so.0.8
```
