package kernel_internal

import (
	"fmt"
	"globals"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"
)

// &-------------------------------------------Config de Kernel-------------------------------------------------------------
type ConfigKernel struct {
	IPMemory              string  `json:"ip_memory"`
	PortMemory            int     `json:"port_memory"`
	IPKernel              string  `json:"ip_kernel"`
	PortKernel            int     `json:"port_kernel"`
	SchedulerAlgorithm    string  `json:"scheduler_algorithm"`
	ReadyIngressAlgorithm string  `json:"ready_ingress_algorithm"`
	Alpha                 float64 `json:"alpha"`
	InitialEstimate       int     `json:"initial_estimate"`
	SuspensionTime        int     `json:"suspension_time"`
	LogLevel              string  `json:"log_level"`
}

type IdentificadorCPU struct {
	CPUID   string
	Puerto  int
	Ip      string
	Ocupado bool
	PID     int
}

var ListaIdentificadoresCPU []IdentificadorCPU = make([]IdentificadorCPU, 0)

var Config_Kernel *ConfigKernel

var Logger *slog.Logger

var ContadorPID int = -1

var canceladoresBlocked = make(map[int]chan struct{})

// &-------------------------------------------Funciones de Kernel-------------------------------------------------------------

func IniciarKernel() {
	//Inicializa la config de kernel
	fmt.Println("Iniciando configuración...")
	globals.IniciarConfiguracion("kernel/config.json", &Config_Kernel)

	//Crea el archivo donde se logea kernel
	fmt.Println("Iniciando logger...")
	Logger = globals.ConfigurarLogger("kernel", Config_Kernel.LogLevel)

	//Prende el server de kernel en un hilo aparte
	fmt.Println("Iniciando servidor de KERNEL, en el puerto:", Config_Kernel.PortKernel)
	go IniciarServerKernel(Config_Kernel.PortKernel)

	//Realiza el handshake con memoria

	HandshakeConMemoria(Config_Kernel.IPMemory, Config_Kernel.PortMemory)

	//Realizar el handshake con Memoria
	/* if !HandshakeCon(CpuId) {
		Logger.Debug("Error, no se pudo realizar el handshake con el Memoria")
		return
	} */

	//Inicia los planificadores
	IniciarPlanificadores()
	fmt.Println("Kernel iniciado correctamente.")
}

func InicializarProcesoCero() {
	if len(os.Args) < 3 {
		Logger.Debug("Error, mal escrito usa: .kernel/kernel.go [archivo_pseudocodigo] [tamanio_proceso]")
		os.Exit(1)
	}

	// Leer el nombre del archivo de pseudocódigo
	nombreArchivoPseudocodigo := os.Args[1]

	// Leer y convertir el tamaño del proceso a entero
	tamanioProceso, err := strconv.Atoi(os.Args[2])
	if err != nil {
		Logger.Debug("Error: el tamaño del proceso debe ser un número entero.", "valor_recibido", os.Args[2])
		os.Exit(1)
	}

	Logger.Debug("Inicializando Proceso Cero",
		"tamanio_proceso", tamanioProceso,
		"nombre_archivo_pseudocodigo", nombreArchivoPseudocodigo)

	InicializarPCB(tamanioProceso, nombreArchivoPseudocodigo)

}

func BuscarProcesoEnCola(pid int, cola *[]globals.PCB) *globals.PCB {
	for i, proceso := range *cola {
		if proceso.PID == pid {
			return &(*cola)[i]
		}
	}
	return nil
}

