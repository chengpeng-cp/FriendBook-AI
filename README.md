# FriendBook AI —— 智能好友管理系统

> 一个前后端分离的实战教学项目：用 **Go + Gin + GORM** 搭后端，用 **React + TypeScript + TailwindCSS** 搭前端，并接入 **AI Agent（Eino）** 实现自然语言操作数据。

**配套教学视频持续更新中，欢迎关注 👇**

[![Bilibili](https://img.shields.io/badge/B站-关注UP主-FF69B4?style=flat-square&logo=bilibili&logoColor=white)](https://space.bilibili.com/481791559)

| 📺 B站视频空间 | 💬 微信 | 📕 小红书 |
|:---:|:---:|:---:|
| [点击跳转](https://space.bilibili.com/481791559) | <img src="assets/images/vx.png" width="160" alt="微信二维码"/> | <img src="assets/images/redbook.png" width="160" alt="小红书二维码"/> |
| 从零到一完整教程 | 毕设咨询、项目咨询、项目合作加好友 | 图文及视频版笔记 |

---

## 目录

- [项目介绍](#项目介绍)
- [技术栈](#技术栈)
- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [项目结构](#项目结构)

## 项目介绍

FriendBook AI 是一个「好友关系管理」系统，不只是普通的 CRUD：

- **经典三层后端**：Handler → Service → Repository，配合 Gin + GORM + MySQL + Redis + JWT，把 Go Web 开发的核心知识点全部走一遍
- **AI 能力实战**：基于字节的 **Eino** 框架搭建 ReAct Agent，在聊天里直接对好友数据做增删改查（"帮我给张三加 5 点好感度"）
- **MCP Server**：把好友管理封装成 MCP 工具，任何支持 MCP 的客户端（如 Claude、Trae）都能直连你的数据库
- **现代前端**：React 18 + TypeScript + TailwindCSS + Zustand + GSAP，数据可视化图表开箱即用

跟着做一遍，你可以完整掌握一套 Go 全栈项目的开发流程。

## 技术栈

### 后端

| 技术 | 版本 | 说明 |
|---|---|---|
| Go | 1.23+ | |
| Gin | v1.10 | Web 框架 |
| GORM | v1.25 | ORM（AutoMigrate 自动建表） |
| MySQL | 8.0 | 主数据库 |
| Redis | 7.0 | 缓存 |
| golang-jwt | v5 | 登录鉴权 |
| Viper | v1.19 | 配置管理（支持环境变量覆盖） |
| Zap | v1.27 | 日志 |
| Eino | v0.7 | AI Agent 框架（字节跳动开源） |
| mcp-go | v0.26 | MCP Server |
| Swagger | swaggo | 接口文档 |

### 前端

| 技术 | 版本 | 说明 |
|---|---|---|
| React | 18.3 | |
| TypeScript | 5.5 | |
| Vite | 5.4 | 构建工具（dev 端口 5173） |
| TailwindCSS | 3.4 | 样式 |
| Zustand | 4.5 | 状态管理 |
| React Router | 6.26 | 路由 |
| GSAP | 3.12 | 动画 |
| Axios | 1.7 | 请求库 |

## 功能特性

- 用户注册 / 登录（JWT 鉴权，bcrypt 密码加密）
- 好友 CRUD + 标签 + 分组管理
- 互动记录：聊天 / 见面 / 礼物，带照片附件
- 好感度与「关系温度」系统，自动计算关系等级
- 定时提醒（纪念日、生日）+ Cron 定时任务
- 数据看板：好友分布、互动趋势、温度排行等图表
- AI 助手：自然语言直接操作好友数据（ReAct Agent）
- MCP Server：对外开放工具协议接口
- 限流中间件、统一响应格式、统一错误码、Zap 日志

## 快速开始

### 环境要求

| 工具 | 版本要求 |
|---|---|
| Go | 1.23+ |
| Node.js | 18+ |
| MySQL | 8.0 |
| Redis | 7.0（可选，需要时才用） |

> 没有 MySQL/Redis 环境？直接跳到 [Docker 一键启动](#方式三docker-一键启动仅后端)。

### 方式一：本地运行（推荐跟着视频做）

#### 1. 克隆项目

```bash
git clone https://github.com/chengpeng-cp/FriendBook-AI.git golang-learn
cd golang-learn
```

#### 2. 启动后端

```bash
cd backend

# ① 复制配置模板（真实配置已被 .gitignore 排除，不会提交到仓库）
cp configs/config_example.yaml configs/config.yaml

# ② 编辑 configs/config.yaml，改三处：
#    - mysql.host：本地环境改成 127.0.0.1，填上你的数据库密码
#    - redis.host：本地环境改成 127.0.0.1
#    - openai.api_key：填你的 API Key（OpenRouter / OpenAI 均可）
#      AI 功能不用的话可以先不填，其余功能不受影响

# ③ 安装依赖并启动（默认监听 8080 端口）
go mod download
go run cmd/server/main.go
```

看到日志输出 `Server started on :8080` 即启动成功。数据表由 GORM AutoMigrate 自动创建，**无需手动执行 SQL**。

#### 3. 启动前端

```bash
cd frontend

npm install
npm run dev
```

浏览器打开 `http://localhost:5173`，注册一个账号即可开始体验。前端请求通过 Vite 代理转发到 8080，**无需任何额外配置**。

### 方式二：Makefile 运行（后端）

```bash
cd backend
make run      # 等同于 go run cmd/server/main.go
make swag     # 重新生成 Swagger 文档（需先安装 swag）
make test     # 运行单元测试 + 覆盖率
```

### 方式三：Docker 一键启动（仅后端）

后端 + MySQL + Redis 一把梭，适合不想装数据库的同学：

```bash
cd backend
docker-compose up -d --build
```

默认拉起 `app(8080) + mysql(3306) + redis(6379)` 三个容器，MySQL 数据持久化在卷 `mysql_data` 中。

### 常见问题

<details>
<summary>端口被占用？</summary>

后端改 `configs/config.yaml` 的 `server.port`；前端改 `frontend/vite.config.ts` 的 `server.port`，同时把代理目标 `http://localhost:8080` 改成后端实际端口。

</details>

<details>
<summary>AI 助手没反应？</summary>

检查 `configs/config.yaml` 的 `openai.api_key` 是否有效、`base_url` 是否与 Key 匹配（OpenRouter 的 Key 就填 `https://openrouter.ai/api/v1`）。AI 功能是可选的，不影响其他功能使用。

</details>

## 项目结构

```
golang-learn/
├── backend/                      # Go 后端
│   ├── cmd/server/main.go        # 入口：启动流程编排
│   ├── configs/                  # 配置文件（config.yaml 已被 gitignore）
│   │   ├── config_example.yaml   # 配置模板（clone 后复制此文件）
│   │   └── config.yaml           # 本地真实配置（不入库）
│   ├── internal/
│   │   ├── ai/                   # AI Agent（Eino ReAct + 工具定义）
│   │   ├── config/               # Viper 配置加载
│   │   ├── handler/              # HTTP 处理层（参数校验 → 调 Service）
│   │   ├── initialize/           # 初始化：DB / Redis / 依赖容器 / Cron
│   │   ├── mcpserver/            # MCP Server（工具协议对外暴露）
│   │   ├── middleware/           # 中间件：JWT / CORS / 限流 / 日志 / Recovery
│   │   ├── model/                # GORM 数据模型
│   │   ├── pkg/                  # 通用包：错误码 / 日志 / 响应 / 工具
│   │   ├── repository/           # 数据访问层（只跟数据库打交道）
│   │   ├── router/               # 路由注册
│   │   └── service/              # 业务逻辑层（事务在这里管）
│   ├── docs/                     # Swagger 文档（make swag 生成）
│   ├── Dockerfile
│   ├── docker-compose.yml
│   ├── Makefile
│   ├── go.mod / go.sum
├── frontend/                     # React 前端
│   ├── src/
│   │   ├── api/                  # Axios 封装 + 各模块接口
│   │   ├── components/           # 通用组件 / 图表 / 布局
│   │   ├── pages/                # 页面：看板 / 好友 / 互动 / 提醒 / 统计 / AI 助手…
│   │   ├── stores/               # Zustand 状态（auth / toast）
│   │   ├── types/                # TS 类型定义
│   │   ├── hooks/ utils/ constants/
│   │   └── App.tsx / main.tsx
│   ├── index.html
│   ├── package.json
│   ├── tailwind.config.js
│   └── vite.config.ts            # dev 代理：/api /mcp → localhost:8080
└── README.md
```

## 关于本课程

本项目是我的 Go 全栈实战系列课程的配套源码，每个教学视频对应一次提交。

- 📺 **B站**：[https://space.bilibili.com/481791559](https://space.bilibili.com/481791559) —— 视频教程主阵地
- 💬 **微信**：扫码加好友，进学员交流群

  <img src="assets/images/vx.png" width="160" alt="微信二维码"/>

- 📕 **小红书**：图文版干货笔记

  <img src="assets/images/redbook.png" width="160" alt="小红书二维码"/>

有问题欢迎在仓库提 Issue，或到群里交流。如果本项目对你有帮助，欢迎 Star ⭐ 支持！
