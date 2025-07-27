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
var ArgumentoInstrucciones []string

var mutexProcesoEjecutando sync.Mutex

var mutexCicloDeInstruccion sync.Mutex

// &-------------------------------------------Funciones de inicialización de CPU-------------------------------------------

func RecibirParametrosConfiguracion() string {

	if len(os.Args) < 3 {
		fmt.Println("Error, mal escrito usa: ./cpu.go [archivo_configuracion] [identificador]")
		os.Exit(1)
	}

	nombreArchivoConfiguracion := os.Args[1]
	CPUId = os.Args[2]

	return nombreArchivoConfiguracion
}

func IniciarCPU(nombreArchivoConfiguracion string) {
	fmt.Println("CPU inicializada con ID", CPUId)

	//Inicializa la config de cpu
	fmt.Println("Iniciando configuración de CPU...")
	globals.IniciarConfiguracion("utils/configs/"+nombreArchivoConfiguracion+".json", &Config_CPU)

	//Crea el archivo donde se logea cpu con su id
	fmt.Println("Iniciando logger de CPU...")
	Logger = ConfigurarLoggerCPU(CPUId, Config_CPU.LogLevel)

	//Declaro algoritmo de reemplazo de TLB y cantidad de entradas
	InicializarTLB()

	//Declaro algoritmo de reemplazo de Cache y cantidad de entradas
	InicializarCache()

	fmt.Println("Iniciando servidor de CPU, en el puerto:", Config_CPU.PortCPU)
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
}

//& -------------------------------------------Funciones del ciclo de instrucción-------------------------------------------

