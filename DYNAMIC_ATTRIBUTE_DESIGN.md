# 动态属性系统设计文档

## 一、概述

### 1.1 背景

当前 OneOps 系统中主机资产的属性是固定的数据库列，每次新增属性需要修改数据库结构和代码，无法满足用户灵活自定义的需求。

### 1.2 目标

实现一个动态属性系统，允许管理员通过管理页面定义主机属性，用户创建主机时可以选择性地填写这些属性，无需修改代码和数据库结构。

### 1.3 核心概念

**动态属性系统** = 属性定义（元数据）+ 属性值（数据）

- **属性定义**：描述"有哪些属性可用"（管理员配置）
- **属性值**：描述"某台主机的某个属性的值是什么"（用户填写）

## 二、数据库设计

### 2.1 核心表结构

#### 2.1.1 属性定义表（attribute_definitions）

存储属性的元数据定义，由管理员通过管理页面配置。

```sql
CREATE TABLE attribute_definitions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '属性ID',
    name VARCHAR(100) NOT NULL COMMENT '属性名称（显示用）',
    key VARCHAR(50) NOT NULL COMMENT '属性键（唯一标识）',
    category VARCHAR(50) NOT NULL COMMENT '分类：system/location/environment/hardware/custom',
    type VARCHAR(20) NOT NULL DEFAULT 'text' COMMENT '类型：text/select/multiselect/number/date/boolean',
    options TEXT COMMENT '选项配置（JSON格式，select/multiselect类型使用）',
    required BOOLEAN DEFAULT FALSE COMMENT '是否必填',
    default_value VARCHAR(255) COMMENT '默认值',
    sort_order INT DEFAULT 0 COMMENT '排序（数字越小越靠前）',
    status TINYINT DEFAULT 1 COMMENT '状态：1启用 0禁用',
    description TEXT COMMENT '属性说明',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_key (key),
    KEY idx_category (category),
    KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='属性定义表';
```

#### 2.1.2 主机属性值表（server_attributes）

存储主机实例的属性值，与主机是多对一关系。

```sql
CREATE TABLE server_attributes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '属性ID',
    server_id BIGINT UNSIGNED NOT NULL COMMENT '主机ID',
    attribute_id BIGINT UNSIGNED NOT NULL COMMENT '属性定义ID',
    attribute_key VARCHAR(50) NOT NULL COMMENT '属性键（冗余字段）',
    attribute_value TEXT COMMENT '属性值',
    value_type VARCHAR(20) DEFAULT 'string' COMMENT '值类型：string/number/boolean/date/array',
    category VARCHAR(50) COMMENT '分类（冗余字段）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_server_attribute (server_id, attribute_id),
    KEY idx_server_id (server_id),
    KEY idx_attribute_id (attribute_id),
    KEY idx_key_value (attribute_key, attribute_value(255)),
    
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
    FOREIGN KEY (attribute_id) REFERENCES attribute_definitions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主机属性值表';
```

### 2.2 预置属性定义

系统初始化时，预置一些常用属性：

```sql
-- 业务系统
INSERT INTO attribute_definitions (name, `key`, category, type, options, sort_order) VALUES
('业务系统', 'business_system', 'system', 'select', '[{"value":"ecommerce","label":"电商系统"},{"value":"crm","label":"CRM系统"},{"value":"erp","label":"ERP系统"}]', 1);

-- 机房
INSERT INTO attribute_definitions (name, `key`, category, type, options, sort_order) VALUES
('机房', 'room', 'location', 'select', '[{"value":"hz","label":"杭州机房"},{"value":"bj","label":"北京机房"},{"value":"sh","label":"上海机房"}]', 2);

-- 机柜
INSERT INTO attribute_definitions (name, `key`, category, type, sort_order) VALUES
('机柜', 'cabinet', 'location', 'text', NULL, 3);

-- 环境
INSERT INTO attribute_definitions (name, `key`, category, type, options, default_value, sort_order) VALUES
('环境', 'env', 'environment', 'select', '[{"value":"prod","label":"生产"},{"value":"test","label":"测试"},{"value":"dev","label":"开发"}]', 'test', 4);

-- 标签
INSERT INTO attribute_definitions (name, `key`, category, type, options, sort_order) VALUES
('标签', 'tags', 'system', 'multiselect', '[{"value":"important","label":"重要"},{"value":"backup","label":"备份节点"},{"value":"monitor","label":"监控节点"}]', 5);

-- 所属项目
INSERT INTO attribute_definitions (name, `key`, category, type, sort_order) VALUES
('所属项目', 'project', 'system', 'text', NULL, 6);
```

