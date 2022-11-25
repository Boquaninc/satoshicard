# README
## 演示教程
### docker

首先拉取镜像
```
docker pull xxxxxx
```
启动容器
```
docker run -d -it xxxxx /bin/bash
cd ~/root/satoshicard
sh start_bsv.sh
sh 1.sh
```
另外启动一个容器

```
docker exec xxxxx /bin/bash
cd ~/root/satoshicard
sh 2.sh
```
双方按照游戏进程玩游戏，注意到，若1号玩家开房间，则`1.sh`写定了其监听127.0.0.1:10001，若2号玩家开房间，则`2.sh`中写定了其监听127.0.0.1:10002


### shell

此演示demo依赖
- regtest bsv全节点
- zokrate
- golang
请自行安装

在克隆该工程之后，如果是linux
```
cd satoshicard
sh setup_linux.sh
```
如果是mac
```
cd satoshicard
sh setup_mac.sh
```

程序编译完成之后，分别打开两个shell
```
./satoshicard \
-help=false \
-listen=0.0.0.0:10001 \
-rpchost=127.0.0.1:19002 \
-rpcusername=regtest \
-rpcpassword=123 \
-gamecontractpath=./desc/satoshicard_release_desc.json \
-lockcontractpath=./desc/satoshicard_timelock_release_desc.json \
-key=ed909bc8d0b35d622a4c3b0c700fce4f1472c533289d5127a782c09c669fb1d7 \
-mode=0
```

```
./satoshicard \
-help=false \
-listen=0.0.0.0:10001 \
-rpchost=127.0.0.1:19002 \
-rpcusername=regtest \
-rpcpassword=123 \
-gamecontractpath=./desc/satoshicard_release_desc.json \
-lockcontractpath=./desc/satoshicard_timelock_release_desc.json \
-key=ed909bc8d0b35d622a4c3b0c700fce4f1472c533289d5127a782c09c669fb1d7 \
-mode=0
```

接着就可以按照游戏进程描述的进行游戏

## 游戏进程
```
1.alice创建房间，进入步骤2
2.bob加入房间，进入步骤3
3.双方输入原像，进入步骤4
4.双方签名，进入步骤5
5.由任意一方部署合约，进入步骤6
6.双方开牌，如双方均完成此操作，进入步骤7，如对手方在规定时间内为完成操作，进入步骤9
7.双方查验手牌，进入步骤8
8.认负认赢，进入步骤10
9.取走对手方押金，进入步骤10
10.游戏结束
```

## 命令格式
```
所有的命令均以“命令:参数”的格式从标准输入输入，如果改命令没有参数，则参数为空，即“命令:”的格式
```
## 命令列表

|  命令   | 功能  | 参数|例子|
|  ----  | ----  |---|---|
| host  | 创建房间 |无|host:|
| join  | 加入房间 |创建者监听的ip以及端口|join:127.0.0.1:10001|
| preimage  | 选择原像 |一个很大很大的十进制数|preimage:846585428578|
| sign  | 对部署合约交易进行签名 |无|sign:|
| publish  | 部署合约 |无|publish:|
| open  | 公布原像，赎回押金 |无|open:|
| takedeposit  | 超时时候取走对手方押金 |无|takedeposit:|
| check  | 查验双方手牌 |无|check：|
| win  | 胜利领走奖励 |倍数|win:5|
| lose  | 认负结束游戏 |无|lose:|


