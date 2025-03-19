# VM Alert Rule Generator

一个基于Gin框架开发的Web服务，用于生成和管理VictoriaMetrics Alert规则。该服务支持根据配置动态生成告警规则组，并符合vmalert的URL自动发现规则。

## 功能特性

- 基于Gin框架的轻量级Web服务
- 支持动态生成告警规则组
- 符合vmalert URL自动发现规则
- 支持通过配置文件自定义规则模板
- 提供Docker容器化部署支持

## 快速开始

### 本地运行

1. 克隆项目到本地

2. 安装依赖
```bash
go mod download
```

3. 运行服务
```bash
go run main.go
```

### Docker部署

1. 构建Docker镜像
```bash
docker build -t registry.kbsonlong.com/library/vmalert-rules-server .
```

2. 运行容器
```bash
docker run -p 8080:8080 -v $(pwd)/template.yaml:/app/template.yaml registry.kbsonlong.com/library/vmalert-rules-server
```

## 配置说明

### template.yaml

`template.yaml`文件用于配置告警规则模板，服务会根据此配置生成对应的告警规则组。

配置示例：
```yaml
groups:
  - name: "example_group"
    rules:
      - alert: "ExampleAlert"
        expr: "up == 0"
        for: "5m"
        labels:
          severity: "critical"
        annotations:
          summary: "Instance {{ $labels.instance }} down"
          description: "{{ $labels.instance }} has been down for more than 5 minutes."
```

## API文档

### 获取告警规则列表

**GET /api/rules**

获取所有告警规则或按组名筛选规则或按组名筛选规则。
**查询参数：**
- `group_name` (可选): 按组名筛选规则


**响应示例：**
```json
[
  {
    "id": 1,
    "name": "HighCPUUsage",
    "alert": "HighCPUUsage",
    "expr": "100 - (avg by(instance) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100) > 80",
    "for": "5m",
    "labels": "{\"severity\":\"warning\"}",
    "annotations": "{\"description\":\"CPU使用率超过80%\",\"summary\":\"服务器CPU使用率过高\"}",
    "group_name": "system_metrics",
    "enabled": true,
    "created_at": "2023-12-01T10:00:00Z",
    "updated_at": "2023-12-01T10:00:00Z"
  }
]
```

### 创建告警规则

**POST /api/rules**

创建新的告警规则。

**请求体：**
```json
{
  "name": "HighMemoryUsage",
  "alert": "HighMemoryUsage",
  "expr": "(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100 > 90",
  "for": "5m",
  "labels": {
    "severity": "warning"
  },
  "annotations": {
    "description": "内存使用率超过90%",
    "summary": "服务器内存使用率过高"
  },
  "group_name": "system_metrics",
  "enabled": true
}
```

**响应示例：**
```json
{
  "id": 2,
  "name": "HighMemoryUsage",
  "alert": "HighMemoryUsage",
  "expr": "(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100 > 90",
  "for": "5m",
  "labels": "{\"severity\":\"warning\"}",
  "annotations": "{\"description\":\"内存使用率超过90%\",\"summary\":\"服务器内存使用率过高\"}",
  "group_name": "system_metrics",
  "enabled": true,
  "created_at": "2023-12-01T11:00:00Z",
  "updated_at": "2023-12-01T11:00:00Z"
}
```

### 更新告警规则

**PUT /api/rules/:id**

更新指定ID的告警规则。

**请求体：**
```json
{
  "name": "HighMemoryUsage",
  "alert": "HighMemoryUsage",
  "expr": "(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100 > 95",
  "for": "10m",
  "labels": {
    "severity": "critical"
  },
  "annotations": {
    "description": "内存使用率超过95%",
    "summary": "服务器内存使用率严重过高"
  },
  "group_name": "system_metrics",
  "enabled": true
}
```

**响应示例：**
```json
{
  "id": 2,
  "name": "HighMemoryUsage",
  "alert": "HighMemoryUsage",
  "expr": "(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100 > 95",
  "for": "10m",
  "labels": "{\"severity\":\"critical\"}",
  "annotations": "{\"description\":\"内存使用率超过95%\",\"summary\":\"服务器内存使用率严重过高\"}",
  "group_name": "system_metrics",
  "enabled": true,
  "created_at": "2023-12-01T11:00:00Z",
  "updated_at": "2023-12-01T12:00:00Z"
}
```

### 删除告警规则

**DELETE /api/rules/:id**

删除指定ID的告警规则。

**响应示例：**
```json
{
  "message": "规则删除成功"
}
```

### 健康检查

**GET /health**

服务健康检查接口。

**响应示例：**
```json
{
  "status": "ok"
}
```

## 注意事项

1. 所有API请求需要在Header中设置`Content-Type: application/json`
2. 创建和更新规则时，labels和annotations字段支持JSON格式的字符串
3. 规则名称在同一组内必须唯一
4. 时间间隔(for)的格式必须符合Prometheus duration格式（如：5m, 1h, 1d等）