### 2.3 数据模型（Go）

```go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AttributeDefinition 属性定义
type AttributeDefinition struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null"`              // 属性名称
	Key         string    `json:"key" gorm:"size:50;not null;uniqueIndex"`     // 属性键
	Category    string    `json:"category" gorm:"size:50;not null;index"`      // 分类
	Type        string    `json:"type" gorm:"size:20;not null;default:'text'"` // 类型
	Options     string    `json:"options" gorm:"type:text"`                    // 选项JSON
	Required    bool      `json:"required" gorm:"default:false"`               // 是否必填
	DefaultValue string   `json:"defaultValue" gorm:"size:255"`               // 默认值
	SortOrder   int       `json:"sortOrder" gorm:"default:0"`                  // 排序
	Status      int       `json:"status" gorm:"default:1"`                     // 状态
	Description string    `json:"description" gorm:"type:text"`                 // 说明
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// AttributeOptions 属性选项（用于 select/multiselect 类型）
type AttributeOptions []struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ServerAttribute 主机属性值
type ServerAttribute struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	ServerID      uint              `json:"serverId" gorm:"not null;index"`
	AttributeID   uint              `json:"attributeId" gorm:"not null;index"`
	AttributeKey  string            `json:"attributeKey" gorm:"size:50;not null;index"`
	AttributeValue string           `json:"attributeValue" gorm:"type:text"`
	ValueType     string            `json:"valueType" gorm:"size:20;default:'string'"`
	Category      string            `json:"category" gorm:"size:50"`
	CreatedAt     time.Time         `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt     time.Time         `json:"updatedAt" gorm:"autoUpdateTime"`
	
	// 关联
	Server        *Server           `json:"server,omitempty" gorm:"foreignKey:ServerID"`
	Definition    *AttributeDefinition `json:"definition,omitempty" gorm:"foreignKey:AttributeID"`
}

// TableName 指定表名
func (AttributeDefinition) TableName() string {
	return "attribute_definitions"
}

func (ServerAttribute) TableName() string {
	return "server_attributes"
}
```

## 三、API 设计

### 3.1 属性定义管理 API

#### 获取属性定义列表
```
GET /api/attributes

Response:
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "name": "业务系统",
      "key": "business_system",
      "category": "system",
      "type": "select",
      "options": "[{\"value\":\"ecommerce\",\"label\":\"电商系统\"}]",
      "required": false,
      "sortOrder": 1,
      "status": 1
    }
  ]
}
```

#### 创建属性定义
```
POST /api/attributes

Request:
{
  "name": "所属项目",
  "key": "project",
  "category": "system",
  "type": "text",
  "required": false,
  "sortOrder": 10,
  "description": "主机所属的项目"
}

Response:
{
  "code": 0,
  "message": "创建成功"
}
```

#### 更新属性定义
```
PUT /api/attributes/:id

Request:
{
  "name": "所属项目",
  "required": true
}

Response:
{
  "code": 0,
  "message": "更新成功"
}
```

#### 删除属性定义
```
DELETE /api/attributes/:id

Response:
{
  "code": 0,
  "message": "删除成功"
}
```

### 3.2 主机属性 API

#### 创建主机时保存属性
```
POST /api/servers

Request:
{
  "hostname": "web-01",
  "ip": "192.168.1.1",
  "groupIds": [1],
  "credentialIds": [1],
  "attributes": [
    {
      "attributeId": 1,
      "attributeKey": "business_system",
      "attributeValue": "ecommerce"
    },
    {
      "attributeId": 2,
      "attributeKey": "room",
      "attributeValue": "hz"
    }
  ]
}
```

#### 获取主机列表（包含属性）
```
GET /api/servers?page=1&pageSize=20

Response:
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "hostname": "web-01",
        "ip": "192.168.1.1",
        "attributes": [
          {
            "attributeKey": "business_system",
            "attributeValue": "ecommerce",
            "definition": {
              "name": "业务系统",
              "type": "select"
            }
          }
        ]
      }
    ],
    "total": 1
  }
}
```

#### 按属性筛选主机
```
GET /api/servers?business_system=ecommerce&room=hz

Response:
{
  "code": 0,
  "data": {
    "list": [...],
    "total": 10
  }
}
```

## 四、前端设计

### 4.1 属性管理页面

