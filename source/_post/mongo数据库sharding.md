---
title: mongo数据库sharding学习记录
date: 2016-12-09 16:36:40
tags:
    - mongo
    - sharding
category:
    - 数据库学习记录
---
参考文章：http://www.runoob.com/mongodb/mongodb-sharding.html

windows系统。mongo3.2版本

建立目录s0,s1,log,config

步骤一：启动Shard Server
```
.\bin\mongod.exe --port 27100 --dbpath=.\mongos\s0 --logpath=.\mongos\log\s0.log --logappend --journal --storageEngine=mmapv1 --shardsvr

.\bin\mongod.exe --port 27101 --dbpath=.\mongos\s1 --logpath=.\mongos\log\s1.log --logappend --journal --storageEngine=mmapv1 --shardsvr
```

步骤二： 启动Config Server
```
.\bin\mongod.exe --port 27200 --dbpath=.\mongos\config --logpath=.\mongos\log\config.log --logappend --journal --storageEngine=mmapv1 --configsvr
```

步骤三： 启动Route Process
```
.\bin\mongos.exe --port 40000 --configdb localhost:27200  --logpath=.\mongos\log\router.log --chunkSize 1
```
注意:
1. 我这里不加--configsvr就出现Surprised to discover that localhost:27200 does not believe it is a config server错误。
2. chunkSize太大的话，会导致插入很多很多数据后才分片。一开始没注意，以为设置错误没有分片。。。chunkSize以M为单位。

步骤四： 配置Sharding

- 连接路由服务器
```
.\bin\mongo.exe admin --port 40000
```

- 设置分片服务器
```
db.runCommand({ addshard:"127.0.0.1:27100" })
db.runCommand({ addshard:"127.0.0.1:27101" })
db.runCommand({ addshard:"127.0.0.1:40000" })
或者
sh.addShard("127.0.0.1:27100")
```

- 设置要分片的数据库
```
db.runCommand({ enablesharding:"qdgame" })
或者：
sh.enableSharding("qdgame")
```

- 关闭balancing
```
sh.disableBalancing("qdgame.backpack")
```
- 设置分片collection的index
```
use qdgame
db.backpack.ensureIndex({usrid:1,id:1})
```

- 设置要分片的collection
```
db.runCommand({ shardcollection: "qdgame.backpack", key: { usrid:1, id:1}})
或者
sh.shardCollection("qdgame.backpack",{usrid:1, id:1})
```

- 开启balancing
```
sh.enableBalancing("qdgame.backpack")
```

注意：设置一定要在admin数据库下。
