# Agent 模块设计

## 1. 模块目标

Agent 模块负责为每门课程提供面向课程资料的问答能力，目标如下：

1. 每门课程拥有独立 Agent。
2. 学生可以向课程 Agent 提问。
3. Agent 仅检索本课程资料。
4. Agent 回答可返回引用来源，增强可追溯性。
5. 教师可配置 Agent 的基础行为。

## 2. 模块边界

Agent 模块负责：

1. 课程 Agent 基础配置。
2. 问答会话管理。
3. 基于课程资料的检索问答。
4. 对话记录和引用来源保存。

Agent 模块不负责：

1. 用户登录。
2. 课程成员维护。
3. 文件上传和文件树维护。

## 3. 数据库表结构

### 3.1 课程 Agent 表 `course_agents`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT | 主键 |
| course_id | BIGINT | 课程 ID，唯一 |
| agent_name | VARCHAR(128) | Agent 名称 |
| prompt_template | TEXT | 系统提示词模板 |
| status | VARCHAR(16) | 状态，`enabled`/`disabled` |
| retrieval_scope | VARCHAR(16) | 检索范围，默认 `course_all` |
| created_by | BIGINT | 创建人 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 3.2 对话表 `agent_conversations`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT | 主键 |
| course_id | BIGINT | 课程 ID |
| agent_id | BIGINT | Agent ID |
| user_id | BIGINT | 发起用户 |
| conversation_title | VARCHAR(255) | 会话标题 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 3.3 消息表 `agent_messages`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT | 主键 |
| conversation_id | BIGINT | 会话 ID |
| sender_type | VARCHAR(16) | `user` 或 `agent` |
| message_content | TEXT | 消息内容 |
| token_usage | INT | 资源消耗记录，可选 |
| created_at | DATETIME | 创建时间 |

### 3.4 引用来源表 `agent_message_sources`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT | 主键 |
| message_id | BIGINT | Agent 回复消息 ID |
| material_node_id | BIGINT | 引用资料节点 ID |
| material_version_id | BIGINT | 引用版本 ID |
| snippet_text | TEXT | 命中片段摘要 |
| created_at | DATETIME | 创建时间 |

## 4. 参考 SQL

```sql
CREATE TABLE course_agents (
    id BIGINT PRIMARY KEY,
    course_id BIGINT NOT NULL UNIQUE,
    agent_name VARCHAR(128) NOT NULL,
    prompt_template TEXT,
    status VARCHAR(16) NOT NULL DEFAULT 'enabled',
    retrieval_scope VARCHAR(16) NOT NULL DEFAULT 'course_all',
    created_by BIGINT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (course_id) REFERENCES courses(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE agent_conversations (
    id BIGINT PRIMARY KEY,
    course_id BIGINT NOT NULL,
    agent_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    conversation_title VARCHAR(255),
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (course_id) REFERENCES courses(id),
    FOREIGN KEY (agent_id) REFERENCES course_agents(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE agent_messages (
    id BIGINT PRIMARY KEY,
    conversation_id BIGINT NOT NULL,
    sender_type VARCHAR(16) NOT NULL,
    message_content TEXT NOT NULL,
    token_usage INT,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (conversation_id) REFERENCES agent_conversations(id)
);

CREATE TABLE agent_message_sources (
    id BIGINT PRIMARY KEY,
    message_id BIGINT NOT NULL,
    material_node_id BIGINT NOT NULL,
    material_version_id BIGINT,
    snippet_text TEXT,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (message_id) REFERENCES agent_messages(id),
    FOREIGN KEY (material_node_id) REFERENCES course_material_nodes(id),
    FOREIGN KEY (material_version_id) REFERENCES course_material_versions(id)
);
```

## 5. 功能说明

### 5.1 课程 Agent 初始化

每门课程创建后，可自动创建一个默认 Agent 配置。

默认能力：

1. 回答课程资料相关问题。
2. 检索当前课程全部可用资料。
3. 返回资料引用来源。

### 5.2 Agent 配置管理

教师和创建者可修改：

1. Agent 名称。
2. 提示词模板。
3. 是否启用。
4. 检索范围策略。

### 5.3 发起问答

学生在课程上下文中向 Agent 提问，Agent 基于本课程资料检索并生成回答。