#### 页面布局
```
┌────────────────────────────────────────────────────────┐
│  属性管理                                    [+ 新增属性] │
├────────────────────────────────────────────────────────┤
│                                                        │
│  分类筛选: [全部] [系统分类] [地理位置] [环境信息]     │
│                                                        │
│  ┌──────────────────────────────────────────────────┐  │
│  │ 属性名称  │属性键   │分类  │类型  │必填│状态│操作│  │
│  ├──────────────────────────────────────────────────┤  │
│  │业务系统  │business │系统  │下拉  │否  │启用│编辑│  │
│  │机房      │room     │位置  │下拉  │否  │启用│编辑│  │
│  │机柜      │cabinet  │位置  │文本  │否  │启用│编辑│  │
│  │环境      │env      │环境  │下拉  │是  │启用│编辑│  │
│  └──────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────┘
```

#### 新增/编辑属性对话框
```
┌────────────────────────────────────────┐
│  新增属性                               │
├────────────────────────────────────────┤
│  属性名称: [所属项目          ]         │
│  属性键:   [project            ]         │
│  分类:     [系统分类 ▼]                  │
│  类型:     [文本输入 ▼]                  │
│  是否必填: [☐]                          │
│  默认值:   [                    ]         │
│  排序:     [1]                          │
│  说明:     [                    ]         │
│                                        │
│  [取消]              [保存]          │
└────────────────────────────────────────┘
```

### 4.2 主机表单改造

#### 表单布局
```
┌────────────────────────────────────────────────────┐
│  创建主机                                         │
├────────────────────────────────────────────────────┤
│                                                    │
│  基础信息 ▼                                       │
│  ┌──────────────────────────────────────────────┐  │
│  │ 主机名: [________________]                   │  │
│  │ 连接IP: [________________]                   │  │
│  │ 内网IP: [________________]                   │  │
│  │ 所属分组: [________________]               │  │
│  │ 连接凭证: [________________]               │  │
│  └──────────────────────────────────────────────┘  │
│                                                    │
│  资产属性 ▼                                       │
│  ┌──────────────────────────────────────────────┐  │
│  │ 业务系统: [电商系统 ▼]                       │  │
│  │ 机房:     [杭州机房 ▼]                       │  │
│  │ 机柜:     [____________]                       │  │
│  │ 环境:     [测试 ▼]                           │  │
│  │ 标签:     [重要] [备份节点]                  │  │
│  │ 所属项目: [____________]                       │  │
│  └──────────────────────────────────────────────┘  │
│                                                    │
│  更多配置 ▼                                       │
│  ┌──────────────────────────────────────────────┐  │
│  │ SSH端口: [22]                                │  │
│  │ 操作系统: [____________]                     │  │
│  │ 硬件配置: [CPU] [内存] [磁盘]                │  │
│  └──────────────────────────────────────────────┘  │
│                                                    │
│  [取消]                          [保存]         │
└────────────────────────────────────────────────────┘
```

## 五、属性类型说明

### 5.1 支持的属性类型

| 类型 | 说明 | 示例 | 选项格式 |
|------|------|------|----------|
| `text` | 单行文本输入 | 机柜号、项目名称 | 无 |
| `select` | 下拉单选 | 业务系统、机房 | `[{"value":"v1","label":"选项1"}]` |
| `multiselect` | 下拉多选 | 标签 | `[{"value":"v1","label":"选项1"}]` |
| `number` | 数字输入 | CPU核数、端口号 | 无 |
| `date` | 日期选择 | 购买日期、过保日期 | 无 |
| `boolean` | 布尔值 | 是否启用、是否备份 | 无 |

### 5.2 属性分类

| 分类 | 说明 | 示例 |
|------|------|------|
| `system` | 系统分类 | 业务系统、标签、所属项目 |
| `location` | 地理位置 | 机房、机柜、机架 |
| `environment` | 环境信息 | 环境（生产/测试/开发） |
| `hardware` | 硬件配置 | CPU型号、内存类型、磁盘类型 |
| `custom` | 自定义 | 用户自定义的任何属性 |

## 六、技术实现要点

### 6.1 属性值验证

```go
func (s *AttributeService) ValidateAttribute(attrID uint, value string) error {
    var def models.AttributeDefinition
    if err := db.First(&def, attrID).Error; err != nil {
        return err
    }
    
    // 必填验证
    if def.Required && value == "" {
        return fmt.Errorf("%s 不能为空", def.Name)
    }
    
    // 类型验证
    switch def.Type {
    case "number":
        if _, err := strconv.ParseFloat(value, 64); err != nil {
            return fmt.Errorf("%s 必须是数字", def.Name)
        }
    case "boolean":
        if value != "true" && value != "false" {
            return fmt.Errorf("%s 必须是是/否", def.Name)
        }
    case "select", "multiselect":
        var options []AttributeOptions
        json.Unmarshal([]byte(def.Options), &options)
        valid := false
        for _, opt := range options {
            if opt.Value == value {
                valid = true
                break
            }
        }
        if !valid {
            return fmt.Errorf("%s 的值不在有效范围内", def.Name)
        }
    }
    
    return nil
}
```

