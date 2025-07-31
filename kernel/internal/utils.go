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

var Config_Kernel *ConfigKernel

var Logger *slog.Logger

var ContadorPID int = -1

var canceladoresBlocked = make(map[int]chan struct{})
var mutexCanceladoresBlocked sync.Mutex

// &-------------------------------------------Funciones de Kernel-------------------------------------------------------------

func RecibirParametrosConfiguracion() (string, string, int) {
	if len(os.Args) < 4 {
		fmt.Println("Error, mal escrito usa: .kernel/kernel.go [archivo_configuracion] [archivo_pseudocodigo] [tamanio_proceso]")
		os.Exit(1)
	}

	// Leer y cargar la configuración del archivo de configuración
	nombreArchivoConfiguracion := os.Args[1]

	// Leer el nombre del archivo de pseudocódigo
	nombreArchivoPseudocodigo := os.Args[2]

	// Leer y convertir el tamaño del proceso a entero
	tamanioProceso, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Error: el tamaño del proceso debe ser un número entero.", "valor_recibido", os.Args[3])
		os.Exit(1)
	}

	return nombreArchivoConfiguracion, nombreArchivoPseudocodigo, tamanioProceso
}

func IniciarKernel(nombreArchivoConfiguracion string) {
	//Inicializa la config de kernel
	fmt.Println("Iniciando configuración " + nombreArchivoConfiguracion + " de kernel...")

	globals.IniciarConfiguracion("utils/configs/"+nombreArchivoConfiguracion+".json", &Config_Kernel)

	//Crea el archivo donde se logea kernel
	fmt.Println("Iniciando logger...")
	Logger = globals.ConfigurarLogger("kernel", Config_Kernel.LogLevel)

	//Prende el server de kernel en un hilo aparte
	fmt.Println("Iniciando servidor de KERNEL, en el puerto:", Config_Kernel.PortKernel)
	go IniciarServerKernel(Config_Kernel.PortKernel)

	//Realiza el handshake con memoria
	HandshakeConMemoria(Config_Kernel.IPMemory, Config_Kernel.PortMemory)

	//Inicia los planificadores
	IniciarPlanificadores()
	fmt.Println("Kernel iniciado correctamente.")
}

func InicializarProcesoCero(tamanioProceso int, nombreArchivoPseudocodigo string) {

	Logger.Debug("Inicializando Proceso Cero",
		"tamanio_proceso", tamanioProceso,
		"nombre_archivo_pseudocodigo", nombreArchivoPseudocodigo)

	InicializarPCB(tamanioProceso, nombreArchivoPseudocodigo)

}

func BuscarProcesoEnCola(pid int, cola *[]*globals.PCB) *globals.PCB {
	for i, proceso := range *cola {
		if proceso.PID == pid {
			return (*cola)[i]
		}
	}
	return nil
}

func InicializarPCB(tamanioEnMemoria int, nombreArchivoPseudo string) {
	mensajeInicializandoPCB := "Inicializando PCB" + " tamanio_en_memoria=" + strconv.Itoa(tamanioEnMemoria) + ", nombre_archivo_pseudo=" + nombreArchivoPseudo
	Logger.Debug(mensajeInicializandoPCB)
	//fmt.Println(mensajeInicializandoPCB)

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
		InicioEjecucion:      time.Time{},
		DesalojoAnalizado:    true,
	}

	// Bloquear el mutex de la cola de New
	ColaMutexes[&ColaNew].Lock()

	// Agregar el proceso a la cola de New
	pcb.MetricasDeEstados[globals.New] = 1
	pcb.MetricasDeTiempos[globals.New] = time.Since(pcb.InicioEstadoActual)
	ColaNew = append(ColaNew, &pcb)

	// Desbloquear el mutex de la cola de New
	ColaMutexes[&ColaNew].Unlock()

	//fmt.Println("PID", ContadorPID, " creado. Procesos en Cola NEW despues de mover el proceso:", len(ColaNew))
	LogCreacionDeProceso(ContadorPID)

	// Al agregar un nuevo proceso a la cola de New, notificamos al planificador de largo plazo
	select {
	case LargoNotifier <- struct{}{}:
		// señal enviada
	default:
		// ya hay una señal pendiente, no enviar otra
	}
}

