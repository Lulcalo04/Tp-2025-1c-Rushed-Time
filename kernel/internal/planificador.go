package kernel_internal

import (
	"bufio"
	"os"
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
var MutexCpuLibres sync.Mutex
var CpuLibres bool = true

func IniciarPlanificadores() {

	//ESPERAR ENTER
	Logger.Debug("Presione ENTER para iniciar los planificadores")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	Logger.Debug("Iniciando planificadores...")

	go PlanificadorLargoPlazo()
	go PlanificadorCortoPlazo()
}

var planificadorLargoMutex sync.Mutex // Mutex para evitar doble planificador de largo plazo
var LargoNotifier = make(chan struct{})

func PlanificadorLargoPlazo() {

	algoritmo := Config_Kernel.SchedulerAlgorithm
	for {
		planificadorLargoMutex.Lock()
		<-LargoNotifier
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
					CortoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
				} else { //Si no hay procesos en SuspReady, ya se que hay en New

					// Pido espacio en memoria para el primer proceso de la cola New
					respuestaMemoria := PedirEspacioAMemoria(ColaNew[0])

					// Si memoria responde que no hay espacio...
					if !respuestaMemoria {
						hayEspacioEnMemoria = false // Seteo la variable del for a false
						return                      // Salgo del for
					}

					MoverProcesoACola(ColaNew[0], &ColaReady)
					CortoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
				}

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
					CortoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
				} else {
					// Pido espacio en memoria para el primer proceso de la cola New
					respuestaMemoria := PedirEspacioAMemoria(ColaNew[pcbMasChico()])

					// Si memoria responde que no hay espacio...
					if !respuestaMemoria {
						hayEspacioEnMemoria = false // Seteo la variable del for a false
						return                      // Salgo del for
					}

					MoverProcesoACola(ColaNew[pcbMasChico()], &ColaReady)
					CortoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
				}

			}
		}
		planificadorLargoMutex.Unlock()
	}
}

var planificadorCortoMutex sync.Mutex // Mutex para evitar doble planificador de corto plazo
var CortoNotifier = make(chan struct{})

func PlanificadorCortoPlazo() {

	algoritmo := Config_Kernel.ReadyIngressAlgorithm
	for {
		planificadorCortoMutex.Lock()
		<-CortoNotifier

		for len(ColaReady) != 0 {
			if algoritmo == "FIFO" && CpuLibres {
				// Seleccionamos el primer proceso de la cola Ready y lo enviamos a Exec
				MoverProcesoACola(ColaReady[0], &ColaExec)

				// Sabemos que el proceso que acabamos de mover a Exec es el último de la cola
				ultimo := ColaExec[len(ColaExec)-1]

				// Intentamos enviar el proceso a la CPU
				if !ElegirCpuYMandarProceso(ultimo) {
					// No se pudo enviar el proceso a la CPU, lo devolvemos a la cola Ready
					MoverProcesoACola(ultimo, &ColaReady)
					MutexCpuLibres.Lock()
					CpuLibres = false // Indicamos que la CPU no está libre
					MutexCpuLibres.Unlock()
					return
				}
			}
			if algoritmo == "SJF" && CpuLibres {
				// Recorremos la cola de Ready
				for _, proceso := range ColaReady {
					// Si el proceso no tiene una estimación de ráfaga calculada, la calculamos
					if proceso.MetricasDeEstados[globals.Exec] != 0 && !proceso.EstimacionDeRafaga.YaCalculado {
						//Est(n+1) =  alfa.R(n) + (1-alfa).Est(n) ;   alfa pertenece [0,1]
						proceso.EstimacionDeRafaga.TiempoDeRafaga = (Config_Kernel.Alpha * float64(proceso.TiempoDeUltimaRafaga.Milliseconds())) + (1-Config_Kernel.Alpha)*proceso.EstimacionDeRafaga.TiempoDeRafaga
						proceso.EstimacionDeRafaga.YaCalculado = true
					}
				}

				// Una vez que calculamos las estimaciones de ráfaga, elegimos el proceso con la estimación más pequeña
				pcbElegido := elegirPcbConEstimacionMasChica()

				// Cambiamos el boolean de YaCalculado a false para que se vuelva a calcular en la próxima iteración
				pcbElegido.EstimacionDeRafaga.YaCalculado = false

				// Movemos el proceso elegido a la cola Exec

				MoverProcesoACola(pcbElegido, &ColaExec)

				ultimo := ColaExec[len(ColaExec)-1]
				if !ElegirCpuYMandarProceso(ultimo) {
					// No se pudo enviar el proceso a la CPU, lo devolvemos a la cola Ready
					MoverProcesoACola(ultimo, &ColaReady)
					CpuLibres = false // Indicamos que la CPU no está libre
					return
				}
			}
			if algoritmo == "SRT" {
				//* Lógica para SJF con desalojo /SRT
				// Recorremos la cola de Ready
				if CpuLibres {
					for _, proceso := range ColaReady {
						// Si el proceso no tiene una estimación de ráfaga calculada, la calculamos
						if proceso.MetricasDeEstados[globals.Exec] != 0 && !proceso.EstimacionDeRafaga.YaCalculado {
							//Est(n+1) =  alfa.R(n) + (1-alfa).Est(n) ;   alfa pertenece [0,1]
							proceso.EstimacionDeRafaga.TiempoDeRafaga = (Config_Kernel.Alpha * float64(proceso.TiempoDeUltimaRafaga.Milliseconds())) + (1-Config_Kernel.Alpha)*proceso.EstimacionDeRafaga.TiempoDeRafaga
							proceso.EstimacionDeRafaga.YaCalculado = true
						}
					}

					// Una vez que calculamos las estimaciones de ráfaga, elegimos el proceso con la estimación más pequeña
					pcbElegido := elegirPcbConEstimacionMasChica()

					// Cambiamos el boolean de YaCalculado a false para que se vuelva a calcular en la próxima iteración
					pcbElegido.EstimacionDeRafaga.YaCalculado = false

					// Movemos el proceso elegido a la cola Exec

					MoverProcesoACola(pcbElegido, &ColaExec)

					ultimo := ColaExec[len(ColaExec)-1]
					if !ElegirCpuYMandarProceso(ultimo) {
						// No se pudo enviar el proceso a la CPU, lo devolvemos a la cola Ready
						MoverProcesoACola(ultimo, &ColaReady)
						MutexCpuLibres.Lock()
						CpuLibres = false // Indicamos que la CPU no está libre
						MutexCpuLibres.Unlock()
						return
					}
				} else {
					// Si no hay cpu libres, elegir a victima de SRT
					//! ANALIZAR TIEMPO RESTANTE DE CADA CPU
					// Si el proceso nuevo de ready (el ultimo) tiene una estimacion menor a los de exec
					// Desalojar al que mas tiempo le quede
				}
			}
		}
		planificadorCortoMutex.Unlock()
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

func MoverProcesoDeExecABlocked(pid int) {

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

		// Busco el PCB del proceso en la cola de SuspBlocked
		pcbDelProceso := BuscarProcesoEnCola(pid, &ColaSuspBlocked)

		//! SACAR EL PROCESO DE SWAP

		//Lo muevo a la cola destino
		MoverProcesoACola(*pcbDelProceso, &ColaSuspReady)
		LargoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
	} else {
		// Como lo encontré en la cola de blocked, lo muevo a la cola destino
		MoverProcesoACola(*pcbDelProceso, &ColaReady)
		CortoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
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
