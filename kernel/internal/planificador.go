package kernel_internal

import (
	"bufio"
	"globals"
	"os"
	"sync"
	"time"
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
				if !PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory) {
					Logger.Debug("No se puede conectar con memoria (Ping no devuelto)")
					break // Salimos del for para esperar un nuevo proceso en New
				}

				//! Estaria piola hacer una funcion que resuma todo esto, pero por ahora lo dejo asi
				//Verifico si hay procesos en la cola SuspReady
				if len(ColaSuspReady) != 0 {

					// Pido espacio en memoria para el primer proceso de la cola SuspReady
					respuestaMemoria := PedirEspacioAMemoria(ColaSuspReady[0])

					// Si memoria responde que no hay espacio..
					if !respuestaMemoria {
						hayEspacioEnMemoria = false // Seteo la variable del for a false
						break                       // Salimos del for para esperar un nuevo proceso en New
					}

					MoverProcesoACola(ColaSuspReady[0], &ColaReady)
					CortoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
				} else { //Si no hay procesos en SuspReady, ya se que hay en New

					// Pido espacio en memoria para el primer proceso de la cola New
					respuestaMemoria := PedirEspacioAMemoria(ColaNew[0])

					// Si memoria responde que no hay espacio...
					if !respuestaMemoria {
						hayEspacioEnMemoria = false // Seteo la variable del for a false
						break                       // Salimos del for para esperar un nuevo proceso en New
					}

					MoverProcesoACola(ColaNew[0], &ColaReady)
					CortoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
				}

			}

			if algoritmo == "PMCP" {
				// Si memoria responde...
				if !PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory) {
					Logger.Debug("No se puede conectar con memoria (Ping no devuelto)")
					break // Salimos del for para esperar un nuevo proceso en New
				}

				//Verifico si hay procesos en la cola SuspReady
				if len(ColaSuspReady) != 0 {
					// Pido espacio en memoria para el primer proceso de la cola New
					respuestaMemoria := PedirEspacioAMemoria(ColaSuspReady[pcbMasChico()])

					// Si memoria responde que no hay espacio...
					if !respuestaMemoria {
						hayEspacioEnMemoria = false // Seteo la variable del for a false
						break                       // Salimos del for para esperar un nuevo proceso en New
					}

					MoverProcesoACola(ColaSuspReady[pcbMasChico()], &ColaReady)
					CortoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
				} else {
					// Pido espacio en memoria para el primer proceso de la cola New
					respuestaMemoria := PedirEspacioAMemoria(ColaNew[pcbMasChico()])

					// Si memoria responde que no hay espacio...
					if !respuestaMemoria {
						hayEspacioEnMemoria = false // Seteo la variable del for a false
						break                       // Salimos del for para esperar un nuevo proceso en New
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

				// Sabemos que el proceso que acabamos de mover a Exec es el último de la cola
				ultimoProcesoEnReady := &ColaExec[len(ColaExec)-1]

				// Intentamos enviar el proceso a la CPU
				if ElegirCpuYMandarProceso(*ultimoProcesoEnReady) {
					// No se pudo enviar el proceso a la CPU, lo devolvemos a la cola Ready
					MoverProcesoACola(*ultimoProcesoEnReady, &ColaExec)
					break // Salimos del for para esperar un nuevo proceso en Ready
				} else {
					MutexCpuLibres.Lock()
					CpuLibres = false // Indicamos que la CPU no está libre
					MutexCpuLibres.Unlock()
				}
			}
			if algoritmo == "SJF" && CpuLibres {
				// Recorremos la cola de Ready
				for i := range ColaReady {
					// Si el proceso no tiene una estimación de ráfaga calculada, la calculamos
					if ColaReady[i].MetricasDeEstados[globals.Exec] != 0 && !ColaReady[i].EstimacionDeRafaga.YaCalculado {
						//Est(n+1) =  alfa.R(n) + (1-alfa).Est(n) ;   alfa pertenece [0,1]
						ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga = (Config_Kernel.Alpha * float64(ColaReady[i].TiempoDeUltimaRafaga.Milliseconds())) +
							(1-Config_Kernel.Alpha)*ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga
						ColaReady[i].EstimacionDeRafaga.YaCalculado = true
					}
				}

				// Una vez que calculamos las estimaciones de ráfaga, elegimos el proceso con la estimación más pequeña
				pcbElegido := elegirPcbConEstimacionMasChica()

				// Cambiamos el boolean de YaCalculado a false para que se vuelva a calcular en la próxima iteración
				pcbElegido.EstimacionDeRafaga.YaCalculado = false

				if ElegirCpuYMandarProceso(*pcbElegido) {
					// Movemos el proceso elegido a la cola Exec
					MoverProcesoACola(*pcbElegido, &ColaExec)
					break // Salimos del for para esperar un nuevo proceso en Ready
				} else {
					// No se pudo enviar el proceso a la CPU porque no habia CPUs libres
					MutexCpuLibres.Lock()
					CpuLibres = false // Indicamos que la CPU no está libre
					MutexCpuLibres.Unlock()
				}
			}
			if algoritmo == "SRT" {
				// Recorremos la cola de Ready
				for i := range ColaReady {
					// Si el proceso no tiene una estimación de ráfaga calculada, la calculamos
					if ColaReady[i].MetricasDeEstados[globals.Exec] != 0 && !ColaReady[i].EstimacionDeRafaga.YaCalculado {
						//Est(n+1) =  alfa.R(n) + (1-alfa).Est(n) ;   alfa pertenece [0,1]
						ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga = (Config_Kernel.Alpha * float64(ColaReady[i].TiempoDeUltimaRafaga.Milliseconds())) +
							(1-Config_Kernel.Alpha)*ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga
						ColaReady[i].EstimacionDeRafaga.YaCalculado = true
					}
				}

				// Una vez que calculamos las estimaciones de ráfaga, elegimos el proceso con la estimación más pequeña
				pcbElegido := elegirPcbConEstimacionMasChica()

				// Cambiamos el boolean de YaCalculado a false para que se vuelva a calcular en la próxima iteración
				pcbElegido.EstimacionDeRafaga.YaCalculado = false

				if ElegirCpuYMandarProceso(*pcbElegido) {
					// Movemos el proceso elegido a la cola Exec
					MoverProcesoACola(*pcbElegido, &ColaExec)
					break // Salimos del for para esperar un nuevo proceso en Ready
				} else {
					// No se pudo enviar el proceso a la CPU porque no habia CPUs libres
					MutexCpuLibres.Lock()
					CpuLibres = false // Indicamos que la CPU no está libre
					MutexCpuLibres.Unlock()
				}
			} else {
				// Si no hay cpu libres, elegir a victima de SRT, el que tenga mayor tiempo restante en la CPU
				pcbVictima := buscarTiempoRestanteEnCpuMasAlto()
				// Agarro el proceso que generó la comparación
				ultimoProcesoEnReady := &ColaReady[len(ColaReady)-1]

				if ultimoProcesoEnReady.EstimacionDeRafaga.TiempoDeRafaga < tiempoRestanteEnCpu(*pcbVictima) {
					// Pido el desalojo a la CPU del proceso víctima
					PeticionDesalojo(pcbVictima.PID, "Planificador")

					// Muevo el proceso víctima a la cola Ready
					MoverProcesoACola(*pcbVictima, &ColaReady)

					// Intentamos enviar el proceso a la CPU
					if ElegirCpuYMandarProceso(*ultimoProcesoEnReady) {

						// Muevo el proceso que llegó de Ready a la cola Exec
						MoverProcesoACola(*ultimoProcesoEnReady, &ColaExec)

						break // Salimos del for para esperar un nuevo proceso en Ready

					} else {
						// No se pudo enviar el proceso a la CPU porque no habia CPUs libres
						MutexCpuLibres.Lock()
						CpuLibres = false // Indicamos que la CPU no está libre
						MutexCpuLibres.Unlock()
					}

				}
			}
		}
		planificadorCortoMutex.Unlock()
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

func elegirPcbConEstimacionMasChica() *globals.PCB {

	if len(ColaReady) == 0 {
		return nil
	}
	minIdx := 0
	for i := 1; i < len(ColaReady); i++ {
		if ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga < ColaReady[minIdx].EstimacionDeRafaga.TiempoDeRafaga {
			minIdx = i
		}
	}
	return &ColaReady[minIdx]

}

func buscarTiempoRestanteEnCpuMasAlto() *globals.PCB {
	if len(ColaExec) == 0 {
		return nil
	}
	maxIdx := 0
	for i := 1; i < len(ColaExec); i++ {
		if tiempoRestanteEnCpu(ColaExec[i]) > tiempoRestanteEnCpu(ColaExec[maxIdx]) {
			maxIdx = i
		}
	}
	return &ColaExec[maxIdx]
}

func tiempoRestanteEnCpu(pcb globals.PCB) float64 {
	//& Esta función calcula el tiempo restante en la CPU para un PCB dado.

	return pcb.EstimacionDeRafaga.TiempoDeRafaga - float64(time.Since(pcb.InicioEstadoActual).Milliseconds())

}
