package kernel_internal

import (
	"fmt"
	"sync"
	"time"
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
	for hayEspacioEnMemoria && (len(ColaNew) != 0 || len(ColaSuspReady) != 0) {
		if algoritmo == "FIFO" {
			// Si memoria responde...
			if !client.PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory, Logger) {
				Logger.Debug("No se puede conectar con memoria (Ping no devuelto)")
				return
			}

			//! Estaria piola hacer una funcion que resuma todo esto, pero por ahora lo dejo asi
			//Verifico si hay procesos en la cola SuspReady
			if len(ColaSuspReady) != 0 {

				// Pido espacio en memoria para el primer proceso de la cola SuspReady
				respuestaMemoria := PedirEspacioAMemoria(ColaSuspReady[0])

				// Si memoria responde que no hay espacio..
				if !respuestaMemoria {
					hayEspacioEnMemoria = false // Seteo la variable del for a false
					return                      // Salgo del for
				}

				MoverProcesoACola(ColaSuspReady[0], &ColaReady)
			} else { //Si no hay procesos en SuspReady, ya se que hay en New

				// Pido espacio en memoria para el primer proceso de la cola New
				respuestaMemoria := PedirEspacioAMemoria(ColaNew[0])

				// Si memoria responde que no hay espacio...
				if !respuestaMemoria {
					hayEspacioEnMemoria = false // Seteo la variable del for a false
					return                      // Salgo del for
				}

				MoverProcesoACola(ColaNew[0], &ColaReady)
			}

			PlanificadorCortoPlazo(Config_Kernel.ReadyIngressAlgorithm)
		}

		if algoritmo == "PMCP" {
			// Si memoria responde...
			if !client.PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory, Logger) {
				Logger.Debug("No se puede conectar con memoria (Ping no devuelto)")
				return
			}

			//Verifico si hay procesos en la cola SuspReady
			if len(ColaSuspReady) != 0 {
				// Pido espacio en memoria para el primer proceso de la cola New
				respuestaMemoria := PedirEspacioAMemoria(ColaSuspReady[pcbMasChico()])

				// Si memoria responde que no hay espacio...
				if !respuestaMemoria {
					hayEspacioEnMemoria = false // Seteo la variable del for a false
					return                      // Salgo del for
				}

				MoverProcesoACola(ColaSuspReady[pcbMasChico()], &ColaReady)
			} else {
				// Pido espacio en memoria para el primer proceso de la cola New
				respuestaMemoria := PedirEspacioAMemoria(ColaNew[pcbMasChico()])

				// Si memoria responde que no hay espacio...
				if !respuestaMemoria {
					hayEspacioEnMemoria = false // Seteo la variable del for a false
					return                      // Salgo del for
				}

				MoverProcesoACola(ColaNew[pcbMasChico()], &ColaReady)
			}
			PlanificadorCortoPlazo(Config_Kernel.ReadyIngressAlgorithm)
		}
	}
}

func PlanificadorCortoPlazo(algoritmo string) {
	for len(ColaReady) != 0 {
		if algoritmo == "FIFO" {
			MoverProcesoACola(ColaReady[0], &ColaExec)
			ultimo := ColaExec[len(ColaExec)-1]
			if !ElegirCpuYMandarProceso(ultimo) {
				// No se pudo enviar el proceso a la CPU, lo devolvemos a la cola Ready
				MoverProcesoACola(ultimo, &ColaReady)
				return
			}
		}
		if algoritmo == "SJF" {
			for _, proceso := range ColaReady {
				if proceso.MetricasDeEstados[globals.Exec] != 0 && proceso.EstimacionDeRafaga.YaCalculado == false {
					//Est(n+1) =  alfa.R(n) + (1-alfa).Est(n) ;   alfa pertenece [0,1]
					proceso.EstimacionDeRafaga.TiempoDeRafaga = (Config_Kernel.Alpha * float64(proceso.TiempoDeUltimaRafaga.Milliseconds())) + (1-Config_Kernel.Alpha) * proceso.EstimacionDeRafaga.TiempoDeRafaga
					proceso.EstimacionDeRafaga.YaCalculado = true
				}
			}
			pcbElegido := elegirPcbConEstimacionMasChica()
			pcbElegido.EstimacionDeRafaga.YaCalculado = false
			MoverProcesoACola(pcbElegido, &ColaExec)
			ultimo := ColaExec[len(ColaExec)-1]
			if !ElegirCpuYMandarProceso(ultimo) {
				// No se pudo enviar el proceso a la CPU, lo devolvemos a la cola Ready
				MoverProcesoACola(ultimo, &ColaReady)
				return
			}
		}
		}
		if algoritmo == "SRT" {
			//* Lógica para SJF con desalojo /SRT
		}
}

