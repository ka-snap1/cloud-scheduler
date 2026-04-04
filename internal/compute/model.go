package compute

// 用户层的计算空间模型，包含计算空间规格和状态等信息
type ComputeSpaceSpec struct {
	// 核心算力需求
	CPU      int    `json:"cpu"`
	GPU      int    `json:"gpu"`
	GPUModel string `json:"gpu_model,omitempty"`
	MemoryGB int    `json:"memory_gb"`
	// 环境
	OSimage string `json:"os_image,omitempty"` //
	// 生命周期
	TTLSecond int `json:"tt"` // 单位hours，如果没有则不予创建

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
}
