package kernel_internal

import (
	"fmt"
	"log"
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

var hayEspacioEnMemoria bool = true

var ColaEstados = map[*[]globals.PCB]globals.Estado{
	&ColaNew:         globals.Estado("NEW"),
	&ColaReady:       globals.Estado("READY"),
	&ColaExec:        globals.Estado("EXEC"),
	&ColaBlocked:     globals.Estado("BLOCKED"),
	&ColaSuspReady:   globals.Estado("SUSP_READY"),
	&ColaSuspBlocked: globals.Estado("SUSP_BLOCKED"),
	&ColaExit:        globals.Estado("EXIT"),
}

func IniciarPlanificador() {

	//? 3. Cuando memo diga no, se termina el bucle hasta q se termine un proceso
	//?    CUANDO PASEMOS ALGUN PROCESO A EXIT HACER MEMO = TRUE Y LLAMAR AL PLANIFICADOR
	switch Config_Kernel.SchedulerAlgorithm {
	case "FIFO":
		{
			//& Planificador largo Plazo
			PlanificadorLargoPlazo("FIFO")
			break
		}
	default:
		{
		}

	}
}

func PlanificadorLargoPlazo(algoritmo string) {

	for hayEspacioEnMemoria {
		if algoritmo == "FIFO" {

			if len(ColaNew) != 0 {
				if !client.PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory) {
					log.Println("No se puede conectar con memoria (Ping no devuelto)")
					return
				}

				respuestaMemoria := PedirEspacioAMemoria(ColaNew[0])

				if respuestaMemoria {
					MoverProcesoACola(ColaNew[0], &ColaReady)
				} else {
					hayEspacioEnMemoria = false
				}
			} else {
				hayEspacioEnMemoria = false
			}
		}

	}

}

func TerminarProceso(proceso globals.PCB) {

	if !client.PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory) {
		log.Println("No se puede conectar con memoria (Ping no devuelto)")
		return
	}

	respuestaMemoria := LiberarProcesoEnMemoria(proceso.PID)
	if respuestaMemoria {
		MoverProcesoACola(proceso, &ColaExit)
		hayEspacioEnMemoria = true
		PlanificadorLargoPlazo(Config_Kernel.SchedulerAlgorithm)
	}

}

func InicializarPCB(pid int, tamanioEnMemoria int) {

	pcb := globals.PCB{
		PID:               pid,
		PC:                0,
		Estado:            globals.New,
		MetricasDeEstados: make(map[globals.Estado]int),
		MetricasDeTiempos: make(map[globals.Estado]int),
		TamanioEnMemoria:  tamanioEnMemoria,
	}

	LogCreacionDeProceso(pid)
	MoverProcesoACola(pcb, &ColaNew)

}

func MoverProcesoACola(proceso globals.PCB, colaDestino *[]globals.PCB) {

	procesoEstadoAnterior := proceso.Estado

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

	InicializarPCB(1, 1024)

	fmt.Println("\n---------------------------------------")
	fmt.Println("Cola New antes:", ColaNew)
	fmt.Println("Cola Ready antes:", ColaReady)
	fmt.Println("---------------------------------------\n ")

	PlanificadorLargoPlazo("FIFO")

	fmt.Println("\n---------------------------------------")
	fmt.Println("Cola New despues:", ColaNew)
	fmt.Println("Cola Ready despues:", ColaReady)
	fmt.Println("---------------------------------------\n ")

}
