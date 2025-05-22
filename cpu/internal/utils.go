package cpu_internal

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
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

type EstructuraMemoria struct {
	TamanioMemoria   int
	TamanioPagina    int
	EntradasPorTabla int
	NivelesDeTabla   int
}

type PCBdeCPU struct {
	PID               int
	PC                int
	InstruccionActual string
	Interrupt         bool
}

var EstructuraMemoriaDeCPU EstructuraMemoria
var ProcesoEjecutando PCBdeCPU
var InterrupcionAtendida bool = false
var HayQueTraducir bool = false
var argumentoInstrucciones []string

func IniciarCPU() {

	//Verifica el identificador de la cpu valido
	CpuId := VerificarIdentificadorCPU()

	//Inicializa la config de cpu
	globals.IniciarConfiguracion("cpu/config.json", &Config_CPU)

	//Crea el archivo donde se logea cpu con su id
	Logger = ConfigurarLoggerCPU(CpuId, Config_CPU.LogLevel)

	//Realizar el handshake con Memoria
	if !HandshakeConMemoria(CpuId) {
		Logger.Debug("Error, no se pudo realizar el handshake con el Memoria")
		return
	}

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

func CicloDeInstruccion() {

	for {
		Fetch()
		Decode()
		Execute()
		if CheckInterrupt() {
			break
		}
	}

}

func Fetch() {
	SolicitarSiguienteInstruccionMemoria(ProcesoEjecutando.PID, ProcesoEjecutando.PC)
}

func Decode() {

	// Devuelve en un slice de strings las palabras de la instruccion actual separadas por espacios
	argumentoInstrucciones = strings.Fields(ProcesoEjecutando.InstruccionActual)

	if (argumentoInstrucciones[0] == "WRITE") || (argumentoInstrucciones[0] == "READ") || (argumentoInstrucciones[0] == "GOTO") {
		// Si la instruccion es WRITE READ O GOTO, Se tiene que utilizar la MMU para traducir la direccion logica a fisica
		HayQueTraducir = true
	}

}

func Execute() {

	//Si decode me dijo q tengo q traducir llamo a MMU

	if HayQueTraducir {
		//DireccionFisica :=MMU(argumentoInstrucciones[1])

	}

}

func CheckInterrupt() bool {

	if ProcesoEjecutando.Interrupt {
		//Cortar bucle de ciclo de instruccion
		InterrupcionAtendida = true
		ProcesoEjecutando.Interrupt = false
		return true
	}

	return false
}
func MMU(direccionLogica string) int {

	direccionLogicaInt, err := strconv.Atoi(direccionLogica)
	if err != nil {
		Logger.Error("Error al convertir direccionLogica a int", "error", err)
		return -1
	}

	nroPagina := direccionLogicaInt / EstructuraMemoriaDeCPU.TamanioPagina

	N := EstructuraMemoriaDeCPU.NivelesDeTabla
	cantEntradas := EstructuraMemoriaDeCPU.EntradasPorTabla

	entradasPorNivel := make([]int, N)

	for x := 1; x <= N; x++ {
		exp := N - x
		divisor := 1
		for i := 0; i < exp; i++ {
			divisor *= cantEntradas
		}
		entradaNivelX := (nroPagina / divisor) % cantEntradas
		entradasPorNivel[x-1] = entradaNivelX
	}

	desplazamiento := direccionLogicaInt % EstructuraMemoriaDeCPU.TamanioPagina

	frame := PeticionFrameAMemoria(entradasPorNivel, ProcesoEjecutando.PID)

	direccionFisica := frame*EstructuraMemoriaDeCPU.TamanioPagina + desplazamiento

	return direccionFisica
}
