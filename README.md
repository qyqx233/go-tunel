# go-tunel

go-tunel 是一个能将内网机器端口映射到外网访问的一个程序(前提是内网机有访问外网权限

### 前提准备

一台具有固定IP的外网服务器,比方说各种xx云服务器

### 程序工作流程

![image](https://github.com/qyqx233/go-tunel/blob/master/res/Diagram-1.png)


首先有三个角色,内网机，部署`inner`程序;云服务器，部署`outer`程序; 用户端`Client`安装Xshell等终端工具

`Inner`会主动与外网机器建立一个命令通道`Cmd`,建立成功后,外网端会选择一个端口(也可以配置文件中指定),记为`P`

用户端需要通过访问云服务器去访问内网，云服务器在收到请求后会在命令通道`Cmd`上发起一个请求,`Inner`收到后会向外网端主动建立一个链接`C`,`outer`将端口`P`的`socket`与`C`绑定，整条链路就通了，用户端就可以通过云服务器去访问内网了

### 编译，部署

编译很简单，执行 `outer/cmd/mk.sh` 将生成的`outer`和`outer.toml`放到云服务器上

执行`inner/cmd/mk.sh` 将生成的`inner`和`inner.toml`放到内网端

#### 配置

假设云服务器IP `could.x.x.x`，监听`3333`端口，内网机公网IP `host.x.x.x`，要将`127.0.0.1:443`端口映射出来，那么可以这么配置


inner.toml
```toml
[Auth]
Name = "test1"
Symkey = "0123456789012345"

[[Transport]] # 可以映射多个服务
TargetHost = "127.0.0.1"
TargetPort = 443
ServerIp = "cloud.x.x.x"
ServerPort = 3333
```

outer.toml
```toml
[[Transport]]
Ip = "host.x.x.x"
TargetHost = "127.0.0.1"
TargetPort = 443
Symkey = "0123456789012345"
LocalPort = 13001

[CmdServer]
Port = 3333
```

其他一些配置可以看配置文件里面的注释