func CicloDeInstruccion() {

	for {
		//Me fijo q el motivo no sea plani, xq si es asi, salteamos hasta check interrupt
		if ProcesoEjecutando.MotivoDesalojo != "Planificador" {
			Fetch()
			Decode()
			Execute()
		}
		if CheckInterrupt() {
			Logger.Debug("Rompi el ciclo de instrucción por interrupción", "PID", ProcesoEjecutando.PID)
			fmt.Println("Rompi el ciclo de instrucción por interrupción para el proceso", ProcesoEjecutando.PID)
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
	ArgumentoInstrucciones = strings.Fields(ProcesoEjecutando.InstruccionActual)

	if (ArgumentoInstrucciones[0] == "WRITE") || (ArgumentoInstrucciones[0] == "READ") {
		// Si la instruccion es WRITE o READ, Se tiene que utilizar la MMU para traducir la direccion logica a fisica
		InstruccionUsaMemoria = true
	}

	mutexProcesoEjecutando.Lock()
	if ProcesoEjecutando.MotivoDesalojo == "Planificador" && (ArgumentoInstrucciones[0] == "IO" || ArgumentoInstrucciones[0] == "DUMP_MEMORY" || ArgumentoInstrucciones[0] == "EXIT") {
		ProcesoEjecutando.MotivoDesalojo = ArgumentoInstrucciones[0]
	}
	mutexProcesoEjecutando.Unlock()

}

func Execute() {

	switch len(ArgumentoInstrucciones) {
	case 1:
		LogInstruccionEjecutada(ProcesoEjecutando.PID, ArgumentoInstrucciones[0], "")
	case 2:
		LogInstruccionEjecutada(ProcesoEjecutando.PID, ArgumentoInstrucciones[0], ArgumentoInstrucciones[1])
	case 3:
		LogInstruccionEjecutada(ProcesoEjecutando.PID, ArgumentoInstrucciones[0], ArgumentoInstrucciones[1]+" "+ArgumentoInstrucciones[2])
	}

	if InstruccionUsaMemoria {
		//Averiguar la pagina de esa Direccion Logica para saber si la tiene cache
		numeroDePagina, direccionLogicaInt := CalculoPagina(ArgumentoInstrucciones[1])
		var desplazamiento int = direccionLogicaInt % EstructuraMemoriaDeCPU.TamanioPagina
		var paginaCache *EntradaCache
		direccionFisica := ObtenerDireccionFisica(numeroDePagina, direccionLogicaInt, desplazamiento)

		if CacheHabilitada { // Si la Cache esta habilitada, se busca la pagina en cache
			paginaCache = BuscarPaginaEnCache(numeroDePagina)
			if paginaCache == nil { // Cache Miss, busco la pagina en memoria
				// Se obtiene la direccion fisica
				// Le pido la pagina a memoria y la guardo en cache
				paginaCache = PedirPaginaAMemoria(ProcesoEjecutando.PID, direccionFisica, numeroDePagina)
			}

			// Escribo/Leo la pagina en cache
			switch ArgumentoInstrucciones[0] {
			case "WRITE":
				// Si la instruccion es WRITE, se escribe en el byte correspondiente
				EscribirEnPaginaCache(paginaCache, desplazamiento, ArgumentoInstrucciones[2])
			case "READ":
				// Si la instruccion es READ, se lee del byte correspondiente
				LeerDePaginaCache(paginaCache, desplazamiento, ArgumentoInstrucciones[2])
			}

		} else { //Cache desabilitada, peticiones a memoria con los recursos directamente

			switch ArgumentoInstrucciones[0] {
			case "WRITE":
				// Si la instruccion es WRITE, se escribe en el byte correspondiente
				EscribirEnPaginaMemoria(ProcesoEjecutando.PID, direccionFisica, ArgumentoInstrucciones[2])
			case "READ":
				// Si la instruccion es READ, se lee del byte correspondiente
				LeerDePaginaMemoria(ProcesoEjecutando.PID, direccionFisica, ArgumentoInstrucciones[2])
			}

		}

	}

	switch ArgumentoInstrucciones[0] {
	case "GOTO":
		//Actualiza el PC al valor de la instruccion
		ActualizarPC(ArgumentoInstrucciones[1])
		return // No se incrementa el PC
	case "IO":
		// Si la instruccion es IO, se realiza una peticion al kernel para que maneje la syscall
		PeticionIOKernel(ProcesoEjecutando.PID, ArgumentoInstrucciones[1], ArgumentoInstrucciones[2])
	case "INIT_PROC":
		// Si la instruccion es INIT_PROC, se realiza una peticion al kernel para que maneje la syscall
		PeticionInitProcKernel(ProcesoEjecutando.PID, ArgumentoInstrucciones[1], ArgumentoInstrucciones[2])
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

	mutexProcesoEjecutando.Lock()

	if ProcesoEjecutando.MotivoDesalojo == "Planificador" && (ArgumentoInstrucciones[0] == "IO" || ArgumentoInstrucciones[0] == "DUMP_MEMORY" || ArgumentoInstrucciones[0] == "EXIT") {
		ProcesoEjecutando.MotivoDesalojo = ArgumentoInstrucciones[0]
		Logger.Debug("Caso de planificador, motivo de desalojo actualizado", "motivo", ProcesoEjecutando.MotivoDesalojo)
	}

	//Si hay interrupcion por atender..
	if ProcesoEjecutando.Interrupt {

		if TLBHabilitada {
			// Liberamos las entradas de la TLB usadas por el proceso.
			LiberarEntradasTLB(ProcesoEjecutando.PID)
		}

		if CacheHabilitada {
			// Liberamos las entradas de la Chaché usadas por el proceso.
			LiberarEntradasCache(ProcesoEjecutando.PID)
		}

		//Le avisamos a kernel que desalojamos
		PeticionDesalojoKernel(ProcesoEjecutando.PID, ProcesoEjecutando.PC, ProcesoEjecutando.MotivoDesalojo)

		//Marcamos que la interrupcion fue atendida
		ProcesoEjecutando.Interrupt = false
		ProcesoEjecutando.MotivoDesalojo = ""

		//Cortamos bucle de ciclo de instruccion
		mutexProcesoEjecutando.Unlock()
		mutexCicloDeInstruccion.Unlock()
		return true
	}

	//Si no hay interrupcion, seguimos el ciclo de instruccion
	mutexProcesoEjecutando.Unlock()
	return false
}

func ObtenerDireccionFisica(numeroDePagina int, direccionLogicaInt int, desplazamiento int) int {

	frame := -1
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
