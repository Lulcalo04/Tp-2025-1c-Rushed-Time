package io_internal

type ConfigIO struct {
	IPKernel   string `json:"ip_kernel"`
	PortKernel string `json:"port_kernel"`
	PortIO     string `json:"port_io"`
	LogLevel   string `json:"log_level"`
}

var Config_IO *ConfigIO
