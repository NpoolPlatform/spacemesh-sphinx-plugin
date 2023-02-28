# Npool go service app template

[![Test](https://github.com/NpoolPlatform/sphinx-plugin/actions/workflows/main.yml/badge.svg?branch=master)](https://github.com/NpoolPlatform/sphinx-plugin/actions/workflows/main.yml)

[目录](#目录)

- [Npool go service app template](#npool-go-service-app-template)
  - [新增币种](#新增币种)
    - [新增功能](#新增功能)
  - [功能](#功能)
  - [命令](#命令)
  - [最佳实践](#最佳实践)
  - [环境变量](#环境变量)
    - [钱包地址格式](#钱包地址格式)
    - [交易上链状态查询默认周期](#交易上链状态查询默认周期)
    - [wallet-status-check](#wallet-status-check)
    - [account-check](#account-check)
    - [升级说明](#升级说明)
    - [推荐](#推荐)
    - [说明](#说明)
  - [优化](#优化)

-----------

## [新增币种](./newcoin.md)

### 新增功能

- [x] 自定义调度周期
- [x] 自定义错误处理
- [ ] 链路追踪
- [ ] 监控
- [ ] 币种单位转换统一处理
- [x] 上报meta信息到proxy
- [ ] 优化配置
- [ ] 相同地址的并发处理
- [ ] payload 记录在 redis
- [ ] 动态调整 **gas fee**
- [ ] 支持多 **pod** 部署
- [x] 连接wallet节点时检测同步状态
- [ ] 支持同一个链plugin下可配置所有token的子集
- [ ] CID查询链上交易状态
- [ ] 提供获取秘钥（文件）接口
- [ ] Proxy提供查询plugin对应信息接口

新币种的支持步骤

1. 配置新币种单位和名称
2. 必须要实现的接口
3. 注册新币种
4. 设置默认 SyncTime

## 功能

- [x] 将服务部署到k8s集群
- [x] 将服务 api 通过 traefik-internet ingress 代理，供外部应用调用(视服务功能决定是否需要)

## 命令

- make init ```初始化仓库，创建go.mod```
- make verify ```验证开发环境与构建环境，检查code conduct```
- make verify-build ```编译目标```
- make test ```单元测试```
- make generate-docker-images ```生成docker镜像```
- make sphinx-plugin ```单独编译服务```
- make sphinx-plugin-image ```单独生成服务镜像```
- make deploy-to-k8s-cluster ```部署到k8s集群```

## 最佳实践

- 每个服务只提供单一可执行文件，有利于 docker 镜像打包与 k8s 部署管理
- 每个服务提供 http 调试接口，通过 curl 获取调试信息
- 集群内服务间 direct call 调用通过服务发现获取目标地址进行调用

## 环境变量

| 币种              | 变量名称               | 支持的值       | 说明                                                                          |
|:------------------|:-----------------------|:---------------|:------------------------------------------------------------------------------|
| Comm              | ENV_COIN_NET           | main or test   |                                                                               |
|                   | ENV_COIN_TYPE          |                |                                                                               |
|                   | ENV_SYNC_INTERVAL      |                | optional,交易状态同步间隔周期(s)                                              |
|                   | ENV_WAN_IP             |                | plugin的wan-ip                                                                |
|                   | ENV_POSITION           |                | plugin的位置信息(如NewYork_NO2)                                               |
|                   | ENV_COIN_LOCAL_API     |                | 多个地址使用,分割                                                             |
|                   | ENV_COIN_PUBLIC_API    |                | 多个地址使用,分割                                                             |
|                   | ENV_WAN_IP             |                | 上报网络IP                                                                    |
|                   | ENV_POSITION           |                | 上报位置信息如:HongKong-05                                                    |
| SmartContractCoin | ENV_CONTRACT           |                | 合约币的合约地址(对于主网合约地址已硬编码,测试网需要指定为自己部署的合约地址) |
| Ethereum          | ENV_BUILD_CHAIN_SERVER | host:grpc_port | 用于eth的plugin在test环境下获取测试合约地址                                   |

配置说明

对于合约地址配置说明

### 钱包地址格式

钱包地址配置格式:
  **'url|auth,url|auth,url|auth'**
注意：所有地址需要用引号引起来，地址间用逗号分割

单个地址格式：

- 格式1 不需要认证

  ````conf
    auth 格式
    示例: https://127.0.0.1:8080
  ````

- 格式2 账号密码体系

  ````conf
    auth 格式
    user@password
    示例: https://127.0.0.1:8080|root@3306
  ````

- 格式3 token 体系

  ````conf
    auth 格式
    token
    示例: https://127.0.0.1:8080|token
  ````

| 格式  | 链               | 说明 |
|-----|------------------|------|
| 格式1 | sol bsc eth tron |      |
| 格式2 | btc              |      |
| 格式3 | fil              |      |

### 交易上链状态查询默认周期
以下表格也是所有类型plugin的列表
|              币种              | 默认值 | 出块时间 |
|:------------------------------:|:------:|:--------:|
|            spacemesh            |  30s   |   60s    |

### wallet-status-check

钱包状态检查，检查节点高度是否与链高度一致
tron链的币种暂无

### account-check

账户验证

### 升级说明

- **需要关闭用户购买商品的入口**
- **失败可以重试, 成功操作不可重试**
- **注意 SQL 只更新了 filecoin 和 bitcoin 币种，其余可参考 filecoin 和 bitcoin, tfilecoin 和 tbitcoin 上报完成才可以执行**

| 条件    | 升级 SQL                     |
|:--------|:-----------------------------|
| mainnet | DO NOTHING                   |
| testnet | [upgrade](./sql/upgrade.sql) |

### 推荐

bitcoin 钱包节点的配置文件中, **rpcclienttimeout=30** 需要配置

### 说明

- 不支持 **Windows**

## 优化

- 镜像多阶段构建
- 尝试关闭 **CGO_ENABLE**
