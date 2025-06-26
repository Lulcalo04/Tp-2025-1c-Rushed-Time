package cpu_internal

import (
	"fmt"
	"globals"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
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
	MotivoDesalojo    string
}

var EstructuraMemoriaDeCPU EstructuraMemoria
var ProcesoEjecutando PCBdeCPU
var SeUsaMemoria bool = false
var argumentoInstrucciones []string

var mutexProcesoEjecutando sync.Mutex

func IniciarCPU() {

	//Verifica el identificador de la cpu valido
	CpuId := VerificarIdentificadorCPU()

	//Inicializa la config de cpu
	globals.IniciarConfiguracion("cpu/config.json", &Config_CPU)

	//Declaro algoritmo de reemplazo de TLB y cantidad de entradas
	InicializarTLB()

	//Declaro algoritmo de reemplazo de Cache y cantidad de entradas
	InicializarCache()

	//Crea el archivo donde se logea cpu con su id
	Logger = ConfigurarLoggerCPU(CpuId, Config_CPU.LogLevel)

	go IniciarServerCPU(Config_CPU.PortCPU)

	//Realizar el handshake con Memoria
	if !HandshakeConMemoria(CpuId) {
		Logger.Debug("Error, no se pudo realizar el handshake con el Memoria")
		return
	}

	//Realiza el handshake con el kernel
	if !HandshakeConKernel(CpuId) {
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
	LogFetchInstruccion(ProcesoEjecutando.PID, ProcesoEjecutando.PC)
}

func Decode() {

	// Devuelve en un slice de strings las palabras de la instruccion actual separadas por espacios
	argumentoInstrucciones = strings.Fields(ProcesoEjecutando.InstruccionActual)

	if (argumentoInstrucciones[0] == "WRITE") || (argumentoInstrucciones[0] == "READ") || (argumentoInstrucciones[0] == "GOTO") {
		// Si la instruccion es WRITE READ O GOTO, Se tiene que utilizar la MMU para traducir la direccion logica a fisica
		SeUsaMemoria = true
	}

}

/* func Execute() {

	LogInstruccionEjecutada(ProcesoEjecutando.PID, argumentoInstrucciones[0], argumentoInstrucciones[1]+" "+argumentoInstrucciones[1]) //! PREGUNTAR QUE ES PARAMETROS
	//Si decode me dijo q tengo q traducir llamo a MMU

	if HayQueTraducir {
		DireccionFisica := ObtenerDireccionFisica(argumentoInstrucciones[1])

		switch argumentoInstrucciones[0] {
		case "WRITE":
			// Si la instruccion es WRITE, se escribe en la direccion fisica
			PeticionWriteAMemoria(DireccionFisica, argumentoInstrucciones[0], argumentoInstrucciones[2], ProcesoEjecutando.PID)
		case "READ":
			// Si la instruccion es READ, se lee de la direccion fisica
			PeticionReadAMemoria(DireccionFisica, argumentoInstrucciones[0], argumentoInstrucciones[2], ProcesoEjecutando.PID)
		case "GOTO":
			// Si la instruccion es GOTO, se cambia el PC al valor de la direccion logica
			PeticionGotoAMemoria(DireccionFisica, argumentoInstrucciones[0], ProcesoEjecutando.PID)
			return // No se incrementa el PC
		}
	}

	//Si estamos aca significa que la instuccion es syscall
	switch argumentoInstrucciones[0] {
	case "IO":
		// Si la instruccion es IO, se realiza una peticion al kernel para que maneje la syscall
		PeticionIOKernel(ProcesoEjecutando.PID, argumentoInstrucciones[1], argumentoInstrucciones[2])
	case "INIT_PROC":
		// Si la instruccion es INIT_PROC, se realiza una peticion al kernel para que maneje la syscall
		PeticionInitProcKernel(ProcesoEjecutando.PID, argumentoInstrucciones[1], argumentoInstrucciones[2])
	case "DUMP_MEMORY":
		// Si la instruccion es DUMP_MEMORY, se realiza una peticion al kernel para que maneje la syscall
		PeticionDumpMemoryKernel(ProcesoEjecutando.PID)
	case "EXIT":
		// Si la instruccion es EXIT, se realiza una peticion al kernel para que maneje la syscall
		PeticionExitKernel(ProcesoEjecutando.PID)
	}

	ProcesoEjecutando.PC++ // Incrementa el PC para la siguiente instruccion
} */

func Execute() {

	LogInstruccionEjecutada(ProcesoEjecutando.PID, argumentoInstrucciones[0], argumentoInstrucciones[1]+" "+argumentoInstrucciones[1]) //! PREGUNTAR QUE ES PARAMETROS

	if SeUsaMemoria {

		//Averiguar la pagina de esa Direccion Logica para saber si la tiene cache
		numeroDePagina, direccionLogicaInt := CalculoPagina(argumentoInstrucciones[1])
		var desplazamiento int = direccionLogicaInt % EstructuraMemoriaDeCPU.TamanioPagina
		var paginaCache *EntradaCache
		if CacheHabilitada {
			paginaCache = BuscarPaginaEnCache(numeroDePagina)
			if paginaCache == nil {
				//Cache Miss, tengo que pedir a memoria (TLB o MMU)
			}
			switch argumentoInstrucciones[0] {
			case "WRITE":
				// Si la instruccion es WRITE, se escribe en el byte correspondiente
				EscribirEnPagina(paginaCache, desplazamiento, argumentoInstrucciones[2])
			case "READ":
				// Si la instruccion es READ, se lee del byte correspondiente
				LeerDePagina(paginaCache, desplazamiento, argumentoInstrucciones[2])
			}
		} else { //Cache desabilitada, peticiones a memoria con los recursos directamente

		}

	}

	switch argumentoInstrucciones[0] {
	case "GOTO":
		//Actualiza el PC al valor de la instruccion
		ActualizarPC(argumentoInstrucciones[1])
		return // No se incrementa el PC
	case "IO":
		// Si la instruccion es IO, se realiza una peticion al kernel para que maneje la syscall
		PeticionIOKernel(ProcesoEjecutando.PID, argumentoInstrucciones[1], argumentoInstrucciones[2])
	case "INIT_PROC":
		// Si la instruccion es INIT_PROC, se realiza una peticion al kernel para que maneje la syscall
		PeticionInitProcKernel(ProcesoEjecutando.PID, argumentoInstrucciones[1], argumentoInstrucciones[2])
	case "DUMP_MEMORY":
		// Si la instruccion es DUMP_MEMORY, se realiza una peticion al kernel para que maneje la syscall
		PeticionDumpMemoryKernel(ProcesoEjecutando.PID)
	case "EXIT":
		// Si la instruccion es EXIT, se realiza una peticion al kernel para que maneje la syscall
		PeticionExitKernel(ProcesoEjecutando.PID)
	}

	mutexProcesoEjecutando.Lock()
	ProcesoEjecutando.PC++ // Incrementa el PC para la siguiente instruccion
	mutexProcesoEjecutando.Unlock()
}

func CheckInterrupt() bool {

	//Si hay interrupcion por atender..
	if ProcesoEjecutando.Interrupt {
		//Mutex para evitar condiciones de carrera
		//Marcamos que la interrupcion fue atendida
		mutexProcesoEjecutando.Lock()
		ProcesoEjecutando.Interrupt = false
		mutexProcesoEjecutando.Unlock()

		//Le avisamos a kernel que desalojamos
		PeticionDesalojoKernel()

		//Cortamos bucle de ciclo de instruccion
		return true
	}

	//Si no hay interrupcion, seguimos el ciclo de instruccion
	return false
}

func ObtenerDireccionFisica(direccionLogica string) int {

	nroPagina, direccionLogicaInt := CalculoPagina(direccionLogica)
	var frame int
	var direccionFisica int
	var desplazamiento int = direccionLogicaInt % EstructuraMemoriaDeCPU.TamanioPagina

	//if CacheHabilitada
	//ver si no esta cargada en cache ?

	//if TLBHabilitada
	//primero llame TLB
	if TLBHabilitada {
		frame = BuscarFrameEnTLB(nroPagina)
	}
	//tlb me pasa el inicio del frame correspondiente y a eso le tengo que sumar el desplazamiento
	//if TLBmiss, frame = -1 y llamo a MMU
	if frame == -1 {
		direccionFisica = MMU(direccionLogicaInt, nroPagina, desplazamiento)
	} else { // Si el frame fue encontrado en la TLB, calculo la direccion fisica directamente
		direccionFisica = frame*EstructuraMemoriaDeCPU.TamanioPagina + desplazamiento
	}

	return direccionFisica
}

func MMU(direccionLogicaInt int, nroPagina int, desplazamiento int) int {

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

	frame := PeticionFrameAMemoria(entradasPorNivel, ProcesoEjecutando.PID)
	LogObtenerMarco(ProcesoEjecutando.PID, nroPagina, frame)

	if TLBHabilitada {
		AgregarEntradaTLB(nroPagina, frame)
		Logger.Debug("Agregando entrada a TLB", "pagina", nroPagina, "marco", frame)
	}

	direccionFisica := frame*EstructuraMemoriaDeCPU.TamanioPagina + desplazamiento

	return direccionFisica
}

func CalculoPagina(direccionLogica string) (nroPag int, dlInt int) {

	direccionLogicaInt, err := strconv.Atoi(direccionLogica)
	if err != nil {
		Logger.Error("Error al convertir direccionLogica a int", "error", err)
		return -1, -1
	}

	nroPagina := direccionLogicaInt / EstructuraMemoriaDeCPU.TamanioPagina

	return nroPagina, direccionLogicaInt
}

func ActualizarPC(nuevoPC string) {

	nuevoPCInt, err := strconv.Atoi(nuevoPC)
	if err != nil {
		Logger.Error("Error al convertir nuevoPC a int", "error", err)
		return
	}
	mutexProcesoEjecutando.Lock()
	ProcesoEjecutando.PC = nuevoPCInt
	mutexProcesoEjecutando.Unlock()

}