func MoverProcesoACola(proceso *globals.PCB, colaDestino *[]*globals.PCB) {

	// Guardar el estado anterior del proceso
	procesoEstadoAnterior := proceso.Estado

	// Obtener el mutex de la cola de origen
	var mutexOrigen *sync.Mutex
	for colaOrigen, estado := range ColaEstados {
		//fmt.Println("ITERO FOR 1 MOVER COLA ")
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

	// Verificar si el proceso ya está en la cola destino
	for _, p := range *colaDestino {
		//fmt.Println("ITERO FOR 2 MOVER COLA ")
		if p.PID == proceso.PID {
			fmt.Printf("El proceso ya está en la cola destino: PID %d\n", proceso.PID)
			return
		}
	}

	// Cambiar el estado del proceso y añadirlo a la cola de destino
	if estadoDestino, ok := ColaEstados[colaDestino]; ok {
		proceso.Estado = estadoDestino
		*colaDestino = append(*colaDestino, proceso)
	}

	// Buscar y eliminar el proceso de su cola actual
	for cola, estado := range ColaEstados {
		//fmt.Println("ITERO FOR 3 MOVER COLA ")

		if procesoEstadoAnterior == estado {
			for i := range *cola {
				if (*cola)[i].PID == proceso.PID {
					fmt.Printf("Eliminando proceso de la cola origen: PID %d, Estado %s\n", proceso.PID, procesoEstadoAnterior)
					*cola = append((*cola)[:i], (*cola)[i+1:]...)
					break
				}
			}
			break
		}
	}

	// Actualizar métricas y tiempos si el estado cambió
	if proceso.Estado != procesoEstadoAnterior {

		// Genero una variable para indicar si el proceso fue reestimado
		procesoReestimado := false

		// Si el proceso sale de Exec (abandona la CPU)
		if procesoEstadoAnterior == globals.Exec {
			// Guardar el tiempo de la última ráfaga (cuanto tiempo ejecutó en CPU)
			proceso.TiempoDeUltimaRafaga = float64(time.Since(proceso.InicioEjecucion).Milliseconds())

			Logger.Debug("Guardando tiempo de última ráfaga", "pid", proceso.PID, "tiempo", proceso.TiempoDeUltimaRafaga)
			//fmt.Println("Guardando tiempo de última ráfaga", "pid", proceso.PID, "tiempo", proceso.TiempoDeUltimaRafaga)

			// Si el proceso sale de Exec y va directo a Ready es por desalojo de SRT
			if proceso.Estado == globals.Ready && (AlgoritmoCortoPlazo == "SRT" || AlgoritmoCortoPlazo == "SJF") {
				reestimarProceso(proceso, "Planificador")
				procesoReestimado = true
			}

		}

		// Si el proceso entra a Ready tengo que reestimar
		if proceso.Estado == globals.Ready && !procesoReestimado && (AlgoritmoCortoPlazo == "SRT" || AlgoritmoCortoPlazo == "SJF") {
			reestimarProceso(proceso, "IO/DUMP")
		}

		// Si el proceso entra a Exec (entra en la CPU)
		if proceso.Estado == globals.Exec {
			proceso.InicioEjecucion = time.Now()
			proceso.DesalojoAnalizado = false // Reiniciar el análisis de desalojo al entrar en Exec
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
		Logger.Debug("El proceso ya estaba en el estado destino", "pid", proceso.PID, "estado", proceso.Estado)
		//fmt.Println("El proceso ya estaba en el estado destino", "pid", proceso.PID, "estado", proceso.Estado)
	}
}

func MoverProcesoDeExecABlocked(pid int) {

	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaExec)
	if pcbDelProceso == nil {
		Logger.Debug("Proceso no encontrado en ColaExec", "pid", pid)
		//fmt.Println("Proceso no encontrado en ColaExec", "pid", pid)
		return
	}

	MoverProcesoACola(pcbDelProceso, &ColaBlocked)
	if pcbDelProceso.Estado == globals.Blocked {
		IniciarContadorBlocked(pcbDelProceso, Config_Kernel.SuspensionTime)
	}

}

func MoverProcesoDeBlockedAExit(pid int) {

	CancelarContadorBlocked(pid)

	// Si no se encuentra el PCB del proceso en la cola de blocked, xq el plani de mediano plazo lo movió a SuspBlocked
	if BuscarProcesoEnCola(pid, &ColaBlocked) == nil {
		Logger.Debug("Error al buscar el PCB del proceso en la cola de blocked", "pid", pid)
		//fmt.Println("Error al buscar el PCB del proceso en la cola de blocked", "pid", pid)

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
		// Busco el PCB del proceso en la cola de SuspBlocked
		pcbDelProceso := BuscarProcesoEnCola(pid, &ColaSuspBlocked)
		if pcbDelProceso == nil {
			Logger.Debug("No se encontro al buscar el PCB del proceso en la cola de SuspBlocked", "pid", pid)
			//fmt.Println("No se encontro al buscar el PCB del proceso en la cola de SuspBlocked", "pid", pid)
			return
		}

		//Si el DesSwap fue bien Lo muevo a la cola SuspReady
		MoverProcesoACola(pcbDelProceso, &ColaSuspReady)
		Logger.Debug("Enviando notificación a LargoNotifier")
		select {
		case LargoNotifier <- struct{}{}:
			// señal enviada
		default:
			// ya hay una señal pendiente, no enviar otra
		}
		Logger.Debug("Notificación enviada a LargoNotifier")

	} else {
		// Como lo encontré en la cola de blocked, lo muevo a la cola destino

		MoverProcesoACola(pcbDelProceso, &ColaReady)
		select {
		case CortoNotifier <- struct{}{}:
			// señal enviada
		default:
			// ya hay una señal pendiente, no enviar otra
		}
	}

}

func TerminarProceso(pid int, colaOrigen *[]*globals.PCB) {

	proceso := BuscarProcesoEnCola(pid, colaOrigen)
	if proceso == nil {
		Logger.Debug("Proceso no encontrado en la cola de origen al terminar proceso", "pid", pid)
		//fmt.Println("Proceso no encontrado en la cola de origen al terminar proceso", "pid", pid)
		return
	}

	if !PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory) {
		Logger.Debug("No se puede conectar con memoria (Ping no devuelto)")
		return
	}

	MoverProcesoACola(proceso, &ColaExit)
	respuestaMemoria := LiberarProcesoEnMemoria(pid)
	if respuestaMemoria {
		select {
		case LargoNotifier <- struct{}{}:
			// señal enviada
		default:
			// ya hay una señal pendiente, no enviar otra
		}
	} else {
		//fmt.Println("Error al liberar memoria del proceso:", pid)
	}

	proceso = BuscarProcesoEnCola(pid, &ColaExit)

	LogFinDeProceso(proceso.PID)
	LogMetricasDeEstado(*proceso)
}