func InicializarPCB(tamanioEnMemoria int, nombreArchivoPseudo string) {
	mensajeInicializandoPCB := "Inicializando PCB" + " tamanio_en_memoria=" + strconv.Itoa(tamanioEnMemoria) + ", nombre_archivo_pseudo=" + nombreArchivoPseudo
	Logger.Debug(mensajeInicializandoPCB)
	fmt.Println(mensajeInicializandoPCB)

	ContadorPID++

	pcb := globals.PCB{
		PID:                ContadorPID,
		PC:                 0,
		Estado:             globals.New,
		PathArchivoPseudo:  nombreArchivoPseudo,
		InicioEstadoActual: time.Now(),
		MetricasDeEstados:  make(map[globals.Estado]int),
		MetricasDeTiempos:  make(map[globals.Estado]time.Duration),
		TamanioEnMemoria:   tamanioEnMemoria,
		EstimacionDeRafaga: globals.EstructuraRafaga{
			TiempoDeRafaga: float64(Config_Kernel.InitialEstimate),
			YaCalculado:    true,
		},
		TiempoDeUltimaRafaga: 0,
	}

	// Bloquear el mutex de la cola de New
	ColaMutexes[&ColaNew].Lock()

	// Agregar el proceso a la cola de New
	pcb.MetricasDeEstados[globals.New] = 1
	pcb.MetricasDeTiempos[globals.New] = time.Since(pcb.InicioEstadoActual)
	ColaNew = append(ColaNew, pcb)

	// Desbloquear el mutex de la cola de New
	ColaMutexes[&ColaNew].Unlock()

	//fmt.Println("PID", ContadorPID, " creado. Procesos en Cola NEW despues de mover el proceso:", len(ColaNew))
	LogCreacionDeProceso(ContadorPID)

	// Al agregar un nuevo proceso a la cola de New, notificamos al planificador de largo plazo
	LargoNotifier <- struct{}{}

}

func MoverProcesoACola(proceso *globals.PCB, colaDestino *[]globals.PCB) {

	// Guardar el estado anterior del proceso
	procesoEstadoAnterior := proceso.Estado

	// Obtener el mutex de la cola de origen
	var mutexOrigen *sync.Mutex
	for colaOrigen, estado := range ColaEstados {
		if proceso.Estado == estado {
			mutexOrigen = ColaMutexes[colaOrigen]
			break
		}
	}

	// Obtener el mutex de la cola de destino
	mutexDestino := ColaMutexes[colaDestino]

	// Bloquear ambas colas (origen y destino)
	if mutexOrigen != nil {
		mutexOrigen.Lock()
		defer mutexOrigen.Unlock()
	}
	if mutexDestino != nil {
		mutexDestino.Lock()
		defer mutexDestino.Unlock()
	}

	// Cambiar el estado del proceso y añadirlo a la cola de destino
	if estadoDestino, ok := ColaEstados[colaDestino]; ok {
		proceso.Estado = estadoDestino

		// Crear una copia del proceso antes de añadirlo a la cola
		procesoCopia := *proceso
		*colaDestino = append(*colaDestino, procesoCopia)
	}

	// Buscar y eliminar el proceso de su cola actual
	for cola, estado := range ColaEstados {
		if procesoEstadoAnterior == estado {
			for i, p := range *cola {
				if p.PID == proceso.PID {
					// Eliminar el proceso de la cola actual
					*cola = append((*cola)[:i], (*cola)[i+1:]...)
					break
				}
			}
			break
		}
	}

	//Probar esto
	//!proceso = BuscarProcesoEnCola(proceso.PID, colaDestino)

	// Actualizar métricas y tiempos si el estado cambió
	if proceso.Estado != procesoEstadoAnterior {

		fmt.Printf("Sume 1 al estado %s del proceso %d\n", proceso.Estado, proceso.PID)
		Logger.Debug("Sume 1 al estado", "estado", proceso.Estado, "pid", proceso.PID)

		// Si el proceso estaba en Exec, guardar el tiempo de la última ráfaga
		if procesoEstadoAnterior == globals.Exec {
			proceso.TiempoDeUltimaRafaga = time.Since(proceso.InicioEstadoActual)
		}

		// Actualizar la métrica de tiempo por estado del proceso
		duracion := time.Since(proceso.InicioEstadoActual)
		proceso.MetricasDeTiempos[procesoEstadoAnterior] += duracion

		// Actualizar la métrica de estado del proceso
		proceso.MetricasDeEstados[proceso.Estado]++

		// Reiniciar el contador de tiempo para el nuevo estado
		proceso.InicioEstadoActual = time.Now()

		LogCambioDeEstado(proceso.PID, string(procesoEstadoAnterior), string(proceso.Estado))
	} else {
		Logger.Debug("$$$$$) El proceso ya estaba en el estado destino", "pid", proceso.PID, "estado", proceso.Estado)
		fmt.Println("$$$$$) El proceso ya estaba en el estado destino", "pid", proceso.PID, "estado", proceso.Estado)
	}
}

