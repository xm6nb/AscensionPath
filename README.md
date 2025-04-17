# AscensionPath 漏洞环境管理系统

![Vue3](https://img.shields.io/badge/Vue-3.3-green)	![Golang](https://img.shields.io/badge/Go-1.20+-blue)

> 全栈漏洞环境管理系统，提供漏洞环境全生命周期管理，基于Docker容器技术实现安全隔离的实验环境

## 引言👀

  不知不觉学习网络安全已经三年，许多诸如笔者初涉此领域的学者，往往在环境配置的泥沼中耗尽热情，或在漏洞复现的迷宫里丢失方向。推出**AscensionPath**一个基于Docker容器化技术的漏洞环境管理系统，一方面是笔者巩固和学习Go语言，另一方面是让初学者从繁琐的环境配置中解放，为每一名网安实践者搭建起从理论到实战的跃迁之梯。相信在不断的实践训练中，那些曾被视作天堑的技术深水区，终将成为孕育彩虹的雨云。

  最后，笔者尚在学习，代码尚待优化请多多包含。如果您喜欢这个项目，请点一个免费的star⭐吧!

## 功能特性 ✨

- **环境管理**

  ✅ 镜像批量上传  ✅ 容器编排  ✅ 端口映射  ✅ 自动回收

- **权限控制**

  🔒 RBAC权限模型  🔒 JWT认证  🔒 操作审计

  👨💻 普通用户环境隔离  👨💻 Admin全局管理

- **开发支持**

  🛠️ Docker API集成  🛠️ Docker Compose多环境部署  🛠️ 端口智能映射

  📊 资源消耗统计  📊 实验时长分析

- **安全机制**
  🔐 文件上传验证  🔐 容器沙箱隔离  🔐 路径穿越防护
  🔐 会话加密     🔐 操作日志追踪

## 技术栈 🛠️

| 领域             | 技术选型                                                                                                                                         |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| **前端**   | [Art Design Pro ](https://www.lingchen.kim/art-design-pro/docs/guide/introduce.html)(Vue3 + TypeScript + Pinia + Element Plus + Vite) + tailwindcss |
| **后端**   | Golang 1.20 + Gin + Docker SDK + JWT + Zap日志                                                                                                   |
| **数据库** | SQLite                                                                                                                                           |
| **部署**   | 可执行程序一键部署                                                                                                                               |

## 系统架构 📦

```txt
src/
├── frontend/             # 前端
│   ├── public/           # 静态资源
│   └── src/
│       ├── assets/       # 资源文件
│       ├── components/   # 全局组件
│       ├── router/       # 路由配置
│       ├── store/        # Pinia状态管理
│       └── views/        # 页面视图
│
└── backend/              # 后端
    ├── cmd/app           # 入口文件
    ├── internal/
    │   ├── handler/      # HTTP控制器
    │   ├── service/      # 业务逻辑
    │   ├── model/        # 数据模型
    │   ├── middleware    # 中间件
    │   └── utils         # 工具
    └── config/           # 配置管理
```

## 快速启动 🚀

### 开发环境

```bash
# 前端
cd frontend
pnpm install
npm run dev

# 后端
cd backend
go mod tidy
go run ./cmd/app/main.go
```

### 生产部署

```bash
cd frontend
npm run build

cd ../backend/cmd/app/
go build main.go

# 端口可选，默认8080
./main [-port xxxx]

# 在浏览器打开即可
```

默认账号密码：admin:123456

## 使用说明🌈

### 镜像

在当前的系统中可以上传**镜像列表.json**文件，也可以上传**Docker compose项目**的压缩包。但要注意的是，列表**.json**有以下格式要求：

- rank：对该漏洞的评分

```json
[
    {
        "image_name": "镜像名称(拉取地址)",
        "image_vul_name": "漏洞镜像环境的名称(可自定义)",
        "image_desc": "对于该漏洞镜像环境的描述(可自定义)",
        "rank": 1,
        "degree": {
            "HoleType": [
                "命令执行","SQL注入","..."
            ],
            "devLanguage": [
                "Java","..."
            ],
            "devClassify": [
                "Gin","..."
            ],
            "DevDatabase": [
                "MYSQL","..."
            ]
        },
        "from": "镜像来源(可自定义)"
    }
]
```

也可以直接在机器上拉取镜像，系统会自动读取本地已有的镜像。

> 笔者已经整理好了Vulfocus的相关镜像，部署成功后上传即可。

### Docker compose

  **Docker compose**项目的压缩包就比较自由了，只需要上传ZIP压缩包即可。系统会自动解压到 `./storeate`目录下并自动寻找该目录下所有的 `docker-compose.yml`文件进行加载，`docker-compose.yml`文件在的目录名会自动作为漏洞环境的名称。在上传的时候要注意多次上传ZIP压缩包内容的相对路径最好别相同防止覆盖。

  系统并没有对ZIP进行任何限制，你甚至可以将**Vulhub**项目打包上传。

### Pull image

  拉取镜像与系统默认的docker源有关，由于国内的网络条件，请注意调整网络网络环境。

### 创建一个实例

1. 拉取镜像 或 上传**Docker compose**压缩包
2. 创建漏洞环境
3. 创建实例

## 权限管理 👮

| 角色  | 权限说明     | 功能范围                   |
| ----- | ------------ | -------------------------- |
| Admin | 系统管理员   | 用户管理/镜像审核/全局监控 |
| User  | 普通实验用户 | 实例操作/个人空间          |

## 安全设计 🔒

| 安全特性   | 实现方式                            |
| ---------- | ----------------------------------- |
| Token加密  | HS256签名算法 + 动态密钥管理        |
| 会话安全   | 18小时自动过期机制 + 服务端状态校验 |
| 防篡改机制 | 签名验证 + 标准Claim校验            |
| 密钥管理   | uuid随机密钥 + 每天凌晨2点自动刷新  |

## 致谢 🤝

- [关于 Art Design Pro | Art-Design-Pro](https://www.lingchen.kim/art-design-pro/docs/guide/introduce.html)
- [vulhub/vulhub: Pre-Built Vulnerable Environments Based on Docker-Compose](https://github.com/vulhub/vulhub)
- [fofapro/vulfocus: 🚀Vulfocus 是一个漏洞集成平台，将漏洞环境 docker 镜像，放入即可使用，开箱即用。](https://github.com/fofapro/vulfocus)

## 声明

若有侵权，请联系我！
