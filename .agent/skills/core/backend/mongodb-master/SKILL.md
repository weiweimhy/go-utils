---
name: MongoDB Master
description: MongoDB 专家级技能，涵盖 Schema 设计、索引优化、聚合管道、分片策略、安全加固与运维监控
---

# Skill: MongoDB Master

## 🎯 触发条件

当以下情况发生时启用：

- "设计 MongoDB 的数据模型"
- "优化这个慢查询的索引"
- "写一个聚合管道 (Aggregation)"
- "规划 MongoDB 分片策略"

👉 自动启用本 Skill

## 🎯 Purpose

提供 MongoDB 专家级指导，帮助设计高性能 Schema、优化查询、规划分片策略、加固安全措施及日常运维。

## 🧩 Capabilities

### 1. Schema 设计

- **嵌入 vs 引用决策**：根据访问模式选择嵌入（1:Few/1:Many）或引用（1:Millions）
- **反范式化权衡**：牺牲存储换取读取性能，避免过度规范化
- **文档大小控制**：保持文档 < 16MB，避免无界数组增长
- **Schema 版本管理**：使用 `schemaVersion` 字段支持渐进式迁移

### 2. 索引策略

- **复合索引设计**：遵循 ESR 规则（Equality → Sort → Range）
- **覆盖索引**：使用投影使查询完全由索引满足
- **TTL 索引**：自动过期会话、日志等临时数据
- **部分索引**：仅索引符合条件的文档，减少存储开销
- **索引健康检查**：使用 `db.collection.explain()` 和 `$indexStats` 识别低效索引

### 3. Aggregation Pipeline

- **Pipeline 优化**：
  - 将 `$match` 和 `$project` 放在管道前端
  - 确保 `$match` 字段有索引支持（IXSCAN vs COLLSCAN）
  - 合并 `$sort` + `$limit` 优化大数据集排序
- **`$lookup` 最佳实践**：使用索引字段进行简单连接
- **内存限制**：超过 100MB 时启用 `allowDiskUse: true`

### 4. 分片 (Sharding)

- **Shard Key 选择**：
  - 高基数（避免热点分区）
  - 写分布均匀
  - 支持目标查询（避免 scatter-gather）
- **分片策略**：哈希分片 vs 范围分片
- **Balancer 管理**：在低峰期运行，监控 chunk 迁移

### 5. 安全加固

- **认证与授权**：启用 SCRAM-SHA-256，实施 RBAC 最小权限原则
- **加密**：
  - 传输层：启用 TLS/SSL
  - 存储层：使用 WiredTiger 加密
  - 字段级：使用 CSFLE (Client-Side Field Level Encryption)
- **审计**：启用审计日志追踪敏感操作
- **网络隔离**：使用 `bindIp` 限制访问，配合防火墙/VPN

### 6. 运维与监控

- **关键指标监控**：
  - 查询响应时间 / 慢查询
  - 复制延迟 (Replication Lag)
  - 连接池使用率
  - 内存 / CPU / 磁盘 I/O
- **工具推荐**：
  - MongoDB Atlas 内置监控
  - Percona Monitoring and Management (PMM)
  - `mongostat` / `mongotop` 命令行工具
- **备份策略**：使用 `mongodump` 或文件系统快照，定期验证恢复流程

## 🧠 Usage

当你需要：

- 设计新的 MongoDB 集合结构
- 优化慢查询或高延迟问题
- 规划从单节点迁移到分片集群
- 进行安全审计或合规检查
- 建立监控和告警体系

👉 启用本 Skill

## 📥 Input

- 业务需求描述
- 现有 Schema 或查询示例
- 性能问题描述（慢查询日志、explain 输出）
- 数据量预估和增长趋势

## 📤 Output

- Schema 设计建议（含嵌入/引用决策）
- 索引创建语句及理由
- Pipeline 优化建议
- 分片策略方案
- 安全加固 Checklist
- 监控指标和告警阈值建议

## 🔍 Source Mapping

| 能力 | 来源 | 说明 |
| :--- | :--- | :--- |
| Schema 设计 | MongoDB 官方文档 | ✅ 遵循 Data Modeling 最佳实践 |
| 索引策略 | MongoDB University | ✅ ESR 规则、索引选择器 |
| Aggregation | MongoDB 官方文档 | ✅ Pipeline 优化指南 |
| 分片 | MongoDB Sharding Manual | ✅ Shard Key 设计原则 |
| 安全 | Percona 安全白皮书 | ✅ 企业级安全实践 |
| 运维 | PMM / Atlas Docs | ✅ 监控和告警配置 |
| NoSQL 思维模型 | `nosql-expert` (外部) | ✅ Query-First 设计哲学 |

## 📚 References

- [MongoDB Data Modeling Introduction](https://www.mongodb.com/docs/manual/core/data-modeling-introduction/)
- [MongoDB Indexes](https://www.mongodb.com/docs/manual/indexes/)
- [Aggregation Pipeline Optimization](https://www.mongodb.com/docs/manual/core/aggregation-pipeline-optimization/)
- [MongoDB Sharding](https://www.mongodb.com/docs/manual/sharding/)
- [MongoDB Security Checklist](https://www.mongodb.com/docs/manual/administration/security-checklist/)
- [Percona MongoDB Security Best Practices](https://www.percona.com/blog/mongodb-security-best-practices/)
- [MongoDB 8.0 What's New](https://www.mongodb.com/docs/manual/release-notes/8.0/)

## ⚠️ Constraints

- ❌ 不替代官方 MongoDB 文档，复杂场景需查阅最新官方指南
- ❌ 不涵盖具体客户端驱动实现（Go/Node/Python 等）
- ❌ 不处理 MongoDB Atlas 特有的云服务配置（如 Data Lake、Atlas Search）

## 🔗 Related Skills

- `backend-patterns` - 后端架构模式
- `api-design` - API 设计规范
- `validation-lint` - 参数与配置校验
