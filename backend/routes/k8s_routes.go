package routes

import (
	"oneops/backend/controller"
	"oneops/backend/handler"
	"oneops/backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupK8sRoutes 设置 K8s 管理相关路由
func SetupK8sRoutes(
	r *gin.Engine,
	k8sClusterController *controller.K8sClusterController,
	k8sPermissionController *controller.K8sPermissionController,
	k8sResourceController *controller.K8sResourceController,
	k8sTerminalHandler *handler.K8sTerminalHandler,
	diagnosticController *controller.DiagnosticController,
) {
	api := r.Group("/api")

	// K8s Pod 终端 WebSocket（不经过 Auth 中间件，由 handler 自行验证）
	api.GET("/k8s/terminal/ws", func(ctx *gin.Context) {
		k8sTerminalHandler.HandleWebSocket(ctx)
	})

	// K8s 集群管理路由（需要认证）
	k8s := api.Group("/k8s")
	k8s.Use(middleware.Auth())
	k8s.Use(middleware.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		// 集群管理
		k8s.GET("/clusters", k8sClusterController.GetClusters)
		k8s.POST("/clusters", k8sClusterController.CreateCluster)
		k8s.GET("/clusters/:id", k8sClusterController.GetClusterByID)
		k8s.PUT("/clusters/:id", k8sClusterController.UpdateCluster)
		k8s.DELETE("/clusters/:id", k8sClusterController.DeleteCluster)

		// 连接测试
		k8s.POST("/clusters/:id/test", k8sClusterController.TestConnection)

		// 集群节点和命名空间
		k8s.GET("/clusters/:id/nodes", k8sClusterController.GetClusterNodes)
		k8s.GET("/clusters/:id/namespaces", k8sClusterController.GetClusterNamespaces)

		// 集群用户管理
		k8s.GET("/clusters/:id/users", k8sClusterController.GetClusterUsers)

		// 权限管理
		k8s.POST("/clusters/:id/permissions", k8sPermissionController.AssignClusterRole)
		k8s.DELETE("/clusters/:id/permissions/:userId", k8sPermissionController.RevokeClusterRole)
		k8s.GET("/users/clusters", k8sPermissionController.GetUserClusters)
		k8s.GET("/clusters/:id/users/:userId/role", k8sPermissionController.GetUserRoleInCluster)
		k8s.POST("/permissions/batch-assign", k8sPermissionController.BatchAssignClusterRoles)

		// Workloads - Deployments
		k8s.GET("/clusters/:id/deployments", k8sResourceController.ListDeployments)
		k8s.GET("/clusters/:id/deployments/:namespace/:name", k8sResourceController.GetDeployment)
		k8s.GET("/clusters/:id/deployments/:namespace/:name/pods", k8sResourceController.GetDeploymentPods)
		k8s.POST("/clusters/:id/deployments", k8sResourceController.CreateDeployment)
		k8s.PUT("/clusters/:id/deployments", k8sResourceController.UpdateDeployment)
		k8s.DELETE("/clusters/:id/deployments", k8sResourceController.DeleteDeployment)
		k8s.POST("/clusters/:id/deployments/scale", k8sResourceController.ScaleDeployment)
		k8s.POST("/clusters/:id/deployments/restart", k8sResourceController.RestartDeployment)

		// Workloads - StatefulSets
		k8s.GET("/clusters/:id/statefulsets", k8sResourceController.ListStatefulSets)
		k8s.GET("/clusters/:id/statefulsets/:namespace/:name", k8sResourceController.GetStatefulSet)
		k8s.GET("/clusters/:id/statefulsets/:namespace/:name/pods", k8sResourceController.GetStatefulSetPods)

		// Workloads - DaemonSets
		k8s.GET("/clusters/:id/daemonsets", k8sResourceController.ListDaemonSets)
		k8s.GET("/clusters/:id/daemonsets/:namespace/:name", k8sResourceController.GetDaemonSet)
		k8s.GET("/clusters/:id/daemonsets/:namespace/:name/pods", k8sResourceController.GetDaemonSetPods)

		// Workloads - Jobs
		k8s.GET("/clusters/:id/jobs", k8sResourceController.ListJobs)
		k8s.GET("/clusters/:id/jobs/:namespace/:name", k8sResourceController.GetJob)
		k8s.GET("/clusters/:id/jobs/:namespace/:name/pods", k8sResourceController.GetJobPods)
		k8s.DELETE("/clusters/:id/jobs", k8sResourceController.DeleteJob)

		// Workloads - CronJobs
		k8s.GET("/clusters/:id/cronjobs", k8sResourceController.ListCronJobs)
		k8s.GET("/clusters/:id/cronjobs/:namespace/:name", k8sResourceController.GetCronJob)
		k8s.GET("/clusters/:id/cronjobs/:namespace/:name/pods", k8sResourceController.GetCronJobPods)
		k8s.DELETE("/clusters/:id/cronjobs", k8sResourceController.DeleteCronJob)
		k8s.PUT("/clusters/:id/cronjobs/suspend", k8sResourceController.SuspendCronJob)

		// Services
		k8s.GET("/clusters/:id/services", k8sResourceController.ListServices)
		k8s.GET("/clusters/:id/services/:namespace/:name", k8sResourceController.GetService)
		k8s.POST("/clusters/:id/services", k8sResourceController.CreateService)
		k8s.PUT("/clusters/:id/services", k8sResourceController.UpdateService)
		k8s.DELETE("/clusters/:id/services", k8sResourceController.DeleteService)

		// Ingresses
		k8s.GET("/clusters/:id/ingresses", k8sResourceController.ListIngress)
		k8s.GET("/clusters/:id/ingresses/:namespace/:name", k8sResourceController.GetIngress)
		k8s.POST("/clusters/:id/ingresses", k8sResourceController.CreateIngress)
		k8s.PUT("/clusters/:id/ingresses", k8sResourceController.UpdateIngress)
		k8s.DELETE("/clusters/:id/ingresses", k8sResourceController.DeleteIngress)

		// Pods
		k8s.GET("/clusters/:id/pods", k8sResourceController.ListPods)
		k8s.GET("/clusters/:id/pods/:namespace/:name", k8sResourceController.GetPod)
		k8s.PUT("/clusters/:id/pods", k8sResourceController.UpdatePod)
		k8s.GET("/clusters/:id/pods/:namespace/:name/logs", k8sResourceController.GetPodLogs)
		k8s.DELETE("/clusters/:id/pods", k8sResourceController.DeletePod)

		// ConfigMaps
		k8s.GET("/clusters/:id/configmaps", k8sResourceController.ListConfigMaps)
		k8s.GET("/clusters/:id/configmaps/:namespace/:name", k8sResourceController.GetConfigMap)
		k8s.POST("/clusters/:id/configmaps", k8sResourceController.CreateConfigMap)
		k8s.PUT("/clusters/:id/configmaps", k8sResourceController.UpdateConfigMap)
		k8s.DELETE("/clusters/:id/configmaps", k8sResourceController.DeleteConfigMap)

		// Secrets
		k8s.GET("/clusters/:id/secrets", k8sResourceController.ListSecrets)
		k8s.GET("/clusters/:id/secrets/:namespace/:name", k8sResourceController.GetSecret)
		k8s.POST("/clusters/:id/secrets", k8sResourceController.CreateSecret)
		k8s.PUT("/clusters/:id/secrets", k8sResourceController.UpdateSecret)
		k8s.DELETE("/clusters/:id/secrets", k8sResourceController.DeleteSecret)

		// Events
		k8s.GET("/clusters/:id/events", k8sResourceController.ListEvents)

		// K8s 终端管理
		k8s.GET("/terminal/active", k8sTerminalHandler.GetActiveSessions)
		k8s.POST("/terminal/sessions/:sessionId/terminate", k8sTerminalHandler.TerminateSession)

		// K8s 诊断功能
		k8s.GET("/diagnostic/commands", diagnosticController.GetDiagnosticCommands)
		k8s.GET("/diagnostic/pods/:clusterId/:namespace", diagnosticController.GetJavaPods)
		k8s.GET("/diagnostic/namespaces/:clusterId", diagnosticController.GetNamespaces)
		k8s.POST("/diagnostic/execute", diagnosticController.ExecuteDiagnostic)
		k8s.GET("/diagnostic/history", diagnosticController.GetDiagnosticHistory)

	}
}