func MoverProcesoDeExecABlocked(pid int) {

	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaExec)
	if pcbDelProceso == nil {
		Logger.Debug("Proceso no encontrado en ColaExec", "pid", pid)
		fmt.Println("Proceso no encontrado en ColaExec", "pid", pid)
		return
	}

	MoverProcesoACola(pcbDelProceso, &ColaBlocked)
	if pcbDelProceso.Estado == globals.Blocked {
		IniciarContadorBlocked(pcbDelProceso, Config_Kernel.SuspensionTime)
	}

}

func MoverProcesoDeBlockedAExit(pid int) {

	CancelarContadorBlocked(pid)

	//Busco el PCB del proceso actualizado en la cola de blocked
	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaBlocked)

	// Si no se encuentra el PCB del proceso en la cola de blocked, xq el plani de mediano plazo lo movió a SuspBlocked
	if pcbDelProceso == nil {
		Logger.Debug("Error al buscar el PCB del proceso en la cola de blocked", "pid", pid)

		// Busco el PCB del proceso en la cola de SuspBlocked
		//pcbDelProceso := BuscarProcesoEnCola(pid, &ColaSuspBlocked)

		//! BORRAR EL PROCESO DE SWAP Y LIBERAR LA MEMORIA

		//Lo muevo a la cola exit y lo termino
		TerminarProceso(pid, &ColaSuspBlocked)
	} else {
		// Como lo encontré en la cola de blocked, lo muevo a la cola exit y lo termino
		TerminarProceso(pid, &ColaBlocked)
	}

}

func MoverProcesoDeBlockedAReady(pid int) {

	CancelarContadorBlocked(pid)

	//Busco el PCB del proceso actualizado en la cola de blocked
	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaBlocked)

	// Si no se encuentra el PCB del proceso en la cola de blocked, xq el plani de mediano plazo lo movió a SuspBlocked
	if pcbDelProceso == nil {
		Logger.Debug("Error al buscar el PCB del proceso en la cola de blocked", "pid", pid)
		fmt.Println("Error al buscar el PCB del proceso en la cola de blocked", "pid", pid)

		// Busco el PCB del proceso en la cola de SuspBlocked
		pcbDelProceso := BuscarProcesoEnCola(pid, &ColaSuspBlocked)
		if pcbDelProceso == nil {
			Logger.Debug("Error al buscar el PCB del proceso en la cola de SuspBlocked", "pid", pid)
			fmt.Println("Error al buscar el PCB del proceso en la cola de SuspBlocked", "pid", pid)
		}

		//! SACAR EL PROCESO DE SWAP

		//Lo muevo a la cola destino
		MoverProcesoACola(pcbDelProceso, &ColaSuspReady)
		Logger.Debug("Enviando notificación a LargoNotifier")
		LargoNotifier <- struct{}{}
		Logger.Debug("Notificación enviada a LargoNotifier")
	} else {
		// Como lo encontré en la cola de blocked, lo muevo a la cola destino
		MoverProcesoACola(pcbDelProceso, &ColaReady)
		CortoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
	}

}

func TerminarProceso(pid int, colaOrigen *[]globals.PCB) {

	proceso := BuscarProcesoEnCola(pid, colaOrigen)

	fmt.Println("Terminando proceso con PID:", pid)
	fmt.Println("1)Encontre el proceso:", proceso.PID)

	if !PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory) {
		Logger.Debug("No se puede conectar con memoria (Ping no devuelto)")
		return
	}

	respuestaMemoria := LiberarProcesoEnMemoria(pid)

	if respuestaMemoria {
		MoverProcesoACola(proceso, &ColaExit)
		hayEspacioEnMemoria = true
		LargoNotifier <- struct{}{} // Como se liberó memoria, notificamos al planificador de largo plazo
	}

	proceso = BuscarProcesoEnCola(pid, &ColaExit)

	fmt.Println("2)Encontre el proceso:", proceso.PID)

	LogFinDeProceso(proceso.PID)
	LogMetricasDeEstado(*proceso)
}