### 6.2 属性值查询优化

```go
// 按属性查询主机（使用索引）
func (s *ServerService) GetServersByAttribute(key, value string) ([]models.Server, error) {
    var servers []models.Server
    err := db.
        Table("servers").
        Joins("JOIN server_attributes sa ON servers.id = sa.server_id").
        Where("sa.attribute_key = ? AND sa.attribute_value = ?", key, value).
        Preload("Attributes").
        Preload("Attributes.Definition").
        Find(&servers).Error
    return servers, err
}

// 多属性组合查询
func (s *ServerService) GetServersByAttributes(filters map[string]string) ([]models.Server, error) {
    var servers []models.Server
    query := db.Table("servers")
    
    // 多次JOIN实现AND查询
    index := 0
    for key, value := range filters {
        alias := fmt.Sprintf("sa%d", index)
        query = query.Joins(
            fmt.Sprintf("JOIN server_attributes %s ON servers.id = %s.server_id", alias, alias),
        ).Where(fmt.Sprintf("%s.attribute_key = ? AND %s.attribute_value = ?", alias, alias), key, value)
        index++
    }
    
    err := query.Preload("Attributes").Find(&servers).Error
    return servers, err
}
```

### 6.3 前端动态表单渲染

```typescript
// 根据属性定义动态渲染表单项
function renderAttributeInput(attr: AttributeDefinition) {
  switch (attr.type) {
    case 'text':
      return h(ElInput, {
        modelValue: getAttributeValue(attr.key),
        placeholder: `请输入${attr.name}`,
        'onUpdate:modelValue': (val) => setAttributeValue(attr.key, val)
      })
    
    case 'select':
      const options = JSON.parse(attr.options || '[]')
      return h(ElSelect, {
        modelValue: getAttributeValue(attr.key),
        placeholder: `请选择${attr.name}`,
        'onUpdate:modelValue': (val) => setAttributeValue(attr.key, val)
      }, () => options.map(opt => 
        h(ElOption, { label: opt.label, value: opt.value })
      ))
    
    case 'multiselect':
      const options = JSON.parse(attr.options || '[]')
      return h(ElSelect, {
        modelValue: getAttributeValue(attr.key),
        multiple: true,
        placeholder: `请选择${attr.name}`,
        'onUpdate:modelValue': (val) => setAttributeValue(attr.key, val)
      }, () => options.map(opt => 
        h(ElOption, { label: opt.label, value: opt.value })
      ))
    
    case 'number':
      return h(ElInputNumber, {
        modelValue: getAttributeValue(attr.key),
        placeholder: `请输入${attr.name}`,
        'onUpdate:modelValue': (val) => setAttributeValue(attr.key, String(val))
      })
    
    case 'date':
      return h(ElDatePicker, {
        modelValue: getAttributeValue(attr.key),
        placeholder: `请选择${attr.name}`,
        'onUpdate:modelValue': (val) => setAttributeValue(attr.key, val)
      })
    
    case 'boolean':
      return h(ElSwitch, {
        modelValue: getAttributeValue(attr.key) === 'true',
        'onUpdate:modelValue': (val) => setAttributeValue(attr.key, String(val))
      })
  }
}
```

## 七、数据迁移方案

### 7.1 迁移步骤

#### 第一步：保留现有字段
```sql
-- 将现有固定列的数据迁移到动态属性表
INSERT INTO server_attributes (server_id, attribute_id, attribute_key, attribute_value, category)
SELECT 
    s.id,
    (SELECT id FROM attribute_definitions WHERE key = 'business_system'),
    'business_system',
    bu.name,
    'system'
FROM servers s
LEFT JOIN business_units bu ON s.business_id = bu.id
WHERE s.business_id IS NOT NULL;
```

#### 第二步：逐步移除固定列（可选）
```sql
-- 保留固定列一段时间，确保迁移成功
-- 后续版本再移除：ALTER TABLE servers DROP COLUMN business_id;
```

### 7.2 兼容性方案

