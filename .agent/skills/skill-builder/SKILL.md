---
name: Skill 生成器
description: 帮助创建新的 Skill 定义，优先复用已有能力，避免重复造轮子
---

# Skill: Skill Builder（技能生成器）

帮助创建新的 Skill 定义，在创建前会进行能力拆解、已有技能检索、复用策略分析。

## 🎯 触发条件

当用户说：

- "帮我做一个 xxx 的 skill"
- "我想要一个可以 xxx 的能力"
- "这个能力该不该单独做成 skill"

👉 自动启用本 Skill

## 🧠 工作流程

### Step 1：能力拆解

解析用户输入，提取：

- **核心目标**：这个 Skill 要解决什么问题？
- **子能力列表**：需要哪些具体能力？
- **使用场景**：什么时候会用到？
- **输入 / 输出形式**：期望的输入输出是什么？

> [!TIP]
> 将复杂能力拆分成多个原子能力，有助于后续复用判断

### Step 2：检索已有技能（必须执行）

**检索顺序**（优先本地）：

1️⃣ **本地 skills 目录**（最优先）

```powershell
# turbo
# 查看当前项目已有的 Skills
Get-ChildItem -Path ".agent/skills" -Recurse -Filter "SKILL.md" | ForEach-Object { $_.FullName }
```

2️⃣ **antigravity-awesome-skills**
👉 <https://github.com/sickn33/antigravity-awesome-skills/tree/main>

3️⃣ **Skill Marketplace**
👉 <https://skillsmp.com/zh>

> [!TIP]
> 使用本项目自带的脚本访问网页：
>
> ```powershell
> # 使用配套脚本访问交互式网页
> python .agent/skills/skill-builder/scripts/web_access.py --url <URL>
> ```

**检索判断与整合**：

- 是否存在同名 Skill？
- 是否存在功能相同但命名不同的 Skill？
- 是否可以通过多个 Skill 组合实现？
- **外部能力对标**：对比外部优质 Skill（如 `antigravity-awesome-skills` 和 `Skill Marketplace`）中的功能点。
  - ✅ **必须整合**：如果外部 Skill 包含本地缺失的高价值能力（如数据库优化、安全实践等），必须将其整合进新 Skill 的设计中。
  - ❌ **严禁闭门造车**：禁止在明知有更全面的外部实现时，创建一个功能简陋的本地版本。

### Step 3：复用策略分析（非常重要）

对每个子能力，判断并填写来源映射表：

| 能力   | 来源                      | 复用策略                                |
| ------ | ------------------------- | --------------------------------------- |
| 能力 1 | 已有 skill 名称 或 "新建" | ✅ 直接复用 / 🔁 组合实现 / 🆕 需要新建 |

> [!IMPORTANT]
> 每个能力必须明确来源，不允许凭空创造

### Step 4：判断是否需要拆分

如果用户需求包含多个不相关的能力，应建议拆分成多个独立 Skill：

**拆分判断标准**：

- ❌ 一个 Skill 做了多件不同的事
- ❌ 能力之间没有逻辑关联
- ❌ 无法用一句话描述这个 Skill 的职责

**如需拆分**：

- 告知用户建议拆分的方案
- 分别为每个 Skill 执行本流程

### Step 5：输出 Skill 设计

如果确认需要创建新 Skill，必须输出以下结构：

```markdown
---
name: <中文名称>
description: <一句话描述>
---

# Skill: <skill-name>

## 🎯 Purpose

（一句话说明这个 skill 做什么）

## 🧩 Capabilities

- 能力 1
- 能力 2
- 能力 3

## 🔍 Source Mapping

（说明每个能力来自哪个已有 skill 或是否新建）

| 能力 | 来源 |
|------|------|
| 能力 1 | 来源说明 |

## 🧠 Usage

（什么时候用它）

## 📥 Input

（期望输入）

## 📤 Output

（输出内容格式）

## 📚 References

（列出参考的外部链接、技能或文档）

- 参考项 1：<URL 或 技能名称>
- 参考项 2：<URL 或 技能名称>

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 说明 |
| :--- | :--- | :--- |
| 流程设计 | `antigravity-system-prompt` | ✅ 遵循 Agent 核心协作哲学 |
| 复用策略 | `dreyfus-model-skill` | ✅ 参考 Dreyfus 技能习得模型 |
| 规范化输出 | `skill-marketplace-standard` | ✅ 对标外部技能市场规范 |

## 📚 参考资料 (References)

- [Antigravity Awesome Skills Guide](https://github.com/sickn33/antigravity-awesome-skills)
- [Prompt Engineering Guide](https://www.promptingguide.ai/)
- [The Dreyfus Model of Skill Acquisition](https://en.wikipedia.org/wiki/Dreyfus_model_of_skill_acquisition)

## ⚠️ Constraints

（使用限制 / 不做什么）

## 🔗 Related Skills

（可配合使用的其他 skill）
```

### Step 6：更新 SKILLS 清单

更新 `skills/SKILLS.md` 文件，将新创建的 Skill 添加到清单中。

## 📌 设计原则 Checklist

每次创建 Skill 前，必须确认以下原则：

- [ ] ❌ 不重复造轮子 —— 已检索本地和外部技能库
- [ ] ✅ 优先组合已有 skill —— 复用策略分析已完成
- [ ] ✅ 一个 skill 只做一件事 —— 职责单一，可用一句话描述
- [ ] ❌ 不写"万能型 skill" —— 没有包含多个不相关能力
- [ ] ✅ 所有能力可解释来源 —— Source Mapping 已填写

## ⚠️ 禁止事项

- ❌ 跳过检索直接创建新 Skill
- ❌ 创建功能与现有 Skill 重叠的 Skill
- ❌ 创建职责不清晰的"万能 Skill"
- ❌ 忽略本地已有 Skills 只看外部来源
