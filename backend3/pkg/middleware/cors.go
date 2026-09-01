package middleware

import (
	"net/http"
	"strings"
	"time"

	"oneops/backend3/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 依据配置构造跨域中间件配置。
// 安全约束：
//   - 默认白名单模式（AllowOrigins 精确匹配），不再硬编码 AllowAllOrigins=true——
//     全开来源 + AllowCredentials 组合会让任意站点携带凭证跨域调用 API，属 CSRF 面
//   - 仅当白名单显式包含 "*" 时才全开，且此时强制关闭 AllowCredentials（CORS 规范禁止 * 与凭证并用）
func CORS(cfg config.CORSConfig) cors.Config {
	c := cors.DefaultConfig()
	c.AllowMethods = cfg.AllowMethods
	c.AllowHeaders = cfg.AllowHeaders
	c.ExposeHeaders = cfg.ExposeHeaders
	if cfg.MaxAge > 0 {
		c.MaxAge = time.Duration(cfg.MaxAge) * time.Second
	}

	for _, o := range cfg.AllowOrigins {
		if o == "*" {
			c.AllowAllOrigins = true
			c.AllowCredentials = false
			return c
		}
	}

	c.AllowOrigins = cfg.AllowOrigins
	c.AllowCredentials = cfg.AllowCredentials
	return c
}

// CORSMiddleware 组装 CORS 中间件，按 CORS 规范区分预检与非预检请求：
//   - WebSocket 升级请求跳过 CORS：浏览器对 WS 强制携带 Origin 头，但 WS 升级不受
//     ACAO 响应头约束（CORS 模型不适用），WS 鉴权由连接层 token 独立承担；若按 HTTP
//     语义校验，localhost/127.0.0.1 或 IP/域名混布的前端访问源与 API Host 不一致时
//     会被误判为跨域而 403，导致 Pod 终端等 WS 功能不可用
//   - 预检请求（OPTIONS + Access-Control-Request-Method）：Origin 不在白名单 → 403，
//     浏览器将不发送实际跨域请求，跨域读取控制在预检阶段完成
//   - 非预检请求：Origin 不在白名单 → 仅不返回 ACAO 响应头并放行（规范语义，浏览器
//     自会拒绝读取响应）；不得直接 403——开发代理（如 127.0.0.1:9527 同源转发）的
//     同源请求也携带 Origin 头，且转发后 Origin 与 API Host 必然不同源，直接 403 会
//     误杀所有走代理的同源流量。跨域读取防护由"无 ACAO 头 + 预检拦截"共同保证
func CORSMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	c := CORS(cfg)
	preflightHandler := cors.New(c)
	return func(ctx *gin.Context) {
		if strings.EqualFold(ctx.GetHeader("Upgrade"), "websocket") {
			ctx.Next()
			return
		}

		origin := ctx.GetHeader("Origin")
		if origin == "" {
			ctx.Next() // 无 Origin 头：非浏览器跨域场景（同源 GET/服务间调用），与 CORS 无关
			return
		}

		// 预检请求：交由 gin-cors 处理（白名单匹配返回 204，不匹配 403 终止）
		if ctx.Request.Method == http.MethodOptions && ctx.GetHeader("Access-Control-Request-Method") != "" {
			preflightHandler(ctx)
			return
		}

		// 非预检请求：匹配白名单则补 ACAO 头（浏览器跨域可读），不匹配仅不加头放行
		if originAllowed(c, origin) {
			ctx.Header("Access-Control-Allow-Origin", origin)
			ctx.Header("Vary", "Origin")
			if c.AllowCredentials {
				ctx.Header("Access-Control-Allow-Credentials", "true")
			}
			if len(c.ExposeHeaders) > 0 {
				ctx.Header("Access-Control-Expose-Headers", strings.Join(c.ExposeHeaders, ", "))
			}
		}
		ctx.Next()
	}
}

// originAllowed 判断 Origin 是否命中白名单：全开、精确匹配或 "*.example.com" 通配子域
func originAllowed(c cors.Config, origin string) bool {
	if c.AllowAllOrigins {
		return true
	}
	for _, allowed := range c.AllowOrigins {
		if allowed == origin {
			return true
		}
		if strings.HasPrefix(allowed, "*.") {
			// 通配子域：*.example.com 匹配 a.example.com，不匹配 example.com 本身
			suffix := allowed[1:] // ".example.com"
			host := origin
			if i := strings.Index(host, "://"); i >= 0 {
				host = host[i+3:]
			}
			if strings.HasSuffix(host, suffix) {
				return true
			}
		}
	}
	return false
}
