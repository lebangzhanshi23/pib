# PIB - Personal Interview Brain

面试复习闭环系统：LLM 结构化输入 + 间隔重复算法 (SM-2)

## 功能

- 📝 结构化输入：丢入面试笔记，AI 提取 Q&A
- 🧠 间隔重复：基于 SM-2 算法智能安排复习时间
- 🏷️ 标签管理：按标签筛选题目
- 📊 复习记录：追踪学习进度

## 快速开始

```bash
# 启动服务
go run ./cmd/server

# API 端口: 8081
```

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/questions | 创建题目 |
| GET | /api/v1/questions | 获取题目列表 |
| GET | /api/v1/questions/review | 获取待复习题目 |
| POST | /api/v1/questions/:id/review | 提交复习 (grade: 0-2) |
| DELETE | /api/v1/questions/:id | 删除题目 |

## Grade 说明

| Grade | 含义 | 间隔变化 |
|-------|------|----------|
| 0 | Forgot (忘记) | 重置为 1 天 |
| 1 | Vague (模糊) | × 1.2 |
| 2 | Remembered (记住) | × EF |

## 配置

修改 `config/config.yaml`:
```yaml
app:
  port: 8081

llm:
  provider: deepseek
  api_key: ${DEEPSEEK_API_KEY}
  model: deepseek-chat
```
