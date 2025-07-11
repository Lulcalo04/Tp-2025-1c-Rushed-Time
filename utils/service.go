package globals

// &-------------------------------------------Structs de handlers (Request y Response)-------------------------------------------

type HandshakeRequest struct {
	Modulo    string `json:"modulo"`
	Ip        string `json:"ip"`
	Port      int    `json:"port"`
	Respuesta bool   `json:"respuesta"`
}

type HandshakeResponse struct {
	Modulo    string `json:"modulo"`
	Ip        string `json:"ip"`
	Port      int    `json:"port"`
	Respuesta bool   `json:"respuesta"`
	Mensaje   string `json:"mensaje"`
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

type InstruccionesResponse struct {
	PID           int      `json:"pid"`
	Instrucciones []string `json:"instrucciones"`
}

type KerneltoCPUDesalojoRequest struct {
	PID    int    `json:"pid"`
	Motivo string `json:"motivo"`
}

type KerneltoCPUDesalojoResponse struct {
	Respuesta bool `json:"respuesta"`
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

type IoHandshakeResponse struct {
	Respuesta bool   `json:"respuesta"`
	Mensaje   string `json:"mensaje"`
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
	PID  int    `json:"pid"`
	PC   int    `json:"pc"`
	Path string `json:"path"` // ! pedir a los chicos que nos manden esto en el request

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

type IOtoKernelDesconexionRequest struct {
	NombreDispositivo string `json:"nombre_dispositivo"`
	IpInstancia       string `json:"ip_instancia"`
	PuertoInstancia   int    `json:"puerto_instancia"`
}

type CPUtoKernelDesalojoRequest struct {
	PID    int    `json:"pid"`
	PC     int    `json:"pc"`
	Motivo string `json:"motivo"`
	CPUID  string `json:"cpu_id"` // Identificador de la CPU que solicita el desalojo
}

type CPUtoKernelDesalojoResponse struct {
	Respuesta bool `json:"response"`
}

type CPUtoMemoriaPageRequest struct {
	PID             int `json:"pid"`
	DireccionFisica int `json:"direccion_fisica"`
}

type MemoriaToCPUPageResponse struct {
	PID             int    `json:"pid"`
	ContenidoPagina []byte `json:"contenido_pagina"` // Contenido de la página solicitada
}

type CPUWriteAMemoriaRequest struct {
	PID             int    `json:"pid"`
	DireccionFisica int    `json:"direccion_fisica"`
	Data            []byte `json:"data"`
}

type CPUWriteAMemoriaResponse struct {
	Respuesta bool `json:"respuesta"`
}

type CPUReadAMemoriaRequest struct {
	PID             int `json:"pid"`
	DireccionFisica int `json:"direccion_fisica"`
	Tamanio         int `json:"data"`
}

type CPUReadAMemoriaResponse struct {
	Respuesta bool   `json:"respuesta"`
	Data      []byte `json:"data"` // Datos leídos de la memoria

}

type CPUActualizarPaginaEnMemoriaRequest struct {
	PID            int    `json:"pid"`
	NumeroDePagina int    `json:"numero_de_pagina"` // Número de página a actualizar
	Data           []byte `json:"data"`             // Datos a sobreescribir en la pagina
}

type CPUActualizarPaginaEnMemoriaResponse struct {
	Respuesta bool `json:"respuesta"`
	Frame     int  `json:"frame"` // Frame donde se actualizó la página
}
type SwappingRequest struct { // Request para el swapping
	PID int `json:"pid"`
}
type SwappingResponse struct { // Response para el swapping
	Respuesta bool `json:"respuesta"`
}
