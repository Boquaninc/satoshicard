# README
## Game rules
### General rules
- Cards count the same number of points as in baccarat: aces = 1, 2-10 = pip value, face cards = zero. As in baccarat, if a group of cards has a total point value greater than 9, then the tens digit is dropped and the point value of the hand is the terminal digit of the sum of the individual points.
- There are five cards. Three-card hand is zero points and the two-card hand has 1 to 9 points. The more points within this range, the higher the rank. For example, (4, 6, Q, 9, 9) is Bull 8
- Impossible to make zero points in the three-card hand. For example, (7, 8, A, 4, J) is No bull
- Three-card hand and two-card hand are both zero points, known as a "Niu Niu." which translates to "bull bull" in English. For example, (7,8,5,K,10) is Niu Niu
- In Niu Niu poker game, the one with a higher ranking wins the game. For example, if player A is Bull 9, player B is Bull 5, player C is Niu Niu, play D is No bull, then player D should pay player A twice the bets, player B one times the bets, player C three times the bets. Player C in this round has the highest ranking, so player C wins player A, player B and player D three times the bets. Player B should pay player A twice the bets, player C three times the bets, while wins player D one times the bets. Player A wins player B twice the bets, player D twice the bets, while pays player C three times the bets. (In this demo game, it only allows two players at a time)
### General hands
- No bull: for example, 10 3 2 Q 6 is No bull, because it is impossible to make zero points in the three-card hand
- Bull 1-9: for example, 10 3 7 Q 6 is Bull 6
- Niu Niu: for example, 10 3 7 Q K is Niu Niu
### Ranking of hands
- Spades > Hearts > Club >Diamond
- No bull, the side with the highest ranking card shall win. The ranks of the cards, from highest to lowest, is: K>Q>J>10>9>8>7>6>5>4>3>2>ACE.
- Bull 1-9, the more points within this range, the higher the rank. From highest to lowest, is 9-8-...A
- Niu Niu, in the event that the highest ranking card does not break a tie, then the highest ranking suit shall win. The order of suits, from highest to lowest, is Spades > Hearts > Club >Diamond.
### Scoring methods
- Five calf -------------------------------------------------- 7 times and the bets
- Bomb (Four cards of the same rank)----------------------- 6 times and the bets
- Gold Bull (All face cards)---------------------------------- 5 times and the bets
- Silver Bull (Four face cards)------------------------------- 4 times and the bets
- Niu Niu--------------------------------------------------- 3 times and the bets
- Bull 7, Bull 8, Bull 9---------------------------------------- 2 times and the bets
- No bull---------------------------------------------------- 1 times and the bets

Face cards are K, Q, J. The scoring method between the banker and the player is fixed. The amount of money won by the banker is directly proportional to the amount of money lost by the player. For example, if the player bets 1 BSV, and the banker gets Niu Niu and wins, then the player should pay 3 times the bets, which is 3 BSV, to the banker. So the amount of winning and losing chips between the two sides is fixed.

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


