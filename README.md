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

## 使用方法

1. 启动服务后，可以通过以下方式访问：
   - 规则列表：`http://localhost:8080/rules`
   - 健康检查：`http://localhost:8080/health`

2. 服务支持通过启动参数配置规则组数量和每组规则数：
```bash
./vm-server -groups 5 -rules 10
```

## API文档

### GET /rules
获取所有告警规则组

**响应示例：**
```json
{
  "groups": [
    {
      "name": "group_1",
      "rules": [...]
    }
  ]
}
```

### GET /health
服务健康检查接口

**响应示例：**
```json
{
  "status": "ok"
}
```

## 注意事项

1. 确保template.yaml文件格式正确
2. 在生产环境中建议使用Docker部署
3. 合理配置规则组数量和规则数，避免资源占用过高

## 许可证

MIT License