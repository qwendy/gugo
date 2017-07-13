---
title: 我的Redis初次实战
date: 2016-09-03 11:04:07
tags:
    - redis
    - 实践
category:
    - 数据库学习记录
---
因为各种原因，我必须在windows上进行服务器开发，所以我需要在windows上运行redis，即使《redis in action》上明确写着不推荐在windows上使用redis。
我在windows上使用的redis是从 https://github.com/dmajkic/redis/downloads 获取的。
我使用redis的桌面客户端为Redis Desktop Manager。使用的原因仅仅是第一个搜到的客户端就是他且没有让我讨厌的地方，就继续用了。软件地址https://redisdesktop.com/download。
极力推荐《redis实战》。从第一章开始，在介绍什么是redis，有什么用，与其他数据库有什么区别之后，就开始从一个项目开始，演示如何使用redis。讲的第一个项目是对文章进行投票，这让我的回忆起我的大学的第一个项目--一个投票系统的实现。
<!-- more -->
在看完这一章之后，我对redis有了个初步了解，并将在这一章学到的知识，运用到我的实际工作的一个小项目中：解析短信回执的xml，统计回执中的信息，统计用户的某个活动发送的短信各种状态的数目。解析后的xml数据大致如下
``` go
type Message struct {
    PhoneNum   string
    State      string
    UserID     int
    ActivityID int
}
```
本项目是解析一个xml文件，得到其中的信息，进行统计，返回统计的结果。本来可以将数据使用Map存储在内存中，等所有的信息统计后，将Map中的信息输出就可以了，但是这个项目以后要扩展成，需要将内存的数据进行持久化处理，定时过期数据和多服务器主从同步数据，所以使用redis无疑是最好的选择。
我仿造《redis实战》第一章的编程方法，对程序进行了改造。设定了：
- user集合： 存储接收到的xml中用户id的集合。集合内的元素类似{1,2,3}
- user_x_acitvity: 存储用户ID为x的活动id的集合，user_1_acitvity集合内的元素类似{1,2,3}
- user_x_acitvity_y: 存储用户ID为x，活动ID为y的手机号码信息。user_1_acitvity_1集合内的元素类似{"15952010343，DELIVRD"}，元素的值为"phoneNum,status"的组合
通过user集合找到有哪些用户收到了回执，通过user_x_acitvity记录用户的哪些活动在回执的记录中。并通过user_x_acitvity_y找到属于此活动的手机号码的状态。看完《redis实战》第一章最大的收获就是设定set的名字。可以通过设定set以特定的名字，来索引到特定的信息。
具体的程序代码在：https://github.com/qwendy/handleFeedback
``` go
func (rc *redisContainer) pushYiMeiXMLData(message YiMeiMessage) {
    activityID := message.Seqid >> 16
    userID := message.Seqid - (activityID << 16)
    // 用户的集合。设置接收到的用户的集合
    rc.client.SAdd("user", userID)
    // 设置某个用户的活动集合
    userActivitySet := fmt.Sprintf("user:%d", userID)
    rc.client.SAdd(userActivitySet, activityID)
    // 活动的手机号码状态集合
    activityPhoneSet := fmt.Sprintf("user_%d_activity_%d", userID, activityID)
    // 手机号码和状态
    phoneStatus := fmt.Sprintf("%s,%s", message.PhoneNum, message.State)
    rc.client.SAdd(activityPhoneSet, phoneStatus)
}

func (rc *redisContainer) Print() {
    for _, userID := range rc.client.SMembers("user").Val() {
        fmt.Printf("user:%s \n", userID)
        userActivitySet := fmt.Sprintf("user:%s", userID)
        for _, activityID := range rc.client.SMembers(userActivitySet).Val() {
            fmt.Printf("activityID:%s  \n", activityID)
            activityPhoneSet := fmt.Sprintf("user_%s_activity_%s", userID, activityID)
            for _, phoneStatus := range rc.client.SMembers(activityPhoneSet).Val() {
                fmt.Printf("phoneStatus:%s  ", phoneStatus)
            }
            fmt.Printf("\n")
        }

    }
}
```
