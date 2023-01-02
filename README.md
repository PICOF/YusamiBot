# YusamiBot
**一个实验性的项目，做出来单纯因为好玩，实现了一些有意思的功能**

*注：该项目是基于 [go-cqhttp](https://github.com/Mrs4s/go-cqhttp) 消息框架，使用 golang 开发的 qq 机器人*
## 实现功能：
- 学习，包含精确学习与模糊学习（所以就没做 help 指令，你看毕竟可以学更有成就感……）
- 加精与查看，直接存入 MongoDB 中
- 看看新闻（获取 60 秒读懂世界内容）
- 新番查询（与 Bangumi 同步）
- 番剧查找（整合同步多个动漫网站数据，包含动漫花园数据）
- pixiv 美图，特别感谢 [LoliconApi](https://api.lolicon.app/#/) *现在由于鹅的大力封锁，全部转成了扫码观看*
- 原神抽卡（？别看我，群友让我加的，原神什么的真不熟）
- jm 本子查询与观看（有傻冒放着人家图床不用给自己服务器流量冲业绩，**后面真会改的**）
- 日记功能，每天写日记多是一件美式，实时更新，有专门的页面可供查看
- 课表查询，目前属于 UESTC 限定功能，没什么验证（飞书：锅来），绑对了你连住你隔壁的男同今天上什么课都一清二楚
- ai 绘图，曾经可以使用，现在穷了，给点（
- ai 聊天，使用的是 CharacterAi，openAi 以后再说？
- openAi 最近比较火，加了一个 openAi 的 chat 功能，但是要求你自己有一个 openAi 的账号（且账号上有额度），因为是为了玩玩匆忙写出来的，很多功能不完善。
- b站订阅提醒，包含直播和动态提醒

## 安装：
首先下载最新的 release 文件，将其解压到文件夹中

本项目使用 websocket 连接至 go-cqhttp，因此将 go-cqhttp 的 config 文件中的相关配置项改成如下样式:

```yaml
  # 反向WS设置
  - ws-reverse:
      # 反向WS Universal 地址
      # 注意 设置了此项地址后下面两项将会被忽略
      universal: ws://127.0.0.1:8081/msg
```

项目由于其他功能会使用部分路径，因此反向 ws 接口为 `/msg`，其余配置如账号等并没有特别特殊的地方

在启动 go-cqhttp 之后，运行项目文件夹中的 Lealra 文件，bot 启动

## 相关功能：
### 学习

| 指令                            | 效果                     |
|-------------------------------|------------------------|
| [回复] 学习 XXX                   | 将回复内容作为 XXX 的精确匹配响应    |
| [回复] 模糊学习 XXX                 | 将回复内容作为 XXX 的模糊匹配响应    |
| [回复] 忘记 XXX                   | 若回复内容是 XXX 的响应，则删除对应记录 |
| 清理僵尸回复<br/>*仅限私聊且需要 admin 账号* | 清理所有僵尸回复，并返回执行结果       |

该功能模块主要有两种不同的模式，一种是精确学习，另一种是模糊学习：

- 精确学习会精确匹配全字段，模糊匹配则检测是否包含
- **优先精确匹配，其次再模糊匹配**

![](readmePic/learn.jpg)

让 bot 遗忘只用回复需要遗忘的消息即可：

![](readmePic/forget.jpg)

**遗忘也会优先遗忘精确学习的回复，其次才是模糊学习**

#### 清理僵尸回复
由于本项目使用的是 qq 图床（也就是最方便 CQ 码操作的格式），很多情况下容易出现图片过期的情况。如果回复是纯图片，就会造成无回复的现象。

针对这个问题项目有三点应对策略：

1. **清除僵尸回复**

   所谓僵尸回复就是占用数据库存储空间但实际上已经过期且无法使用的回复，只要回复中图片存在过期那么都可以将其视为失去了原本的表述意义，因此需要将其全部清理：

   ![](readmePic/zombieResp.jpg)

   **切记该功能仅能在私聊中使用！且要求使用者是预先设置好的 admin 用户！**
2. **通过发送图片保存图床稳定性**

   不推荐使用，在配置文件中有详细说明，频繁发送消息还可能让你更容易被鹅盯上最后 bot 死翘翘就不好玩了

   配置中设置后自动执行，周期为 48 小时
3. **base64备份（未完成）**

   实际上发出图片后被 go-cqhttp 解析为 CQ code 会有两个属性：一个是 file 为本地缓存文件名，另外一个 url 是 qq 图床地址，只有两个都失效后该图片才会发不出去，本地缓存文件名为图像的 md5，qq 图床也使用 md5 进行标识，我们可以创建一个 md5->base64 的映射表，缺什么补什么，并且一旦使用 base64 重新发送一次，前面两个属性就又会重新生效（md5 不变）

### 加精与查看

| 指令                                 | 效果                |
|------------------------------------|-------------------|
| [回复某消息] 设精<br/>*只要最后两个字是设精都会触发*    | 将回复的消息加入精华        |
| 精华 [date=x][uid=x][name=x][word=x] | 查询精华消息，四个参数可以随意使用 |

实际上该功能在使用过程中还是发现有诸多问题，因此后续会加以更正

目前支持聊天记录加精（包括单层转发），根据内容（word）、昵称（name）、uid（uid）、时间（date）进行匹配，四项参数可以任意搭配

只要回复消息尾部包含“设精”，则会自动设精，有相应权限还可以直接加入群聊精华：

![](readmePic/selectedSearch.jpg)
![](readmePic/setSelectedMsg.png)

### 看看新闻
输入看看新闻即可查询当天新闻，有时可能未更新，则显示前一天新闻

| 指令   | 效果                 |
|------|--------------------|
| 看看新闻 | 返回当天的 60 秒读懂世界新闻图片 |

如图：

![](readmePic/news.jpg)

### 番剧助手
相关功能分为两部分：
#### 新番查询
该功能是基于在 [Bangumi](https://bangumi.tv/calendar) 的网站上爬取的数据进行开发的，所有番剧情报均和 Bangumi 上一致

| 指令                      | 效果                            |
|-------------------------|-------------------------------|
| 番剧 今日更新                 | 显示今日更新的番剧                     |
| 番剧 全部番剧                 | 显示当周全部番剧                      |
| 番剧 搜索 [参数]<br/>*参数数量>0* | 在新番更新表中查询番剧，返回包含搜索字段的结果，大小写敏感 |

![](readmePic/bangumiA.jpg)
![](readmePic/bangumiB.jpg)

#### 番剧查找
该功能爬取了三个网站：[NT 动漫网](http://www.ntyou.cc/)、[次元城动漫](https://www.cycdm01.top/)以及[动漫花园资源站](https://dongmanhuayuan.myheartsite.com/)

其中 NT 动漫网和次元城动漫是可以在线观看的网站，动漫花园资源站则只提供资源链接

| 指令                                          | 效果                |
|---------------------------------------------|-------------------|
| 我要看&#124;我现在就要看&#124;番剧搜索 [参数]<br/>*参数数量>0* | 在三个网站上搜索番剧相关资讯并返回 |

#### 配置参数
功能相关配置参数有：

```yaml
bangumi:
  bangumiSearch: true # 是否开启相关功能
  maxPoolSize: 3 # 同一时间的最大搜索数，所有组群共用
```

![](readmePic/bangumiSearchA.jpg)
![](readmePic/bangumiSearchB.jpg)
![](readmePic/bangumiSearchC.jpg)

### Ai Chat
目前为止使用了两个不同的源，一个是前段时间非常火的 [openAi](https://openai.com/)，另一个是国外 Ai 聊天网站 [characterAi](https://beta.character.ai/)

#### openAi
使用该功能要求在官网有供模型使用的 Token，注册一个账号有十多刀的免费额度。

功能相关配置如下：

```yaml
openAi:
  token: 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 #your token
  settings: # 详细配置见官方 api 文档
    chat:
      memory: false # 是否开启记忆化，不建议，token 烧太快
      model: text-davinci-003
      maxTokens: 1024
      temperature: 0.9
      topP: 1
      frequencyPenalty: 0
      presencePenalty: 0.6
    edit:
      model: code-davinci-edit-001
```

配置参数的名字和官方 API 请求参数基本一致，详细释义请查阅[官方 API 文档](https://beta.openai.com/docs/api-reference/completions)

>*上面这个 Token 当然是过期的*

| 指令                  | 效果                           |
|---------------------|------------------------------|
| /talk [聊天内容]        | 发送聊天内容                       |
| /talk 刷新            | 清除所有内容，包括预设                  |
| /talk 配置            | 显示当前配置                       |
| /talk 撤回            | 在允许记忆化的情况下有效，让 ai 重新回复你的上一句话 |
| /preset [预设内容]      | 添加预设                         |
| /edit [原始句子] [修改要求] | 按照要求修改句子                     |

目前记忆化有一点小问题，主要是 go 好像没有解析 openAi 内容 token 长度的库，所以可能会出现必须要刷新的情况，这个我后面想办法解决

还有就是，**开记忆化 token 烧的实在是太快了！慎开！**

>*如果你想他记住什么信息，更建议使用预设而不是记忆化*

![](readmePic/openAi.jpg)

#### CharacterAi
如果要使用该功能，需要挂代理，并在配置文件中正确填写 antiCf 相关的配置：

```yaml
antiCf: # 针对相关反爬的设定
  ja3: "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-51-57-47-53-10,0-23-65281-10-11-35-16-5-51-43-13-45-28-21,29-23-24-25-256-257,0"
  userAgent: "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:87.0) Gecko/20100101 Firefox/87.0"
  timeout: 200 # 单位s，设置太小了容易收不到回复，看个人情况调整
  proxy: "http://127.0.0.1:7890" # 代理地址，characterAi要翻墙所以必须挂
```

如果请求出现被拦截的情况，可以尝试更换 ja3 和 user-agent

| 指令              | 效果                               |
|-----------------|----------------------------------|
| [@ 机器人] 聊天      | 开启新的会话                           |
| 刷新人格列表          | 重新从 characterAi 配置文件中读取信息        |
| 刷新<br/>*会话中使用*  | 清空聊天记录                           |
| 重连<br/>*会话中使用*  | 不清空聊天记录，尝试重新连接                   |
| 换一句<br/>*会话中使用* | 更换另一个回答                          |
| 再见<br/>*会话中使用*  | 暂时结束会话，但记录仍旧存在，下次与相同人格对话可以继续上次聊天 |

人格列表配置文件格式如下：
```yaml
- botId: Bo7tKlIwU_BWQBxXjiZ9X65KtXtAVfMVAFKmw8M_vjg
  botName: Yusami
```
一定要按照这样填写为数组格式，botId 就是 url 后面那一串：

![](readmePic/botId.png)

**关于该功能还有几点要说的：**

- 为了避免有大聪明忘记关会话，bot 6 分钟内得不到下一条会话内容自动退出
- 该会话是私人的，彼此之间不影响
- 在等待回复时，bot 收到的所有消息都会主动丢弃

![](readmePic/characterAi.jpg)

*待续*