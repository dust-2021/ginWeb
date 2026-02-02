## 虚拟局域网联机服务器

基于wireguard实现的虚拟组网工具，需配合客户端使用。

### 运行方式：
1. 已有数据库，修改配置文件后启动：
   参照`config-example.yaml`配置
```yaml
# 需要修改的配置项
# 虚拟局域网网段
vlan: [10, 20]
# 32位字符串用于密码盐和
secret: "0eb67000000e86948ce2a73da05b96e1"

# 管理员账号，目前只需配置用户名密码
adminUser:
  username: "mole"
  phone: "13900000000"
  email: "xxx@qq.com"
  password: "123456"

database:
# 是否生成所有数据表
  initial: true
  link: "username:password@tcp(127.0.0.1:3306)/ginweb"
  poolSize: 10
  maxConnect: 5
redis:
  host: "localhost"
  port: 6379
```
直接运行`exampleApp`可执行文件，因为修改了三方库源码，下载项目编译会失败，`config.yaml`放在同级目录，运行需要root权限。

2. 使用`docker`启动：
   ```shell
      # 构建应用镜像
      docker build -t molev1.0 .
      # 创建数据库镜像后启动
      docker-compose up -d
   ```

客户端项目：https://github.com/dust-2021/mole.git
