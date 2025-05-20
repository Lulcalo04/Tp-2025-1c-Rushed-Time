package cpu_internal

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"utils/globals"
)

type ConfigCPU struct {
	PortCPU          int    `json:"port_cpu"`
	IpCpu            string `json:"ip_cpu"`
	IPMemory         string `json:"ip_memory"`
	PortMemory       int    `json:"port_memory"`
	IPKernel         string `json:"ip_kernel"`
	PortKernel       int    `json:"port_kernel"`
	TLBEntries       int    `json:"tlb_entries"`
	TLBReplacement   string `json:"tlb_replacement"`
	CacheEntries     int    `json:"cache_entries"`
	CacheReplacement string `json:"cache_replacement"`
	CacheDelay       int    `json:"cache_delay"`
	LogLevel         string `json:"log_level"`
}

var Config_CPU *ConfigCPU

var Logger *slog.Logger

type PCB_CPU struct {
	PID               int
	PC                int
	InstruccionActual string
}

var ProcesoEjecutando PCB_CPU

func IniciarCPU() {

	//Verifica el identificador de la cpu valido
	CpuId := VerificarIdentificadorCPU()

	//Inicializa la config de cpu
	globals.IniciarConfiguracion("cpu/config.json", &Config_CPU)

	//Crea el archivo donde se logea cpu con su id
	Logger = ConfigurarLoggerCPU(CpuId, Config_CPU.LogLevel)

	//Realiza el handshake con el kernel
	if HandshakeConKernel(CpuId) {
		CicloDeInstruccion()
	} else {
		Logger.Debug("Error, no se pudo realizar el handshake con el kernel")
		return
	}

}

func VerificarIdentificadorCPU() string {

	if len(os.Args) < 2 {
		fmt.Println("Error, mal escrito usa: ./cpu.go [identificador]")
		os.Exit(1)
	}
	CpuId := os.Args[1]

	return CpuId
}

func MMU(direccionLogica string) {

}

func CicloDeInstruccion() {

	for {
		Fetch()
		Decode()
		//Execute()
		//Check Interrupt()
	}

}

func Fetch() {
	SolicitarSiguienteInstruccionMemoria(ProcesoEjecutando.PID, ProcesoEjecutando.PC)
}

func Decode() {

	// Devuelve en un slice de strings las palabras de la instruccion actual separadas por espacios
	argumentoInstrucciones := strings.Fields(ProcesoEjecutando.InstruccionActual)

	if (argumentoInstrucciones[0] == "WRITE") || (argumentoInstrucciones[0] == "READ") || (argumentoInstrucciones[0] == "GOTO") {
		// Si la instruccion es WRITE READ O GOTO, Se tiene que utilizar la MMU para traducir la direccion logica a fisica
		MMU(argumentoInstrucciones[1])
	}

}
