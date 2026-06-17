package dto

// ServerQueryParams 服务器查询参数
type ServerQueryParams struct {
	Page         int    `form:"page" binding:"min=1"`
	PageSize     int    `form:"pageSize" binding:"min=1,max=100"`
	Hostname     string `form:"hostname" binding:"omitempty,max=100"`
	IP           string `form:"ip" binding:"omitempty,ip"`
	InnerIP      string `form:"innerIp" binding:"omitempty,ip"`
	Env          string `form:"env" binding:"omitempty,oneof=dev test prod"`
	Status       string `form:"status" binding:"omitempty,oneof=active inactive maintenance"`
	Provider     string `form:"provider" binding:"omitempty,max=50"`
	GroupID      *uint  `form:"groupId" binding:"omitempty,min=1"`
	AgentStatus  string `form:"agentStatus" binding:"omitempty,oneof=running stopped unavailable"`
	BusinessUnitID *uint `form:"businessUnitId" binding:"omitempty,min=1"`
	Tags         string `form:"tags" binding:"omitempty"`
}

// K8sResourceQueryParams K8s资源查询参数（复用分页参数）
type K8sResourceQueryParams struct {
	Page       int    `form:"page" binding:"min=1"`
	PageSize   int    `form:"pageSize" binding:"min=1,max=100"`
	Namespace  string `form:"namespace" binding:"required,max=255"`
}

// K8sPodQueryParams K8s Pod查询参数（支持labelSelector）
type K8sPodQueryParams struct {
	Page          int    `form:"page" binding:"min=1"`
	PageSize      int    `form:"pageSize" binding:"min=1,max=100"`
	Namespace     string `form:"namespace" binding:"required,max=255"`
	LabelSelector string `form:"labelSelector" binding:"omitempty"`
}

// ServerCreateParams 创建服务器参数
type ServerCreateParams struct {
	Hostname      string `json:"hostname" binding:"required,min=1,max=100,hostname"`
	IP             string `json:"ip" binding:"required,ip"`
	InnerIP        string `json:"innerIp" binding:"omitempty,ip"`
	SSHPort        int    `json:"sshPort" binding:"required,min=1,max=65535"`
	Env            string `json:"env" binding:"required,oneof=dev test prod"`
	Status         string `json:"status" binding:"omitempty,oneof=active inactive maintenance"`
	Provider       string `json:"provider" binding:"omitempty,max=50"`
	OS             string `json:"os" binding:"omitempty,max=50"`
	Arch           string `json:"arch" binding:"omitempty,max=50"`
	CPU            int    `json:"cpu" binding:"omitempty,min=1"`
	Memory         int    `json:"memory" binding:"omitempty,min=1"`
	Disk           int    `json:"disk" binding:"omitempty,min=1"`
	SSHCredentialID uint   `json:"sshCredentialId" binding:"omitempty,min=1"`
	CredentialID   uint   `json:"credentialId" binding:"omitempty,min=1"`
	CredentialIDs  []uint `json:"credentialIds" binding:"omitempty,min=1,dive"`
	GroupIDs       []uint `json:"groupIds" binding:"omitempty,min=1,dive"`
	Remarks        string `json:"remarks" binding:"omitempty,max=500"`
}

// ServerUpdateParams 更新服务器参数
type ServerUpdateParams struct {
	Hostname      *string `json:"hostname" binding:"omitempty,min=1,max=100,hostname"`
	IP            *string `json:"ip" binding:"omitempty,ip"`
	InnerIP       *string `json:"innerIp" binding:"omitempty,ip"`
	SSHPort       *int    `json:"sshPort" binding:"omitempty,min=1,max=65535"`
	Env           *string `json:"env" binding:"omitempty,oneof=dev test prod"`
	Status        *string `json:"status" binding:"omitempty,oneof=active inactive maintenance"`
	Provider      *string `json:"provider" binding:"omitempty,max=50"`
	OS            *string `json:"os" binding:"omitempty,max=50"`
	Arch          *string `json:"arch" binding:"omitempty,max=50"`
	CPU           *int    `json:"cpu" binding:"omitempty,min=1"`
	Memory        *int    `json:"memory" binding:"omitempty,min=1"`
	Disk          *int    `json:"disk" binding:"omitempty,min=1"`
	Remarks       *string `json:"remarks" binding:"omitempty,max=500"`
}

// UserCreateParams 创建用户参数
type UserCreateParams struct {
	Username     string   `json:"username" binding:"required,min=3,max=50,alphanum"`
	Password     string   `json:"password" binding:"required,min=6,max=100"`
	Email        string   `json:"email" binding:"required,email"`
	RealName     string   `json:"realName" binding:"required,max=100"`
	Phone        string   `json:"phone" binding:"omitempty,max=20"`
	RoleIDs      []uint   `json:"roleIds" binding:"required,min=1,dive"`
	Status       int      `json:"status" binding:"omitempty,oneof=0 1"`
}

// UserUpdateParams 更新用户参数
type UserUpdateParams struct {
	Password  *string `json:"password" binding:"omitempty,min=6,max=100"`
	Email     *string `json:"email" binding:"omitempty,email"`
	RealName  *string `json:"realName" binding:"omitempty,max=100"`
	Phone     *string `json:"phone" binding:"omitempty,max=20"`
	RoleIDs   *[]uint `json:"roleIds" binding:"omitempty,min=1,dive"`
	Status    *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// MenuCreateParams 创建菜单参数
type MenuCreateParams struct {
	Name       string `json:"name" binding:"required,min=1,max=50"`
	Icon       string `json:"icon" binding:"omitempty,max=100"`
	Path       string `json:"path" binding:"required,max=200"`
	Permission string `json:"permission" binding:"omitempty,max=100"`
	MenuType   string `json:"menuType" binding:"required,oneof=menu directory"`
	ParentID   uint   `json:"parentId" binding:"omitempty,min=0"`
	Sort       int    `json:"sort" binding:"omitempty,min=0"`
	Status     int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// MenuUpdateParams 更新菜单参数
type MenuUpdateParams struct {
	Name       *string `json:"name" binding:"omitempty,min=1,max=50"`
	Icon       *string `json:"icon" binding:"omitempty,max=100"`
	Path       *string `json:"path" binding:"omitempty,max=200"`
	Permission *string `json:"permission" binding:"omitempty,max=100"`
	MenuType   *string `json:"menuType" binding:"omitempty,oneof=menu directory"`
	ParentID   *uint   `json:"parentId" binding:"omitempty,min=0"`
	Sort       *int    `json:"sort" binding:"omitempty,min=0"`
	Status     *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// RoleCreateParams 创建角色参数
type RoleCreateParams struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Code        string `json:"code" binding:"required,min=1,max=50,alphanumunderscore"`
	Description string `json:"description" binding:"omitempty,max=500"`
	MenuIDs     []uint `json:"menuIds" binding:"required,min=1,dive"`
	Status      int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// RoleUpdateParams 更新角色参数
type RoleUpdateParams struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=50"`
	Code        *string `json:"code" binding:"omitempty,min=1,max=50,alphanumunderscore"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	MenuIDs     *[]uint `json:"menuIds" binding:"omitempty,min=1,dive"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// LoginParams 登录参数
type LoginParams struct {
	Username string `json:"username" binding:"required,min=1,max=50"`
	Password string `json:"password" binding:"required,min=1,max=100"`
}

// ValidateCustom 自定义验证
func (s ServerQueryParams) ValidateCustom() error {
	// 可以添加自定义验证逻辑
	return nil
}
