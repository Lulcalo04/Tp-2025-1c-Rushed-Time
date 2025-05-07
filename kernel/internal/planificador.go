package kernel_internal

import (
	"fmt"
	"sync"
	"utils/client"
	"utils/globals"
)

var ColaNew []globals.PCB
var ColaReady []globals.PCB
var ColaExec []globals.PCB
var ColaBlocked []globals.PCB
var ColaSuspReady []globals.PCB
var ColaSuspBlocked []globals.PCB
var ColaExit []globals.PCB

//var algoritmoCortoPlazo string
//var algoritmoLargoPlazo string

var ColaEstados = map[*[]globals.PCB]globals.Estado{
	&ColaNew:         globals.Estado("NEW"),
	&ColaReady:       globals.Estado("READY"),
	&ColaExec:        globals.Estado("EXEC"),
	&ColaBlocked:     globals.Estado("BLOCKED"),
	&ColaSuspReady:   globals.Estado("SUSP_READY"),
	&ColaSuspBlocked: globals.Estado("SUSP_BLOCKED"),
	&ColaExit:        globals.Estado("EXIT"),
}

var mutexNew sync.Mutex
var mutexReady sync.Mutex
var mutexExec sync.Mutex
var mutexBlocked sync.Mutex
var mutexSuspReady sync.Mutex
var mutexSuspBlocked sync.Mutex
var mutexExit sync.Mutex

var ColaMutexes = map[*[]globals.PCB]*sync.Mutex{
	&ColaNew:         &mutexNew,
	&ColaReady:       &mutexReady,
	&ColaExec:        &mutexExec,
	&ColaBlocked:     &mutexBlocked,
	&ColaSuspReady:   &mutexSuspReady,
	&ColaSuspBlocked: &mutexSuspBlocked,
	&ColaExit:        &mutexExit,
}

var hayEspacioEnMemoria bool = true

func IniciarPlanificadores() {

	PlanificadorLargoPlazo(Config_Kernel.SchedulerAlgorithm)
	PlanificadorCortoPlazo(Config_Kernel.ReadyIngressAlgorithm)

}

func PlanificadorLargoPlazo(algoritmo string) {
	// Mientras hay
	for hayEspacioEnMemoria && len(ColaNew) != 0 {
		if algoritmo == "FIFO" {
			// Si memoria responde...
			if !client.PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory, Logger) {
				Logger.Debug("No se puede conectar con memoria (Ping no devuelto)")
				return
			}
			// Pido espacio en memoria para el primer proceso de la cola New
			respuestaMemoria := PedirEspacioAMemoria(ColaNew[0])
			// Si memoria responde que no hay espacio...
			if !respuestaMemoria {
				hayEspacioEnMemoria = false // Seteo la variable del for a false
				return                      // Salgo del for
			}
			MoverProcesoACola(ColaNew[0], &ColaReady)
		}
		if algoritmo == "PMCP" {
			//! Lógica para PMCP
		}
	}
}

func PlanificadorCortoPlazo(algoritmo string) {
	for len(ColaReady) != 0 {
		if algoritmo == "FIFO" {
			MoverProcesoACola(ColaReady[0], &ColaExec)
		}
		if algoritmo == "SJF" {
			//! Lógica para SJF sin desalojo
		}
		if algoritmo == "SRT" {
			//! Lógica para SJF con desalojo /SRT
		}
	}
}

func MoverProcesoACola(proceso globals.PCB, colaDestino *[]globals.PCB) {
	// ! El mutex actualmente lo estamos usando con todas las colas, pero seguramente estamos haciendo mucho overhead
	// ! porque no todas las colas se usan al mismo tiempo. Hay que ver si podemos optimizar eso.
	// Guardar el estado anterior del proceso
	procesoEstadoAnterior := proceso.Estado

	// Obtener el mutex de la cola de origen
	var mutexOrigen *sync.Mutex
	for cola, estado := range ColaEstados {
		if proceso.Estado == estado {
			mutexOrigen = ColaMutexes[cola]
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

	// Buscar y eliminar el proceso de su cola actual
	for cola, estado := range ColaEstados {
		if proceso.Estado == estado {
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

	// Agregar el proceso a la cola destino
	if estadoDestino, ok := ColaEstados[colaDestino]; ok {
		proceso.Estado = estadoDestino
	}

	if proceso.Estado != procesoEstadoAnterior {
		LogCambioDeEstado(proceso.PID, string(procesoEstadoAnterior), string(proceso.Estado))
	}

	*colaDestino = append(*colaDestino, proceso)
}

func Prueba() {

	InicializarPCB(1024)

	fmt.Println("Antes del plani de largo plazo...")

	fmt.Println("\n---------------------------------------")
	fmt.Println("Cola New antes:", ColaNew)
	fmt.Println("Cola Ready antes:", ColaReady)
	fmt.Println("---------------------------------------\n ")

	fmt.Println("Despues del plani de largo plazo...")

	PlanificadorLargoPlazo("FIFO")

	fmt.Println("\n---------------------------------------")
	fmt.Println("Cola New despues:", ColaNew)
	fmt.Println("Cola Ready despues:", ColaReady)
	fmt.Println("---------------------------------------\n ")

	fmt.Println("Despues del plani de corto plazo...")

	PlanificadorCortoPlazo("FIFO")

	fmt.Println("\n---------------------------------------")
	fmt.Println("Cola Ready:", ColaReady)
	fmt.Println("Cola Exec:", ColaExec)
	fmt.Println("---------------------------------------\n ")

}
