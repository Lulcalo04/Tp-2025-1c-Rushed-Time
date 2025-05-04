package io_internal

import "log/slog"

type ConfigIO struct {
	IPKernel   string `json:"ip_kernel"`
	PortKernel int    `json:"port_kernel"`
	PortIO     int    `json:"port_io"`
	LogLevel   string `json:"log_level"`
}

var Config_IO *ConfigIO

var Logger *slog.Logger
