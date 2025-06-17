package globals

import (
	"encoding/json"
	"log/slog"
	"os"
)

// &-------------------------------------------Tipos de datos para el manejo de los estados de los procesos-------------------------------------------

type Estado string

const (
	New         Estado = "NEW"
	Ready       Estado = "READY"
	Exec        Estado = "EXEC"
	Blocked     Estado = "BLOCKED"
	SuspReady   Estado = "SUSP_READY"
	SuspBlocked Estado = "SUSP_BLOCKED"
	Exit        Estado = "EXIT"
)

// &-------------------------------------------Structs de PCB para los procesos-------------------------------------------
type PCB struct {
	PID               int            `json:"pid"`
	PC                int            `json:"pc"`
	Estado            Estado         `json:"estado"`
	MetricasDeEstados map[Estado]int `json:"metricas_de_estados"`
	MetricasDeTiempos map[Estado]int `json:"metricas_de_tiempos"` //en milisegundos, con una libreria especifica
	TamanioEnMemoria  int            `json:"tamanio_en_memoria"`  //Por ahora lo tomamos como entero, pero puede variar
}

// &-------------------------------------------Structs de handlers (Request y Response)-------------------------------------------

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
	Instruccion     string `json:"instruccion"`
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
	IpInstancia	 string `json:"ip_instancia"`
	PuertoInstancia int    `json:"puerto_instancia"`
}

// &-------------------------------------------Inicio de configuraciones-------------------------------------------

func IniciarConfiguracion(filePath string, config interface{}) {
	//! NO PUDIMOS CONFIGURAR EL LOG, NO SE PUEDE HACER NADA
	//! HASTA QUE NO SE CARGUE LA CONFIGURACION

	configFile, err := os.Open(filePath)
	if err != nil {
		slog.Debug(err.Error())
		os.Exit(1)
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(config)
	if err != nil {
		slog.Debug("Error al decodificar la configuración", "error", err.Error())
		os.Exit(1)
	}
	if config == nil {
		slog.Debug("No se pudo cargar la configuración")
		os.Exit(1)
	}

}

// &-------------------------------------------Inicio de funciones de logger-------------------------------------------

func ConfigurarLogger(nombreDelModulo string, logLevelModulo string) *slog.Logger {

	// Definimos la ruta del log
	rutaDelLog := nombreDelModulo + "/" + nombreDelModulo + ".log"
	// Definimos el nivel de log
	nivel := PasarStringALogLevel(logLevelModulo)

	logFile, err := os.OpenFile(rutaDelLog, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	handler := slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: nivel})

	// Guardamos la instancia del logger en una variable global o accesible por el módulo
	Logger := slog.New(handler)

	return Logger
}

func PasarStringALogLevel(nivel string) slog.Level {
	var logLevel slog.Level

	switch nivel {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	return logLevel
}
