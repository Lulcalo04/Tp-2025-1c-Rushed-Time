package kernel_internal

import (
	"fmt"
	"utils/globals"
)

var ColaNew []globals.PCB
var ColaReady []globals.PCB
var ColaExec []globals.PCB
var ColaBlocked []globals.PCB
var ColaSuspReady []globals.PCB
var ColaSuspBlocked []globals.PCB
var ColaExit []globals.PCB

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

	/* switch(Config_Kernel.SchedulerAlgorithm){
	case "FIFO":
	{
		//& Planificador largo Plazo

	} */
}

func PlanificadorLargoPlazo(algoritmo string) {

	if algoritmo == "FIFO" {

		//! PEDIR ESPACIO EN MEMORIA

		respuestaMemoria := true

		if respuestaMemoria {

			MoverProcesoACola(ColaNew[0], &ColaReady)
		}
		/* else {
			//Dejar Esperando al proceso
		} */

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

	// Crear 3 PCBs de ejemplo
	InicializarPCB(1, 1024)
	InicializarPCB(2, 2048)
	InicializarPCB(3, 4096)

	// Imprimir las colas para verificar
	fmt.Println("Cola New:", ColaNew)
	fmt.Println("Cola Ready:", ColaReady)

	MoverProcesoACola(ColaNew[0], &ColaReady)

	// Imprimir las colas después de agregar un proceso a la cola
	fmt.Println("Cola New después de agregar a Ready:", ColaNew)
	fmt.Println("Cola Ready después de agregar:", ColaReady)

	MoverProcesoACola(ColaNew[0], &ColaReady)

	fmt.Println("Cola New después de agregar a Ready:", ColaNew)
	fmt.Println("Cola Ready después de agregar:", ColaReady)

}
