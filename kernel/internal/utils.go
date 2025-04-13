package kernel_internal

type ConfigKernel struct {
	IPMemory           string `json:"ip_memory"`
	PortMemory         string `json:"port_memory"`
	PortKernel         string `json:"port_kernel"`
	SchedulerAlgorithm string `json:"scheduler_algorithm"`
	NewAlgorithm       string `json:"new_algorithm"`
	Alpha              string `json:"alpha"`
	SuspensionTime     int    `json:"suspension_time"`
	LogLevel           string `json:"log_level"`
}

var Config_Kernel *ConfigKernel
