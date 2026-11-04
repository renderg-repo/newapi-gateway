# VideoFetchMode 数据迁移说明（PostgreSQL）

## 背景

`private_data.result_url` 以前在部分场景会写入网关自身的
`/v1/videos/{task_id}/content` 地址（`BuildProxyURL`）。这会让
`VideoProxy` 的 `direct_url` 分支回头请求自己，导致 Doubao 直连渠道出现
401/502。

本次改动后：

- `private_data.result_url` 只保存真实的上游/CDN 地址（`direct_url`）；
- `private_data.video_fetch_mode` 保存抓取模式：
  `direct_url` | `channel_download` | `provider_content` | `provider_resolve`；
- `channel_download`（移动云 `/aicc/`）不再需要 `result_url`；
- API 返回的 `result_url` 由 `relay.publicTaskResultURL` 在读取时按需
  拼成网关 `/v1/videos/{task_id}/content`，不再写回数据库。

字段位于 `private_data` JSON 内部，**不需要新增数据库列/GORM AutoMigrate**。

> 本文只提供 dry-run 与执行步骤，不在本任务中对生产库执行。

## 1. 盘点受影响任务（只读）

```sql
-- 所有 result_url 指向网关自身 content 接口的任务
SELECT id, task_id, platform, channel_id, status,
       private_data->>'result_url' AS result_url,
       private_data->>'video_fetch_mode' AS video_fetch_mode
FROM tasks
WHERE private_data->>'result_url' LIKE '%/v1/videos/%/content%'
ORDER BY id;

-- 按渠道类型聚合，确认范围
SELECT c.type, COUNT(*) AS affected
FROM tasks t
JOIN channels c ON c.id = t.channel_id
WHERE t.private_data->>'result_url' LIKE '%/v1/videos/%/content%'
GROUP BY c.type
ORDER BY c.type;
```

视频任务的 `tasks.platform` 存的是渠道类型字符串，Doubao 为 `'54'`。
移动云与直连火山都使用 `ChannelTypeDoubaoVideo`，通过渠道 `base_url`
是否包含 `/aicc/` 区分。

## 2. 备份（执行前必须）

```sql
CREATE TABLE tasks_video_fetch_mode_backup AS
SELECT id, private_data
FROM tasks
WHERE private_data->>'result_url' LIKE '%/v1/videos/%/content%';
```

## 3. 执行（在一个事务内）

### 3.1 Doubao 直连（无 `/aicc/`）：回填上游 URL + `direct_url`

数据来源为任务保存的上游响应 `data.content.video_url`（轮询时写入，
`redactVideoResponseBody` 不会删除该字段）。

```sql
BEGIN;

UPDATE tasks t
SET private_data = jsonb_set(
        jsonb_set(
            COALESCE(t.private_data::jsonb, '{}'::jsonb),
            '{result_url}',
            to_jsonb(t.data->'content'->>'video_url'),
            true
        ),
        '{video_fetch_mode}',
        '"direct_url"'::jsonb,
        true
    )
FROM channels c
WHERE c.id = t.channel_id
  AND c.type = 54
  AND t.platform = '54'
  AND COALESCE(c.base_url, '') NOT LIKE '%/aicc/%'
  AND t.private_data->>'result_url' LIKE '%/v1/videos/%/content%'
  AND COALESCE(t.data->'content'->>'video_url', '') <> '';

-- 无法从 data 回填的直连任务，先列出人工确认，不要直接清空
SELECT t.id, t.task_id, t.channel_id, t.data
FROM tasks t
JOIN channels c ON c.id = t.channel_id
WHERE c.type = 54
  AND t.platform = '54'
  AND COALESCE(c.base_url, '') NOT LIKE '%/aicc/%'
  AND t.private_data->>'result_url' LIKE '%/v1/videos/%/content%'
  AND COALESCE(t.data->'content'->>'video_url', '') = '';

COMMIT;
```

### 3.2 Doubao 移动云（`/aicc/`）：设置 `channel_download` 并清除自引用 URL

```sql
BEGIN;

UPDATE tasks t
SET private_data = (COALESCE(t.private_data::jsonb, '{}'::jsonb) - 'result_url')
    || jsonb_build_object('video_fetch_mode', 'channel_download')
FROM channels c
WHERE c.id = t.channel_id
  AND c.type = 54
  AND t.platform = '54'
  AND COALESCE(c.base_url, '') LIKE '%/aicc/%'
  AND t.private_data->>'result_url' LIKE '%/v1/videos/%/content%';

COMMIT;
```

> 不要在 `/aicc/` 渠道上把 `result_url` 回填成 `data.content.video_url`：
> 该地址是加密内容，必须由 `VideoProxy` 走 `/download` 解密。

## 4. 验证

```sql
-- Doubao 任务不应再残留网关自引用 URL
SELECT COUNT(*) AS remaining_self_proxy
FROM tasks
WHERE platform = '54'
  AND private_data->>'result_url' LIKE '%/v1/videos/%/content%';

-- 模式分布
SELECT private_data->>'video_fetch_mode' AS mode, COUNT(*)
FROM tasks
WHERE platform = '54'
GROUP BY 1
ORDER BY 1;

-- 移动云任务不应有 result_url
SELECT COUNT(*) AS aicc_with_result_url
FROM tasks t
JOIN channels c ON c.id = t.channel_id
WHERE c.type = 54
  AND COALESCE(c.base_url, '') LIKE '%/aicc/%'
  AND COALESCE(t.private_data->>'result_url', '') <> '';
```

预期：前两条中 Doubao 自引用计数为 0（无法回填的少数任务需人工处理）；
`aicc_with_result_url` 为 0。

## 5. 回滚

```sql
BEGIN;

UPDATE tasks t
SET private_data = b.private_data
FROM tasks_video_fetch_mode_backup b
WHERE b.id = t.id;

COMMIT;
```

确认回滚无误后再删除备份表：

```sql
DROP TABLE tasks_video_fetch_mode_backup;
```

## 6. 兼容性说明

- 迁移前创建的历史任务没有 `video_fetch_mode`，`VideoProxy` 会通过
  `taskcommon.ResolveVideoFetchMode` 按渠道类型与 `base_url` 实时推断，
  因此旧任务在迁移前也不会立即失效。
- 但对于旧的 **Doubao 直连** 任务，其 `result_url` 仍是网关自引用地址；
  在迁移完成前，`VideoProxy` 的 `direct_url` 分支会再次请求自身。
  上线前应先完成 3.1 的回填，或接受这部分任务暂时 502 并由迁移修复。
- `publicTaskResultURL` 只在 API 响应层拼网关地址，不会写回
  `private_data.result_url`。
