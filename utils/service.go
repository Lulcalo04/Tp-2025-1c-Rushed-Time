package globals

// &-------------------------------------------Structs de handlers (Request y Response)-------------------------------------------

type HandshakeRequest struct {
	Modulo string `json:"modulo"`
	Nombre string `json:"nombre"`
}

type PingResponse struct {
	Modulo  string `json:"modulo"`
	Mensaje string `json:"mensaje"`
}

type PeticionMemoriaRequest struct {
	Modulo     string `json:"modulo"`
	ProcesoPCB PCB    `json:"proceso_pcb"`
}

type PeticionMemoriaResponse struct {
	Modulo    string `json:"modulo"`
	Respuesta bool   `json:"respuesta"`
	Mensaje   string `json:"mensaje"`
}

type LiberacionMemoriaRequest struct {
	Modulo string `json:"modulo"`
	PID    int    `json:"pid"`
}

type LiberacionMemoriaResponse struct {
	Modulo    string `json:"modulo"`
	Respuesta bool   `json:"respuesta"`
	Mensaje   string `json:"mensaje"`
}

type DumpMemoryRequest struct {
	Modulo string `json:"modulo"`
	PID    int    `json:"pid"`
}

type DumpMemoryResponse struct {
	Modulo    string `json:"modulo"`
	Respuesta bool   `json:"respuesta"`
	Mensaje   string `json:"mensaje"`
}

type InstruccionesRequest struct {
	PID int `json:"pid"`
}

type InstruccionesResponse struct {
	PID           int      `json:"pid"`
	Instrucciones []string `json:"instrucciones"`
}

type DesalojoRequest struct {
	PID int `json:"pid"`
}

type DesalojoResponse struct {
	PID int `json:"pid"`
	PC  int `json:"pc"`
}

type InitProcSyscallRequest struct {
	PID           int    `json:"pid"`
	NombreArchivo string `json:"nombreArchivo"`
	Tamanio       int    `json:"tamanio"`
}

type InitProcSyscallResponse struct {
	Respuesta bool `json:"respuesta"`
}

type ExitSyscallRequest struct {
	PID int `json:"pid"`
}

type ExitSyscallResponse struct {
	Respuesta bool `json:"respuesta"`
}

type DumpMemorySyscallRequest struct {
	PID int `json:"pid"`
}

type DumpMemorySyscallResponse struct {
	Respuesta bool `json:"respuesta"`
}

type IoSyscallRequest struct {
	PID               int    `json:"pid"`
	NombreDispositivo string `json:"nombreDispositivo"`
	Tiempo            int    `json:"tiempo"`
}

type IoSyscallResponse struct {
	Respuesta bool `json:"respuesta"`
}

type IoHandshakeRequest struct {
	IPio   string `json:"ip_io"`
	PortIO int    `json:"port_io"`
	Nombre string `json:"nombre"`
}

type CPUToKernelHandshakeRequest struct {
	CPUID  string `json:"cpu_id"`
	Puerto int    `json:"puerto"`
	Ip     string `json:"ip"`
}

type CPUToKernelHandshakeResponse struct {
	Modulo    string `json:"modulo"`
	Respuesta bool   `json:"respuesta"`
	Mensaje   string `json:"mensaje"`
}

type ProcesoAEjecutarRequest struct {
	PID int `json:"pid"`
	PC  int `json:"pc"`
}

type ProcesoAEjecutarResponse struct {
	PID    int    `json:"pid"`
	PC     int    `json:"pc"`
	Motivo string `json:"motivo"`
}

type IORequest struct {
	NombreDispositivo string `json:"nombre_dispositivo"`
	PID               int    `json:"pid"`
	Tiempo            int    `json:"tiempo"`
}

type IOResponse struct {
	NombreDispositivo string `json:"nombre_dispositivo"`
	PID               int    `json:"pid"`
	Respuesta         bool   `json:"respuesta"`
}

type InstruccionAMemoriaRequest struct {
	PID int `json:"pid"`
	PC  int `json:"pc"`
}

type InstruccionAMemoriaResponse struct {
	InstruccionAEjecutar string `json:"instruccion"`
}

type CPUToMemoriaHandshakeRequest struct {
	CPUID string `json:"cpu_id"`
}

type CPUToMemoriaHandshakeResponse struct {
	TamanioMemoria   int `json:"tamanio_memoria"`
	TamanioPagina    int `json:"tamanio_pagina"`
	EntradasPorTabla int `json:"entradas_por_tabla"`
	NivelesDeTabla   int `json:"niveles_de_tabla"`
}

type SolicitudFrameRequest struct {
	PID              int   `json:"pid"`
	EntradasPorNivel []int `json:"entradas_por_nivel"`
}

type SolicitudFrameResponse struct {
	Frame int `json:"frame"`
}

type CPUWriteAMemoriaRequest struct {
	PID             int    `json:"pid"`
	Instruccion     string `json:"instruccion"`// creo que no es necesario mandarlo a Memoria
	DireccionFisica int    `json:"direccion_fisica"`
	Data            string `json:"data"`
}

type CPUWriteAMemoriaResponse struct {
	Respuesta bool `json:"respuesta"`
}

type CPUReadAMemoriaRequest struct {
	PID             int    `json:"pid"`
	Instruccion     string `json:"instruccion"`
	DireccionFisica int    `json:"direccion_fisica"`
	Data            string `json:"data"`
}

type CPUReadAMemoriaResponse struct {
	Respuesta bool `json:"respuesta"`
	Data      int `json:"data"` // Datos leídos de la memoria

}

type CPUGotoAMemoriaRequest struct {
	PID             int    `json:"pid"`
	Instruccion     string `json:"instruccion"`
	DireccionFisica int    `json:"direccion_fisica"`
}

type CPUGotoAMemoriaResponse struct {
	Respuesta bool `json:"respuesta"`
}

type IOtoKernelDesconexionRequest struct {
	NombreDispositivo string `json:"nombre_dispositivo"`
	IpInstancia       string `json:"ip_instancia"`
	PuertoInstancia   int    `json:"puerto_instancia"`
}