func elegirPcbConEstimacionMasChica() globals.PCB {
	var maxIndex float64
	var pcb globals.PCB
	for i, num := range ColaReady {
		if i == 0 {
			maxIndex = num.EstimacionDeRafaga.TiempoDeRafaga
			pcb = num
		} else if num.EstimacionDeRafaga.TiempoDeRafaga < maxIndex {
			maxIndex = num.EstimacionDeRafaga.TiempoDeRafaga
			pcb = num
		}
	}
	return pcb
}

func MoverProcesoACola(proceso globals.PCB, colaDestino *[]globals.PCB) {
	//& El mutex actualmente lo estamos usando con todas las colas, pero seguramente estamos haciendo mucho overhead
	//& porque no todas las colas se usan al mismo tiempo. Hay que ver si podemos optimizar eso.

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
		mutexOrigen.Lock()         // Bloquear mutexOrigen si no es nil
		defer mutexOrigen.Unlock() // Defer para desbloquear al final de la función
	}
	if mutexDestino != nil {
		mutexDestino.Lock()         // Bloquear mutexDestino
		defer mutexDestino.Unlock() // Defer para desbloquear al final de la función
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
		*colaDestino = append(*colaDestino, proceso)
	}

	if proceso.Estado != procesoEstadoAnterior {

		// Si el proceso estaba en Exec, hay que guardar el tiempo de la última ráfaga
		if procesoEstadoAnterior == globals.Exec {
			proceso.TiempoDeUltimaRafaga = time.Since(proceso.InicioEstadoActual) // Toma el tiempo transcurrido desde que el proceso entró al estado Exec
		}

		// Actualizar la métrica de tiempo por estado del proceso
		duracion := time.Since(proceso.InicioEstadoActual)           // Toma el tiempo transcurrido desde que el proceso entró al estado origen
		proceso.MetricasDeTiempos[procesoEstadoAnterior] += duracion // Actualiza el tiempo acumulado en el estado anterior

		// Actualizar la métrica de estado del proceso
		proceso.MetricasDeEstados[proceso.Estado]++

		// Reiniciar el contador de tiempo para el nuevo estado
		proceso.InicioEstadoActual = time.Now()

		LogCambioDeEstado(proceso.PID, string(procesoEstadoAnterior), string(proceso.Estado))
	}

}

func MoverProcesoABlocked(pid int) {

	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaExec)

	MoverProcesoACola(*pcbDelProceso, &ColaBlocked)

	IniciarContadorBlocked(*pcbDelProceso, Config_Kernel.SuspensionTime)

}

func MoverProcesoDeBlockedAExit(pid int) {

	CancelarContadorBlocked(pid)

	//Busco el PCB del proceso actualizado en la cola de blocked
	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaBlocked)

	// Si no se encuentra el PCB del proceso en la cola de blocked, xq el plani de mediano plazo lo movió a SuspBlocked
	if pcbDelProceso == nil {
		Logger.Debug("Error al buscar el PCB del proceso en la cola de blocked", "pid", pid)

		// Busco el PCB del proceso en la cola de SuspBlocked
		pcbDelProceso := BuscarProcesoEnCola(pid, &ColaSuspBlocked)

		//! BORRAR EL PROCESO DE SWAP Y LIBERAR LA MEMORIA

		//Lo muevo a la cola destino
		MoverProcesoACola(*pcbDelProceso, &ColaExit)
	} else {
		// Como lo encontré en la cola de blocked, lo muevo a la cola destino
		MoverProcesoACola(*pcbDelProceso, &ColaExit)
	}

}

func MoverProcesoDeBlockedAReady(pid int) {

	CancelarContadorBlocked(pid)

	//Busco el PCB del proceso actualizado en la cola de blocked
	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaBlocked)

	// Si no se encuentra el PCB del proceso en la cola de blocked, xq el plani de mediano plazo lo movió a SuspBlocked
	if pcbDelProceso == nil {
		Logger.Debug("Error al buscar el PCB del proceso en la cola de blocked", "pid", pid)

		// Busco el PCB del proceso en la cola de SuspBlocked
		pcbDelProceso := BuscarProcesoEnCola(pid, &ColaSuspBlocked)

		//! SACAR EL PROCESO DE SWAP

		//Lo muevo a la cola destino
		MoverProcesoACola(*pcbDelProceso, &ColaSuspReady)
	} else {
		// Como lo encontré en la cola de blocked, lo muevo a la cola destino
		MoverProcesoACola(*pcbDelProceso, &ColaReady)
	}

}

func pcbMasChico() int {
	//& Esta función es un ejemplo de cómo se podría implementar el algoritmo PMCP (Proceso Más Chiquito Primero)
	//& que se basa en el tamaño del PCB para decidir cuál proceso mover a la cola Ready.

	// Encontrar el PCB más chico
	minIndex := 0
	for i := 1; i < len(ColaNew); i++ {
		if ColaNew[i].TamanioEnMemoria < ColaNew[minIndex].TamanioEnMemoria {
			minIndex = i
		}
	}
	return minIndex
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
