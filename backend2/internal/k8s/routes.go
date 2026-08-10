package k8s

import (
	"oneops/backend2/internal/system"
	"oneops/backend2/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// SetupK8sRoutes 设置 K8s 管理相关路由
func SetupK8sRoutes(
	r *gin.Engine,
	k8sClusterController *K8sClusterController,
	k8sPermissionController *K8sPermissionController,
	k8sResourceController *K8sResourceController,
	k8sTerminalHandler *K8sTerminalHandler,
	diagnosticController *DiagnosticController,
) {
	api := r.Group("/api")

	// K8s Pod 终端 WebSocket（不经过 Auth 中间件，由 handler 自行验证）
	api.GET("/k8s/terminal/ws", func(ctx *gin.Context) {
		k8sTerminalHandler.HandleWebSocket(ctx)
	})

	// K8s 集群管理路由（需要认证）
	k8s := api.Group("/k8s")
	Use(middleware.Auth())
	Use(system.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		// 集群管理
		GET("/clusters", k8sClusterController.GetClusters)
		POST("/clusters", k8sClusterController.CreateCluster)
		GET("/clusters/:id", k8sClusterController.GetClusterByID)
		PUT("/clusters/:id", k8sClusterController.UpdateCluster)
		DELETE("/clusters/:id", k8sClusterController.DeleteCluster)

		// 连接测试
		POST("/clusters/:id/test", k8sClusterController.TestConnection)

		// 集群节点和命名空间
		GET("/clusters/:id/nodes", k8sClusterController.GetClusterNodes)
		GET("/clusters/:id/namespaces", k8sClusterController.GetClusterNamespaces)

		// 集群用户管理
		GET("/clusters/:id/users", k8sClusterController.GetClusterUsers)

		// 权限管理
		POST("/clusters/:id/permissions", k8sPermissionController.AssignClusterRole)
		DELETE("/clusters/:id/permissions/:userId", k8sPermissionController.RevokeClusterRole)
		GET("/users/clusters", k8sPermissionController.GetUserClusters)
		GET("/clusters/:id/users/:userId/role", k8sPermissionController.GetUserRoleInCluster)
		POST("/permissions/batch-assign", k8sPermissionController.BatchAssignClusterRoles)

		// Workloads - Deployments
		GET("/clusters/:id/deployments", k8sResourceController.ListDeployments)
		GET("/clusters/:id/deployments/:namespace/:name", k8sResourceController.GetDeployment)
		GET("/clusters/:id/deployments/:namespace/:name/pods", k8sResourceController.GetDeploymentPods)
		POST("/clusters/:id/deployments", k8sResourceController.CreateDeployment)
		PUT("/clusters/:id/deployments", k8sResourceController.UpdateDeployment)
		DELETE("/clusters/:id/deployments", k8sResourceController.DeleteDeployment)
		POST("/clusters/:id/deployments/scale", k8sResourceController.ScaleDeployment)
		POST("/clusters/:id/deployments/restart", k8sResourceController.RestartDeployment)

		// Workloads - StatefulSets
		GET("/clusters/:id/statefulsets", k8sResourceController.ListStatefulSets)
		GET("/clusters/:id/statefulsets/:namespace/:name", k8sResourceController.GetStatefulSet)
		GET("/clusters/:id/statefulsets/:namespace/:name/pods", k8sResourceController.GetStatefulSetPods)

		// Workloads - DaemonSets
		GET("/clusters/:id/daemonsets", k8sResourceController.ListDaemonSets)
		GET("/clusters/:id/daemonsets/:namespace/:name", k8sResourceController.GetDaemonSet)
		GET("/clusters/:id/daemonsets/:namespace/:name/pods", k8sResourceController.GetDaemonSetPods)

		// Workloads - Jobs
		GET("/clusters/:id/jobs", k8sResourceController.ListJobs)
		GET("/clusters/:id/jobs/:namespace/:name", k8sResourceController.GetJob)
		GET("/clusters/:id/jobs/:namespace/:name/pods", k8sResourceController.GetJobPods)
		DELETE("/clusters/:id/jobs", k8sResourceController.DeleteJob)

		// Workloads - CronJobs
		GET("/clusters/:id/cronjobs", k8sResourceController.ListCronJobs)
		GET("/clusters/:id/cronjobs/:namespace/:name", k8sResourceController.GetCronJob)
		GET("/clusters/:id/cronjobs/:namespace/:name/pods", k8sResourceController.GetCronJobPods)
		DELETE("/clusters/:id/cronjobs", k8sResourceController.DeleteCronJob)
		PUT("/clusters/:id/cronjobs/suspend", k8sResourceController.SuspendCronJob)

		// Services
		GET("/clusters/:id/services", k8sResourceController.ListServices)
		GET("/clusters/:id/services/:namespace/:name", k8sResourceController.GetService)
		POST("/clusters/:id/services", k8sResourceController.CreateService)
		PUT("/clusters/:id/services", k8sResourceController.UpdateService)
		DELETE("/clusters/:id/services", k8sResourceController.DeleteService)

		// Ingresses
		GET("/clusters/:id/ingresses", k8sResourceController.ListIngress)
		GET("/clusters/:id/ingresses/:namespace/:name", k8sResourceController.GetIngress)
		POST("/clusters/:id/ingresses", k8sResourceController.CreateIngress)
		PUT("/clusters/:id/ingresses", k8sResourceController.UpdateIngress)
		DELETE("/clusters/:id/ingresses", k8sResourceController.DeleteIngress)

		// Pods
		GET("/clusters/:id/pods", k8sResourceController.ListPods)
		GET("/clusters/:id/pods/:namespace/:name", k8sResourceController.GetPod)
		PUT("/clusters/:id/pods", k8sResourceController.UpdatePod)
		GET("/clusters/:id/pods/:namespace/:name/logs", k8sResourceController.GetPodLogs)
		DELETE("/clusters/:id/pods", k8sResourceController.DeletePod)

		// ConfigMaps
		GET("/clusters/:id/configmaps", k8sResourceController.ListConfigMaps)
		GET("/clusters/:id/configmaps/:namespace/:name", k8sResourceController.GetConfigMap)
		POST("/clusters/:id/configmaps", k8sResourceController.CreateConfigMap)
		PUT("/clusters/:id/configmaps", k8sResourceController.UpdateConfigMap)
		DELETE("/clusters/:id/configmaps", k8sResourceController.DeleteConfigMap)

		// Secrets
		GET("/clusters/:id/secrets", k8sResourceController.ListSecrets)
		GET("/clusters/:id/secrets/:namespace/:name", k8sResourceController.GetSecret)
		POST("/clusters/:id/secrets", k8sResourceController.CreateSecret)
		PUT("/clusters/:id/secrets", k8sResourceController.UpdateSecret)
		DELETE("/clusters/:id/secrets", k8sResourceController.DeleteSecret)

		// Events
		GET("/clusters/:id/events", k8sResourceController.ListEvents)

		// K8s 终端管理
		GET("/terminal/active", k8sTerminalHandler.GetActiveSessions)
		POST("/terminal/sessions/:sessionId/terminate", k8sTerminalHandler.TerminateSession)

		// K8s 诊断功能
		GET("/diagnostic/commands", diagnosticController.GetDiagnosticCommands)
		GET("/diagnostic/pods/:clusterId/:namespace", diagnosticController.GetJavaPods)
		GET("/diagnostic/namespaces/:clusterId", diagnosticController.GetNamespaces)
		POST("/diagnostic/execute", diagnosticController.ExecuteDiagnostic)
		GET("/diagnostic/history", diagnosticController.GetDiagnosticHistory)

	}
}