var CpuLiberada bool = true // Variable para indicar si la CPU fue liberada por el proceso que estaba ejecutando
var MutexCpuLiberada sync.Mutex

func AnalizarDesalojo(cpuId string, pid int, pc int, motivoDesalojo string) {
	mensajePedidoDesalojo := fmt.Sprintf("Analizando desalojo de CPU: ID %s, PID %d, PC %d, Motivo %s", cpuId, pid, pc, motivoDesalojo)
	//fmt.Println(mensajePedidoDesalojo)
	Logger.Debug(mensajePedidoDesalojo)

	var pcbDelProceso *globals.PCB
	switch motivoDesalojo {
	case "Planificador":
		LogDesalojoPorSJF_SRT(pid)
		pcbDelProceso = BuscarProcesoEnCola(pid, &ColaReady)
		if pcbDelProceso == nil {
			//fmt.Println("No se encontró el proceso en ColaReady, buscando en Exec")
			pcbDelProceso = BuscarProcesoEnCola(pid, &ColaExec)
			if pcbDelProceso == nil {
				//fmt.Println("No se encontró el proceso en ColaExec, buscando en Blocked")
				pcbDelProceso = BuscarProcesoEnCola(pid, &ColaBlocked)
				if pcbDelProceso == nil {
					//fmt.Println("No se encontró el proceso en ColaBlocked, buscando en SuspBlocked")
					pcbDelProceso = BuscarProcesoEnCola(pid, &ColaSuspBlocked)
					if pcbDelProceso == nil {
						//fmt.Println("No se encontró el proceso en ColaSuspBlocked, buscando en ColaExit")
						pcbDelProceso = BuscarProcesoEnCola(pid, &ColaExit)
					}
				}
			}
		}
		pcbDelProceso.PC = pc
		pcbDelProceso.DesalojoAnalizado = true
	case "IO":
		Logger.Debug("Desalojo por IO", "pid", pid)
		pcbDelProceso = BuscarProcesoEnCola(pid, &ColaBlocked)
		if pcbDelProceso == nil {
			pcbDelProceso = BuscarProcesoEnCola(pid, &ColaSuspBlocked)
			if pcbDelProceso == nil {
				pcbDelProceso = BuscarProcesoEnCola(pid, &ColaReady)
				if pcbDelProceso == nil {
					pcbDelProceso = BuscarProcesoEnCola(pid, &ColaExit)
				}
			}
		}
		pcbDelProceso.PC = pc
		pcbDelProceso.DesalojoAnalizado = true
	case "DUMP_MEMORY":
		Logger.Debug("Desalojo por DUMP_MEMORY", "pid", pid)
		pcbDelProceso = BuscarProcesoEnCola(pid, &ColaReady)
		if pcbDelProceso == nil {
			pcbDelProceso = BuscarProcesoEnCola(pid, &ColaExit)
		}
		pcbDelProceso.PC = pc
		pcbDelProceso.DesalojoAnalizado = true
	case "EXIT":
		Logger.Debug("Desalojo por EXIT", "pid", pid)
	default:
		Logger.Debug("Error, motivo de desalojo no válido", "motivo", motivoDesalojo)
		return
	}

	for i, cpu := range ListaIdentificadoresCPU {

		//fmt.Println("ITERO FOR ANALIZAR DESALOJO ")
		if cpu.CPUID == cpuId {
			MutexIdentificadoresCPU.Lock()
			ListaIdentificadoresCPU[i].Ocupado = false
			ListaIdentificadoresCPU[i].DesalojoSolicitado = false
			MutexIdentificadoresCPU.Unlock()

			MutexCpuLibres.Lock()
			CpuLibres = true // Indicamos que hay CPU libres para recibir nuevos procesos
			MutexCpuLibres.Unlock()

			MutexCpuLiberada.Lock()
			CpuLiberada = true // Variable para indicar si la CPU fue liberada por el proceso que estaba ejecutando
			MutexCpuLiberada.Unlock()
			select {
			case CortoNotifier <- struct{}{}:
				// señal enviada
			default:
				// ya hay una señal pendiente, no enviar otra
			}
		}
	}

	mensajeCpuLiberada := fmt.Sprintf("Se liberó la CPU con ID: %s", cpuId)
	//fmt.Println(mensajeCpuLiberada)
	Logger.Debug(mensajeCpuLiberada)

	select {
	case CortoNotifier <- struct{}{}:
		// señal enviada
	default:
		// ya hay una señal pendiente, no enviar otra
	}
}

