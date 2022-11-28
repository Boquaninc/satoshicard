# README
## Demo Tutorials
### docker

pull image first
```
docker pull xxxxxx
```
start container
```
docker run -d -it xxxxx /bin/bash
cd ~/root/satoshicard
sh start_bsv.sh
sh 1.sh
```
start another container

```
docker exec xxxxx /bin/bash
cd ~/root/satoshicard
sh 2.sh
```
Both sides play the game as it progresses. Please note that `1.sh` is written to monitor 127.0.0.1:10001 if player #1 opens the room, and `2.sh` is written to monitor 127.0.0.1:10002 if player #2 opens the room


### shell

This demo depends on
- regtest bsv full node
- zokrate
- golang
Please install it yourself

After cloning the project, if it is linux
```
cd satoshicard
sh setup_linux.sh
```
if it is mac
```
cd satoshicard
sh setup_mac.sh
```

Once the program has been compiled, open two separate shells
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

Then you can play the game according to the game process description

## game process
```
1.
Alice enters:
    `host:`
Create a room, go to step 2

2.
Bob enters:
    `join:127.0.0.1:10001`
Join the room, go to step 3

3.
Alice and Bob both enter:
    `preimage:123456`
As the preimage for dealing, go to step 4

4.
Alice and Bob both sign:
    `sign:`
go to step 5

5.
Contract deployed by either party:
    `publish:`
go to step 6
6.
Alice and Bob both show their cards:
    `open:`
If both parties complete this operation, go to step 7, if the counterparty does not complete the operation within the specified time, go to step 9

7.
Both players check the other player's hand: 
    `check:`
go to step 8
8.
Time to receive the result, winner enters：
    `win:3`
Loser enters：
    `lose:`
Go to step 10

9.
Take the counterparty's deposit, enter:
    `takedeposit:`
Go to step 10
10.
Game over
```
Among them, the parameter of step 2 is Alice's ip port, the parameter of step 3 is a random large number, and the parameter of step 8 is determined by the actual situation of winning or losing

## Command Format
```
All commands are entered in the "command:parameter" format, or null if the command has no parameters, i.e. the "command:" format
```
## Command List

|  Command   | Function  | Parameter|Example|
|  ----  | ----  |---|---|
| host  | create a room |none|host:|
| join  | join the room |the ip and port the room creator monitors|join:127.0.0.1:10001|
| preimage  | select preimage |a very large decimal number|preimage:846585428578|
| sign  | sign the deployment contract transaction |none|sign:|
| publish  | deploy contract |none|publish:|
| open  | publish preimage, get deposit back |none|open:|
| takedeposit  | take the deposit of the counterparty when the timeout expires |none|takedeposit:|
| check  | both players check the other player's hand |none|check：|
| win  | claim the reward for victory |mutiple|win:5|
| lose  | end the game by conceding defeat |none|lose:|