func AnalizarDesalojo(cpuId string, pid int, pc int, motivoDesalojo string) {
	mensajePedidoDesalojo := fmt.Sprintf("Analizando desalojo de CPU: ID %s, PID %d, PC %d, Motivo %s", cpuId, pid, pc, motivoDesalojo)
	fmt.Println(mensajePedidoDesalojo)
	Logger.Debug(mensajePedidoDesalojo)

	for i, cpu := range ListaIdentificadoresCPU {
		if cpu.CPUID == cpuId {
			ListaIdentificadoresCPU[i].Ocupado = false
		}
	}

	MutexCpuLibres.Lock()
	CpuLibres = true // Indicamos que hay CPU libres para recibir nuevos procesos
	MutexCpuLibres.Unlock()

	mensajeCpuLiberada := fmt.Sprintf("Se liberó la CPU con ID: %s", cpuId)
	fmt.Println(mensajeCpuLiberada)
	Logger.Debug(mensajeCpuLiberada)

	var pcbDelProceso *globals.PCB
	switch motivoDesalojo {
	case "Planificador":
		LogDesalojoPorSJF_SRT(pid)
		pcbDelProceso = BuscarProcesoEnCola(pid, &ColaExec)
		pcbDelProceso.PC = pc
	case "IO":
		Logger.Debug("Desalojo por IO", "pid", pid)
		pcbDelProceso = BuscarProcesoEnCola(pid, &ColaBlocked)
		if pcbDelProceso == nil {
			pcbDelProceso = BuscarProcesoEnCola(pid, &ColaSuspBlocked)
		}
		pcbDelProceso.PC = pc
	case "DUMP_MEMORY":
		Logger.Debug("Desalojo por DUMP_MEMORY", "pid", pid)
		pcbDelProceso = BuscarProcesoEnCola(pid, &ColaBlocked)
		if pcbDelProceso == nil {
			pcbDelProceso = BuscarProcesoEnCola(pid, &ColaSuspBlocked)
		}
		pcbDelProceso.PC = pc
	case "EXIT":
		Logger.Debug("Desalojo por EXIT", "pid", pid)
	default:
		Logger.Debug("Error, motivo de desalojo no válido", "motivo", motivoDesalojo)
		return
	}
}

func IniciarContadorBlocked(pcb *globals.PCB, milisegundos int) {
	cancel := make(chan struct{})
	canceladoresBlocked[pcb.PID] = cancel

	Logger.Debug("Iniciando contador de blocked para el proceso", "pid", pcb.PID, "tiempo", milisegundos)
	fmt.Println("Iniciando contador de blocked para el proceso", "pid", pcb.PID, "tiempo", milisegundos)

	go func() {
		timer := time.NewTimer(time.Duration(milisegundos) * time.Millisecond)
		select {
		case <-timer.C:
			fmt.Println("Contador de Susp Blocked cumplido para el proceso", pcb.PID)
			// Verifica que el proceso siga en Blocked
			if BuscarProcesoEnCola(pcb.PID, &ColaBlocked) != nil {

				MoverProcesoACola(pcb, &ColaSuspBlocked)

				//! PEDIR A MEMORIA QUE HAGA EL SWAP

				Logger.Debug("Enviando notificación a LargoNotifier")
				LargoNotifier <- struct{}{}
				Logger.Debug("Notificación enviada a LargoNotifier")
			}
		case <-cancel:
			fmt.Println("Contador de Susp Blocked cancelado para el proceso", pcb.PID)
			timer.Stop()
			// El proceso salio de Blocked antes de tiempo
		}
		delete(canceladoresBlocked, pcb.PID)
	}()
}

// Llama a esto cuando el proceso salga de Blocked por otro motivo
func CancelarContadorBlocked(pid int) {
	if cancel, ok := canceladoresBlocked[pid]; ok {
		close(cancel)
		delete(canceladoresBlocked, pid)
	}
}
