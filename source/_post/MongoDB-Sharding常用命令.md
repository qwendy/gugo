---
title: MongoDB-Sharding常用命令
date: 2016-12-22 20:08:07
tags:
    - mongo
    - sharding
category:
    - 数据库学习记录
---
资料来源： https://docs.mongodb.com/v3.2/tutorial/administer-shard-tags/
详细命令参考： https://docs.mongodb.com/v3.2/reference/sharding/
# 分片常用
##  配置Sharding

- 连接路由服务器
```
.\bin\mongo.exe admin --port 40000
```

- 设置分片服务器
```
db.runCommand({ addshard:"127.0.0.1:27100" })
db.runCommand({ addshard:"127.0.0.1:27101" })
db.runCommand({ addshard:"127.0.0.1:27017" })
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
<!-- more -->
- 设置要分片的collection
```
db.runCommand({ shardcollection: "qdgame.backpack", key: { _id:1}})
或者
sh.shardCollection("qdgame.backpack",{usrid:1, id:1})
```

- 开启balancing
```
sh.enableBalancing("qdgame.backpack")
```

注意：设置一定要在admin数据库下。
## 使用hash分片一个collection
```
sh.shardCollection( "database.collection", { <field> : "hashed" } )
```

## 改变chunk大小
```
db.settings.save( { _id:"chunksize", value: <size> } )
```
## 添加index索引
```
db.collection.ensureIndex({a:1,b:-1})
```
# 管理分片标签
## 标记一个分片
当连接到mongos实例时，使用sh.addShardTag（）方法将标记与特定分片关联。单个分片可以具有多个标签，并且多个分片也可以具有相同的标签。
```
sh.addShardTag("shard0000", "NYC")
sh.addShardTag("shard0001", "NYC")
sh.addShardTag("shard0002", "SFO")
sh.addShardTag("shard0002", "NRT")
```
在连接到mongos实例时，您可以使用sh.removeShardTag（）方法从特定分片中删除标签，如以下示例所示，该示例从分片中删除NRT标签：
```
sh.removeShardTag("shard0002", "NRT")
```

## 标记shard key范围
在连接到mongos实例时使用sh.addTagRange（）方法来给某个范围的shard key添加标签。任何给定的分片键范围都只具有一个分配的标签。您不能重叠定义的范围，也不能多次标记同一范围。
```
sh.addTagRange("records.users", { zipcode: "10001" }, { zipcode: "10281" }, "NYC")
sh.addTagRange("records.users", { zipcode: "11201" }, { zipcode: "11240" }, "NYC")
sh.addTagRange("records.users", { zipcode: "94102" }, { zipcode: "94135" }, "SFO")
```
## 从shard key范围中删除标记
```
use config
db.tags.remove({ _id: { ns: "records.users", min: { zipcode: "10001" }}, tag: "NYC" })
```

## 访问标签
sh.status（）列出与每个分片相关联的分片（如果有的话）的标签。shard的标记存在于config数据库的shards
集合中的shard文档中。
```
use config
db.shards.find({ tags: "NYC" })
```
您可以在config数据库的tags集合中找到所有命名空间的标记范围。 sh.status（显示所有标签范围。要返回用NYC标记的所有shard key范围，请使用以下操作：
```
use config
db.shards.find({ tags: "NYC" })
```


# 例子：根据地区来划分数据
## 场景
在messages集合中创建测试数据如下：
```
{
  "_id" : ObjectId("56f08c447fe58b2e96f595fa"),
  "country" : "US",
  "userid" : 123,
  "message" : "Hello there",
  ...,
}
{
  "_id" : ObjectId("56f08c447fe58b2e96f595fb"),
  "country" : "UK",
  "userid" : 456,
  "message" : "Good Morning"
  ...,
}
{
  "_id" : ObjectId("56f08c447fe58b2e96f595fc"),
  "country" : "DE",
  "userid" : 789,
  "message" : "Guten Tag"
  ...,
}
```
### shard key
messages集合使用{ country : 1, userid : 1 }组合索引作为shard key。
每个文档中的country 字段允许在每个不同的country值上创建标签范围。
相对于country，userid字段为shard key提供一个高基数低频率的元素
### 写操作
在Tag相关的分片中，mongos将到来的文档和配置的tag范围比较。如果文档符合配置标签的范围，mongos将此文档路由到此标签的分片中
这插入环节，MongoDB可以将不合格任何tag范围的文档路由到随机的分片中。

### 读操作
如果这个请求至少包含country字段，MongoDB可以路由请求到特定的分片中。
例如MongoDB可以尝试对以下查询执行有针对性的读取操作：
```
chatDB = db.getSiblingDB("chat")
chatDB.messages.find( { "country" : "UK" , "userid" : "123" } )
```

### 均衡器
均衡器迁移带有tag的块（Chunk）到适合的分片（shard）。在迁移之前，分片可能包含违反配置的tag和tag范围的块。一旦平衡完成，碎片应只包含不违反其分配的标记和标记范围的块。
添加个删除标签可以块的迁移。这取决于数据集的大小和标记范围影响的块数量，这些迁移可能会影响集群的性能。

## 操作
### 取消均衡器
为了降低性能影响，可以在集合上禁用平衡器，以确保在配置新标记时不会发生迁移。考虑在特定的计划窗口中运行均衡器。
```
sh.disableBalancing("chat.message")
```
使用sh.isBalancerRunning()来检查平衡器进程当前是否正在运行。等待任何当前平衡回合完成后再继续。

### 为每一个分片添加标签
````
sh.addShardTag(<shard name>, "NA")
```
您可以通过运行sh.status（）来查看分配给给定分片的标签。

### 定义标签范围
使用sh.addTagRange()方法。这个方法需要：
- 完成的目标collection的名字
- 范围的下限
- 范围的上限
- 标签的名字

```
sh.addTagRange(
  "chat.messages",
  { "country" : "US", "userid" : MinKey },
  { "country" : "US", "userid" : MaxKey },
  "NA"
)
```
### 开启均衡器
```
sh.enableBalancing("chat.message")
```
使用sh.isBalancerRunning()来检查平衡器进程当前是否正在运行

### 查看更改
您可以使用sh.status()
