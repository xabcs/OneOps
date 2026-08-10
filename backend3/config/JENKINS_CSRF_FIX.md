# Jenkins CSRF Token 问题修复

## 问题描述

```
Error 403 No valid crumb was included in the request
```

## 根本原因

Jenkins 的 CSRF 保护机制需要 **Cookie + Crumb** 双重验证：

1. 获取 Crumb 时，Jenkins 会设置一个会话 Cookie
2. 执行脚本时，必须同时发送：
   - 相同的 Cookie（会话标识）
   - Crumb Token（CSRF Token）

**之前的错误：**
- 获取 Crumb 和执行脚本使用了不同的 `http.Client`
- 每个 Client 有自己独立的 Cookie 存储
- Jenkins 收到脚本请求时，发现没有对应的 Cookie 会话，拒绝请求

## 修复方案

### 修复前代码

```go
// ❌ 错误方式：两个独立的 Client
func getJenkinsCrumb() {
    client := &http.Client{}  // Client 1
    client.Do(request)        // 获取 Crumb
}

func executeJenkinsScript() {
    client := &http.Client{}  // Client 2（不同实例）
    client.Do(request)        // ❌ 没有 Cookie，403 错误
}
```

### 修复后代码

```go
// ✅ 正确方式：共享同一个 Client 和 CookieJar
func executeJenkinsScript(baseURL, username, password, script string) (string, error) {
    // 1. 创建 CookieJar
    jar, err := cookiejar.New(nil)
    if err != nil {
        return "", fmt.Errorf("创建 Cookie Jar 失败: %w", err)
    }

    // 2. 创建共享 Cookie 的 HTTP Client
    client := &http.Client{
        Jar: jar,  // ✅ 使用同一个 CookieJar
    }

    // 3. 获取 CSRF Token（Cookie 自动保存到 jar）
    crumbReq, _ := http.NewRequest("GET", crumbURL, nil)
    crumbReq.SetBasicAuth(username, password)
    crumbResp, _ := client.Do(crumbReq)  // ✅ Cookie 已保存

    // 解析 Crumb
    var crumb JenkinsCrumb
    json.Unmarshal(crumbBody, &crumb)

    // 4. 执行脚本（自动携带相同的 Cookie）
    req, _ := http.NewRequest("POST", scriptURL, strings.NewReader(data))
    req.SetBasicAuth(username, password)
    req.Header.Set("Jenkins-Crumb", crumb.Crumb)  // ✅ 发送 Crumb

    resp, _ := client.Do(req)  // ✅ 自动携带 Cookie，成功！

    return string(body), nil
}
```

## 技术原理

### HTTP 请求流程

```
步骤 1: GET /crumbIssuer/api/json
┌─────────────────────────────┐
│ Request:                    │
│   Authorization: Basic xxx  │
│                             │
│ Response:                   │
│   Set-Cookie: JSESSIONID=abc│ ← Cookie 自动保存到 jar
│   Body: {"crumb": "123"}    │
└─────────────────────────────┘

步骤 2: POST /scriptText
┌─────────────────────────────┐
│ Request:                    │
│   Cookie: JSESSIONID=abc    │ ← ✅ 自动从 jar 中获取
│   Jenkins-Crumb: 123        │
│   Authorization: Basic xxx  │
│                             │
│ Response:                   │
│   200 OK                    │ ← ✅ 成功！
│   Body: 脚本执行结果        │
└─────────────────────────────┘
```

### CookieJar 工作原理

```go
type CookieJar interface {
    // 自动保存响应中的 Set-Cookie
    SetCookies(u *url.URL, cookies []*http.Cookie)

    // 自动为请求添加 Cookie 头
    Cookies(u *url.URL) []*http.Cookie
}

// 使用过程：
client := &http.Client{Jar: jar}

// 第一个请求
client.Do(request1)  // 响应中的 Cookie 自动保存到 jar

// 第二个请求（相同域名）
client.Do(request2)  // jar 自动添加 Cookie 到请求头
```

## 修复验证

### 测试命令

```bash
# 1. 获取 Crumb 并保存 Cookie
curl -u admin:'password' \
  -c /tmp/cookies.txt \
  'http://jenkins.hzmeipingmi.com/crumbIssuer/api/json'

# 2. 使用相同的 Cookie 执行脚本
curl -u admin:'password' \
  -b /tmp/cookies.txt \
  -X POST \
  -H "Jenkins-Crumb: $(cat /tmp/cookies.txt | grep Jenkins-Crumb | awk '{print $NF}')" \
  -d "script=println 'Hello'" \
  http://jenkins.hzmeipingmi.com/scriptText
```

### 预期结果

**修复前：**
```
❌ 403 No valid crumb was included in the request
```

**修复后：**
```
✅ 200 OK
Hello
Result: [null]
```

## 代码变更摘要

**修改文件：** `backend/services/application_permission_service.go`

**关键变更：**

1. 添加导入：
```go
import "net/http/cookiejar"
```

2. 修改 `executeJenkinsScript` 函数：
   - 创建 CookieJar
   - 创建共享 Cookie 的 HTTP Client
   - 在同一个 Client 中获取 Crumb 和执行脚本
   - 删除独立的 `getJenkinsCrumb` 函数

## 为什么之前可以测试成功？

在命令行测试时，我们使用了：
```bash
# 使用 -c 保存 cookie，-b 使用 cookie
curl -c cookies.txt ...  # 保存
curl -b cookies.txt ...  # 使用
```

但在 Go 代码中，之前使用了两个不同的 `http.Client` 实例，导致 Cookie 不共享。

## 总结

**问题：** CSRF Token 验证失败（403）

**原因：** Cookie 不共享，Jenkins 无法验证会话

**解决：** 使用 `cookiejar` 共享 Cookie

**状态：** ✅ 已修复

---

**测试方法：**

重新部署后端，点击"同步用户"或"同步角色"，应该可以成功执行。
