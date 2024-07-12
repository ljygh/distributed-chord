# Chord
Chord algorithm was implemented in this project according to this paper:  
I. Stoica et al., "Chord: a scalable peer-to-peer lookup protocol for Internet applications," in IEEE/ACM Transactions on Networking, vol. 11, no. 1, pp. 17-32, Feb. 2003, doi: 10.1109/TNET.2002.808407.
keywords: {Peer to peer computing;Protocols;Internet;Routing;Application software;Computer science;Analytical models;Costs;Centralized control;Network servers},

## Environment
Intall lastest version of golang: https://go.dev/doc/install

## Compile
```
go mod init chord
go build
```

## Run
Below are some commands used to run functionality on the chord program  

go build  
./chord 0 8000  
join: ./chord 10 8004 127.0.0.1 8000 false &  
./create.sh  
./join_multiple.sh runs all the commands concurrently  
print: prints information of node  
store 1 (store a file called 1)  
files are stored in the succesor  

The arguments take in an id, port and an IP-adress

## Algorithm
### 0. Hash
node和key（file）都通过sha-1 hash得到相应的id。

### 1. Data structure of a node
1. finger[k].start: (n + 2^(k-1)) mod 2^m
2. finger[k].node: first node >= finger[k].start
3. successor: finger[1].start
4. predecessor

### 2. Find predecessor and successor
See page 5 in that page.
1. n.find_successor(id): n.find_predecessor(id).sucessor
2. n.find_predecessor(id)： 如果这个node就是predecessor，那么就找到了。如果不是，那就看看最靠近这个id的node是不是。
3. n.closest_preceding_finger(id): 找到finger table 中最靠近这个id的node。

### 3. Create
Page 6 in that paper. All variables in the node are itself.

### 4. Join
Page 6 in that paper.
1. n.join(n'): Init finger table, Update others
2. n.init_finger_table(n'): 更新successor和predecessor，把predecessor的successor改了(`Keys need to be transferred here`)。校对每一个finger，通过find_sucessor。
3. n.update_others: 把每一个可能修改的finger node的node都update 指定的finger。
4. n.update_finger_table: 检查指定的finger，如果需要更新则更新。

### 5. Lookup or upload
通过find_sucessor找到指定的node进行查询和上传。

### 6. Stabilization
This is used to fix mistakes caused by concurrent joining and node failures. See page 7 in that paper.
1. n.join(n')：只指定successor，在这个项目里，还是用的上面版本的join。
2. n.stabilize: 检查successor的predecessor，如果有误，更换successor并notify新的successor。
3. n.notify(n'): n'检查predecessor并更换(`Keys need to be transferred here`)。
4. n.fix_fingers: 间歇性的检查fingers，一次随机检查一个。

### 7. File backup and check failures
1. Upload的时候同时将文件上传到node的predecessor。
2. Stabilize的时候check successor是否alive，如果非alive，将这个node中的backup files recover到正确的nodes。





