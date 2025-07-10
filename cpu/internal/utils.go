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

var CPUId string
var EstructuraMemoriaDeCPU EstructuraMemoria
var ProcesoEjecutando PCBdeCPU
var InstruccionUsaMemoria bool = false
var argumentoInstrucciones []string

var mutexProcesoEjecutando sync.Mutex

func IniciarCPU() {

	//Verifica el identificador de la cpu valido
	CPUId = VerificarIdentificadorCPU()
	fmt.Println("CPU inicializada con ID", CPUId)

	//Inicializa la config de cpu
	fmt.Println("Iniciando configuración de CPU...")
	globals.IniciarConfiguracion("cpu/config.json", &Config_CPU)

	//Declaro algoritmo de reemplazo de TLB y cantidad de entradas

	//InicializarTLB()

	//Declaro algoritmo de reemplazo de Cache y cantidad de entradas
	//InicializarCache()

	//Crea el archivo donde se logea cpu con su id
	fmt.Println("Iniciando logger de CPU...")
	Logger = ConfigurarLoggerCPU(CPUId, Config_CPU.LogLevel)

	fmt.Println("Iniciando servidor de CPU...")
	go IniciarServerCPU(Config_CPU.PortCPU)

	//Realizar el handshake con Memoria
	fmt.Println("Realizando handshake con Memoria...")
	if !HandshakeConMemoria(CPUId) {
		Logger.Debug("Error, no se pudo realizar el handshake con el Memoria")
		return
	}
	fmt.Println("Handshake con Memoria realizado correctamente")

	//Realiza el handshake con el kernel
	fmt.Println("Realizando handshake con Kernel...")
	if !HandshakeConKernel(CPUId) {
		Logger.Debug("Error, no se pudo realizar el handshake con el kernel")
		return
	}
	fmt.Println("Handshake con Kernel realizado correctamente")

	select {}
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
	InstruccionUsaMemoria = false // Reseteamos la variable que indica si la instruccion usa memoria
	SolicitarSiguienteInstruccionMemoria(ProcesoEjecutando.PID, ProcesoEjecutando.PC)
	LogFetchInstruccion(ProcesoEjecutando.PID, ProcesoEjecutando.PC)
}

func Decode() {

	// Devuelve en un slice de strings las palabras de la instruccion actual separadas por espacios
	argumentoInstrucciones = strings.Fields(ProcesoEjecutando.InstruccionActual)
	fmt.Println(argumentoInstrucciones)

	if (argumentoInstrucciones[0] == "WRITE") || (argumentoInstrucciones[0] == "READ") || (argumentoInstrucciones[0] == "GOTO") {
		// Si la instruccion es WRITE READ O GOTO, Se tiene que utilizar la MMU para traducir la direccion logica a fisica
		InstruccionUsaMemoria = true
	}

}

func Execute() {

	switch len(argumentoInstrucciones) {
	case 1:
		LogInstruccionEjecutada(ProcesoEjecutando.PID, argumentoInstrucciones[0], "")
	case 2:
		LogInstruccionEjecutada(ProcesoEjecutando.PID, argumentoInstrucciones[0], argumentoInstrucciones[1])
	case 3:
		LogInstruccionEjecutada(ProcesoEjecutando.PID, argumentoInstrucciones[0], argumentoInstrucciones[1]+" "+argumentoInstrucciones[2])
	}

	if InstruccionUsaMemoria {
		fmt.Println("argumentoInstrucciones:", argumentoInstrucciones)
		//Averiguar la pagina de esa Direccion Logica para saber si la tiene cache
		numeroDePagina, direccionLogicaInt := CalculoPagina(argumentoInstrucciones[1])
		var desplazamiento int = direccionLogicaInt % EstructuraMemoriaDeCPU.TamanioPagina
		var paginaCache *EntradaCache

		if CacheHabilitada { // Si la Cache esta habilitada, se busca la pagina en cache
			paginaCache = BuscarPaginaEnCache(numeroDePagina)
			if paginaCache == nil { // Cache Miss, busco la pagina en memoria
				// Se obtiene la direccion fisica
				direccionFisica := ObtenerDireccionFisica(numeroDePagina, direccionLogicaInt, desplazamiento)
				// Le pido la pagina a memoria y la guardo en cache
				paginaCache = PedirPaginaAMemoria(ProcesoEjecutando.PID, direccionFisica, numeroDePagina)
			}

			// Escribo/Leo la pagina en cache
			switch argumentoInstrucciones[0] {
			case "WRITE":
				// Si la instruccion es WRITE, se escribe en el byte correspondiente
				EscribirEnPaginaCache(paginaCache, desplazamiento, argumentoInstrucciones[2])
			case "READ":
				// Si la instruccion es READ, se lee del byte correspondiente
				LeerDePaginaCache(paginaCache, desplazamiento, argumentoInstrucciones[2])
			}

		} else { //Cache desabilitada, peticiones a memoria con los recursos directamente

			switch argumentoInstrucciones[0] {
			case "WRITE":
				// Si la instruccion es WRITE, se escribe en el byte correspondiente
				EscribirEnPaginaMemoria(ProcesoEjecutando.PID, direccionLogicaInt, argumentoInstrucciones[2])
			case "READ":
				// Si la instruccion es READ, se lee del byte correspondiente
				LeerDePaginaMemoria(ProcesoEjecutando.PID, direccionLogicaInt, argumentoInstrucciones[2])
			}

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

		if TLBHabilitada {
			// Liberamos las entradas de la TLB usadas por el proceso.
			LiberarEntradasTLB(ProcesoEjecutando.PID)
		}

		if CacheHabilitada {
			// Liberamos las entradas de la Chaché usadas por el proceso.
			LiberarEntradasCache(ProcesoEjecutando.PID)
		}

		//Le avisamos a kernel que desalojamos
		PeticionDesalojoKernel()

		//Cortamos bucle de ciclo de instruccion
		return true
	}

	//Si no hay interrupcion, seguimos el ciclo de instruccion
	return false
}

func ObtenerDireccionFisica(numeroDePagina int, direccionLogicaInt int, desplazamiento int) int {

	var frame int
	var direccionFisica int

	//Si la TLB esta habilitada
	if TLBHabilitada { // Busco el la traducción (dirección física del frame) en la TLB
		frame = BuscarFrameEnTLB(numeroDePagina)
	}

	if frame == -1 { //TLB Miss: Si el frame no fue encontrado en la TLB, llamo a MMU para realizar la traducción
		direccionFisica = MMU(direccionLogicaInt, numeroDePagina, desplazamiento)
	} else { //TLB Hit: Si el frame fue encontrado en la TLB, calculo la direccion fisica directamente
		//TLB me pasa el inicio del frame correspondiente y a eso le tengo que sumar el desplazamiento
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
