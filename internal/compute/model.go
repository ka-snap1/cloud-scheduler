package compute

// 用户层的计算空间模型，包含计算空间规格和状态等信息
type ComputeSpaceSpec struct {
	// 核心算力需求
	CPU      int    `json:"cpu"`
	GPU      int    `json:"gpu"`
	GPUModel string `json:"gpu_model,omitempty"`
	MemoryGB int    `json:"memory_gb"`
	// 环境
	OSImage string `json:"os_image,omitempty"` //
	// 生命周期
	TTL int `json:"tt"` // 单位hours，如果没有则不予创建
	// 实例来源
	Provider string `json:"instance_source,omitempty"` // 例如：阿里云、腾讯云、华为云等
	// 资源配置来源模板
	ProfileRef string `json:"profile_ref,omitempty"`

	// todo
	// enablePublicIP bool `json:"enable_public_ip"` // 是否需要公网IP
	// bootstrap string `json:"bootstrap"` // 启动脚本&配置，用户可以提供一个脚本或配置文件，在计算空间创建后自动执行，以进行环境初始化、软件安装等操作。
	// region string `json:"region"` // 计算空间所在的地理区域，用户可以指定计算空间部署在哪个区域，以满足数据隐私、延迟等需求。
}

type ComputeSpace struct {
	// 计算空间ID
	Id string `json:"id"`
	// 计算空间规格
	Spec ComputeSpaceSpec `json:"spec"`
	// 计算空间状态
	Status ComputeSpaceStatus `json:"status"`
}

// todo 设计计算空间状态模型，包含计算空间的生命周期状态、运行状态等信息
type ComputeSpaceStatus struct {
	Phase    string `json:"phase"`              // 例如：Pending、Running、Stopped、Terminated等
	Provider string `json:"provider,omitempty"` // 例如：阿里云、腾讯云、华为云等
	// 云厂商侧实例ID
	ProviderID string `json:"provider_id,omitempty"`

	PrivateIP string `json:"private_ip,omitempty"` // 计算空间的内网IP地址，如果有的话
	PublicIP  string `json:"public_ip,omitempty"`  // 计算空间的公网IP地址，如果有的话

	InstanceId string `json:"instance_id,omitempty"` // 实例ID，例如阿里云ECS的InstanceId、腾讯云CVM的InstanceId等
	SSHuser    string `json:"ssh_user,omitempty"`    // ssh用户名
}