典型过程：

1. 校验提问者是否属于课程成员。
2. 校验 Agent 是否启用。
3. 检索本课程资料。
4. 生成回答。
5. 保存问题、回答和引用来源。

### 5.4 对话记录管理

系统按课程和用户保存问答会话，便于后续回看。

### 5.5 引用来源展示

回答中应返回引用资料，例如：

1. 文件名
2. 资料节点 ID
3. 摘要片段
4. 可跳转的预览地址

## 6. 接口设计

本模块所有接口响应都应遵循统一返回格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 6.1 获取 Agent 配置

接口：`GET /api/courses/{courseId}/agent`

功能：

1. 查看当前课程 Agent 配置。

实现要点：

1. 课程成员可查看基础配置。
2. 敏感配置可按角色裁剪展示，例如底层策略字段只对教师开放。

接口范围：

1. 只返回当前课程 Agent 信息。

权限规则：

1. 课程成员可调用。

### 6.2 更新 Agent 配置

接口：`PUT /api/courses/{courseId}/agent`

功能：

1. 修改 Agent 配置。

请求参数建议：

```json
{
  "agentName": "课程助教",
  "promptTemplate": "你是本课程的助教，只能基于课程资料回答问题。",
  "status": "enabled",
  "retrievalScope": "course_all"
}
```

实现要点：

1. 校验当前用户是否为教师或创建者。
2. 明确限制 Agent 只能使用课程内资料。

接口范围：

1. 仅修改课程 Agent 配置。

权限规则：

1. 教师、创建者可调用。

### 6.3 创建会话

接口：`POST /api/courses/{courseId}/agent/conversations`

功能：

1. 创建一个新的问答会话。

请求参数建议：

```json
{
  "title": "第一章复习"
}
```

实现要点：

1. 当前课程成员可创建。
2. 会话归属于当前用户和当前课程。

接口范围：

1. 仅创建会话容器。

权限规则：

1. 课程成员可调用。

### 6.4 获取会话列表

接口：`GET /api/courses/{courseId}/agent/conversations`

功能：

1. 查看问答会话列表。

实现要点：

1. 学生默认只看自己的会话。
2. 教师和创建者可查看课程内全部会话，或按参数筛选。

接口范围：

1. 只返回会话摘要。

权限规则：

1. 课程成员可调用，结果范围受角色限制。

### 6.5 获取会话详情

接口：`GET /api/courses/{courseId}/agent/conversations/{conversationId}`

功能：

1. 查看指定会话中的消息记录。

实现要点：

1. 学生只能查看自己的会话。
2. 教师和创建者可查看课程内任意会话。

接口范围：

1. 返回消息列表和引用来源。

权限规则：

1. 课程成员可调用，结果范围受角色限制。

### 6.6 提问

接口：`POST /api/courses/{courseId}/agent/ask`

功能：

1. 向课程 Agent 发起一次提问。

请求参数建议：

```json
{
  "conversationId": 9001,
  "question": "第一章中操作系统的定义是什么？"
}
```

响应内容建议：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "answer": "根据课程资料，操作系统是......",
    "sources": [
      {
        "materialNodeId": 2,
        "fileName": "第一章课件.pdf",
        "snippet": "操作系统是管理计算机硬件与软件资源的系统软件"
      }
    ]
  }
}
```

实现要点：

1. 校验用户是课程成员。
2. 校验会话属于当前课程且属于当前用户，或用户具有教师以上权限。
3. 检索范围必须限定为当前课程资料。
4. 保存用户问题、Agent 回答和引用来源。

接口范围：

1. 只处理课程内单次问答。
2. 不允许跨课程问答。

权限规则：

1. 学生、教师、创建者均可调用。
2. 非课程成员不可调用。

## 7. 权限规则

1. 学生：可发起提问，查看自己的会话。
2. 教师：可配置 Agent，查看课程内问答情况。
3. 创建者：拥有教师全部权限。

## 8. 实现注意事项

1. Agent 检索必须严格以课程 ID 作为过滤条件。
2. 被删除或禁用的资料不应进入默认检索范围。
3. 回答需要带来源，避免“黑盒回答”。
4. 如果后续需要控制成本，可增加提问频率限制、消息长度限制和会话归档策略。