```go
// 过渡期：同时支持固定列和动态属性
type Server struct {
    // 保留固定列（兼容性）
    BusinessID   uint   `json:"businessId" gorm:"index"`
    CabinetID    uint   `json:"cabinetId" gorm:"index"`
    
    // 新增动态属性
    Attributes   []ServerAttribute `json:"attributes,omitempty" gorm:"foreignKey:ServerID"`
}

// 读取时优先使用动态属性，回退到固定列
func (s *Server) GetBusinessSystem() string {
    // 优先从动态属性获取
    for _, attr := range s.Attributes {
        if attr.AttributeKey == "business_system" {
            return attr.AttributeValue
        }
    }
    // 回退到固定列
    if s.Business != nil {
        return s.Business.Name
    }
    return ""
}
```

## 八、性能优化

### 8.1 索引策略

```sql
-- 复合索引（key-value 查询）
CREATE INDEX idx_key_value ON server_attributes(attribute_key, attribute_value(255));

-- 覆盖索引（常用查询）
CREATE INDEX idx_server_key_value ON server_attributes(server_id, attribute_key, attribute_value(100));
```

### 8.2 缓存策略

```go
// 缓存属性定义（变更不频繁）
var attributeDefinitionsCache = make(map[uint]*AttributeDefinition)
var cacheMutex sync.RWMutex

func (s *AttributeService) GetAttributeDefinition(id uint) (*AttributeDefinition, error) {
    // 先读缓存
    cacheMutex.RLock()
    if def, ok := attributeDefinitionsCache[id]; ok {
        cacheMutex.RUnlock()
        return def, nil
    }
    cacheMutex.RUnlock()
    
    // 缓存未命中，查数据库
    var def AttributeDefinition
    if err := db.First(&def, id).Error; err != nil {
        return nil, err
    }
    
    // 更新缓存
    cacheMutex.Lock()
    attributeDefinitionsCache[id] = &def
    cacheMutex.Unlock()
    
    return &def, nil
}

// 属性定义变更时清除缓存
func (s *AttributeService) InvalidateCache() {
    cacheMutex.Lock()
    attributeDefinitionsCache = make(map[uint]*AttributeDefinition)
    cacheMutex.Unlock()
}
```

### 8.3 查询优化

```go
// 批量预加载属性
func (s *ServerService) GetServersWithAttributes(page, pageSize int) ([]Server, int64, error) {
    var servers []Server
    var total int64
    
    // 使用子查询优化，避免多次JOIN
    err := db.
        Preload("Attributes.Definition").
        Order("servers.id DESC").
        Offset((page - 1) * pageSize).
        Limit(pageSize).
        Find(&servers).Error
    
    db.Model(&Server{}).Count(&total)
    
    return servers, total, err
}
```

## 九、安全考虑

### 9.1 权限控制

```go
// 只有管理员可以管理属性定义
func (c *AttributeController) CreateAttribute(ctx *gin.Context) {
    // 检查管理员权限
    if !isAdmin(ctx) {
        ctx.JSON(403, gin.H{"error": "无权限"})
        return
    }
    // ...
}

func isAdmin(ctx *gin.Context) bool {
    user := ctx.GetBool("isAdmin")
    return user
}
```

### 9.2 输入验证

```go
// 防止SQL注入（使用参数化查询）
func (s *AttributeService) GetServersByAttribute(key, value string) ([]Server, error) {
    var servers []Server
    err := db.
        Where("? IN (SELECT attribute_value FROM server_attributes WHERE attribute_key = ?)", 
            value, key).
        Find(&servers).Error
    return servers, err
}

// 防止XSS（输出时转义）
func renderAttributeValue(value string) string {
    return html.EscapeString(value)
}
```

## 十、测试计划

### 10.1 单元测试

```go
func TestAttributeDefinitionCRUD(t *testing.T) {
    // 测试属性定义的增删改查
}

func TestAttributeValidation(t *testing.T) {
    // 测试属性值验证
}

func TestAttributeQuery(t *testing.T) {
    // 测试按属性查询主机
}
```

### 10.2 集成测试

1. 创建属性定义
2. 创建主机并填写属性
3. 查询主机列表，验证属性显示
4. 按属性筛选主机
5. 更新属性定义，验证表单变化
6. 删除属性定义，验证数据清理

### 10.3 性能测试

- 1000台主机，每台10个属性
- 查询性能测试
- 索引效果验证

## 十一、发布计划

### 11.1 阶段一：后端API开发（2-3天）
- 属性定义CRUD API
- 主机属性保存/查询API
- 属性验证逻辑

### 11.2 阶段二：前端开发（3-4天）
- 属性管理页面
- 主机表单改造
- 主机列表显示

### 11.3 阶段三：测试与优化（1-2天）
- 功能测试
- 性能测试
- Bug修复

### 11.4 阶段四：发布（1天）
- 数据库迁移脚本
- 发布说明
- 用户培训

**总计：7-10个工作日**