func IniciarContadorBlocked(pcb *globals.PCB, milisegundos int) {
	cancel := make(chan struct{})

	// Bloquea el acceso al mapa antes de modificarlo
	mutexCanceladoresBlocked.Lock()
	canceladoresBlocked[pcb.PID] = cancel
	mutexCanceladoresBlocked.Unlock()

	pid := pcb.PID // Captura el valor de PID localmente

	go func(pidLocal int, cancel chan struct{}) {
		Logger.Debug("Iniciando contador de blocked para el proceso", "pid", pidLocal, "tiempo", milisegundos)
		//fmt.Println("Iniciando contador de blocked para el proceso", "pid", pidLocal, "tiempo", milisegundos)

		timer := time.NewTimer(time.Duration(milisegundos) * time.Millisecond)
		select {
		case <-timer.C:
			//fmt.Println("Contador de Susp Blocked cumplido para el proceso", pidLocal)
			if BuscarProcesoEnCola(pidLocal, &ColaBlocked) != nil {
				MoverProcesoACola(pcb, &ColaSuspBlocked)
				// ENVIAMOS EL PROCESO A SWAP
				PedirSwapping(pcb.PID)
				Logger.Debug("Enviando notificación a LargoNotifier")
				select {
				case LargoNotifier <- struct{}{}:
					// señal enviada
				default:
					// ya hay una señal pendiente, no enviar otra
				}
			}
		case <-cancel:
			//fmt.Println("Contador de Susp Blocked cancelado para el proceso", pidLocal)
			timer.Stop()
		}
	}(pid, cancel) // Pasa el PID y el canal como argumentos a la goroutine
}

func CancelarContadorBlocked(pid int) {
	mutexCanceladoresBlocked.Lock()
	if cancel, ok := canceladoresBlocked[pid]; ok {
		//fmt.Println("Cancelando contador para PID:", pid)
		close(cancel)
		delete(canceladoresBlocked, pid)
	} else {
		//fmt.Println("No se encontró un contador para PID:", pid)
	}
	mutexCanceladoresBlocked.Unlock()
}
