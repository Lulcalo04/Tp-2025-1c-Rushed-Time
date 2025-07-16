package kernel_internal

import (
	"bufio"
	"fmt"
	"globals"
	"os"
	"sync"
	"time"
)

var ColaNew = make([]*globals.PCB, 0)
var MutexNew sync.Mutex
var ColaReady = make([]*globals.PCB, 0)
var MutexReady sync.Mutex
var ColaExec = make([]*globals.PCB, 0)
var MutexExec sync.Mutex
var ColaBlocked = make([]*globals.PCB, 0)
var MutexBlocked sync.Mutex
var ColaSuspReady = make([]*globals.PCB, 0)
var MutexSuspReady sync.Mutex
var ColaSuspBlocked = make([]*globals.PCB, 0)
var MutexSuspBlocked sync.Mutex
var ColaExit = make([]*globals.PCB, 0)
var MutexExit sync.Mutex

var ColaEstados = map[*[]*globals.PCB]globals.Estado{
	&ColaNew:         globals.Estado("NEW"),
	&ColaReady:       globals.Estado("READY"),
	&ColaExec:        globals.Estado("EXEC"),
	&ColaBlocked:     globals.Estado("BLOCKED"),
	&ColaSuspReady:   globals.Estado("SUSP_READY"),
	&ColaSuspBlocked: globals.Estado("SUSP_BLOCKED"),
	&ColaExit:        globals.Estado("EXIT"),
}

var ColaMutexes = map[*[]*globals.PCB]*sync.Mutex{
	&ColaNew:         &MutexNew,
	&ColaReady:       &MutexReady,
	&ColaExec:        &MutexExec,
	&ColaBlocked:     &MutexBlocked,
	&ColaSuspReady:   &MutexSuspReady,
	&ColaSuspBlocked: &MutexSuspBlocked,
	&ColaExit:        &MutexExit,
}

var HayEspacioEnMemoria bool = true
var MutexHayEspacioEnMemoria sync.Mutex

var MutexCpuLibres sync.Mutex
var CpuLibres bool = true

func IniciarPlanificadores() {

	//ESPERAR ENTER
	Logger.Debug("Presione ENTER para iniciar los planificadores")
	fmt.Println("Presione ENTER para iniciar los planificadores")

	bufio.NewReader(os.Stdin).ReadBytes('\n')

	fmt.Println("Iniciando planificadores...")

	go PlanificadorLargoPlazo()
	go PlanificadorCortoPlazo()

	Logger.Debug("Planificadores iniciados")
	fmt.Println("Planificadores iniciados")
}

var MutexPlanificadorLargo sync.Mutex // Mutex para evitar doble planificador de largo plazo
var LargoNotifier = make(chan struct{}, 999)

func PlanificadorLargoPlazo() {
	algoritmo := Config_Kernel.SchedulerAlgorithm
	fmt.Println("Planificador de largo plazo iniciado (algoritmo:", algoritmo, ")")
	Logger.Debug("Planificador de largo plazo iniciado", "algoritmo", algoritmo)

	for {
		<-LargoNotifier
		MutexPlanificadorLargo.Lock()

		// Mientras hay espacio en memoria y procesos en las colas
		for HayEspacioEnMemoria && (len(ColaNew) != 0 || len(ColaSuspReady) != 0) {

			if algoritmo == "FIFO" {
				// Verifica conexión con memoria
				if !PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory) {
					Logger.Debug("No se puede conectar con memoria (Ping no devuelto)")
					fmt.Println("No se puede conectar con memoria (Ping no devuelto)")
					break // Salimos del for para esperar un nuevo proceso en New
				}

				// Procesa las colas SuspReady o New
				if len(ColaSuspReady) != 0 {
					if !procesarCola(ColaSuspReady) {
						break
					}
				} else {
					if !procesarCola(ColaNew) {
						break
					}
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
					respuestaMemoria := PedirEspacioAMemoria(*ColaSuspReady[pcbMasChico()])

					// Si memoria responde que no hay espacio...
					if !respuestaMemoria {
						MutexHayEspacioEnMemoria.Lock()
						HayEspacioEnMemoria = false // Seteo la variable del for a false
						MutexHayEspacioEnMemoria.Unlock()
						break // Salimos del for para esperar un nuevo proceso en New
					}

					MoverProcesoACola(ColaSuspReady[pcbMasChico()], &ColaReady)
					CortoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
				} else {
					// Pido espacio en memoria para el primer proceso de la cola New
					respuestaMemoria := PedirEspacioAMemoria(*ColaNew[pcbMasChico()])

					// Si memoria responde que no hay espacio...
					if !respuestaMemoria {
						MutexHayEspacioEnMemoria.Lock()
						HayEspacioEnMemoria = false // Seteo la variable del for a false
						MutexHayEspacioEnMemoria.Unlock()
						break // Salimos del for para esperar un nuevo proceso en New
					}

					MoverProcesoACola(ColaNew[pcbMasChico()], &ColaReady)
					CortoNotifier <- struct{}{} // Notifico que hay un proceso listo para ejecutar
				}

			}
		}
		MutexPlanificadorLargo.Unlock()
		// Drenar señales adicionales en el canal para evitar interrupciones
		select {
		case <-LargoNotifier:
			Logger.Debug("P.LP: señal recibida, acumulada para una proxima iteración")
			fmt.Println("P.LP: señal recibida, acumulada para una proxima iteración")
		default:
			// No hay más señales, continuar normalmente
		}
	}
}

func procesarCola(cola []*globals.PCB) bool {
	// Pide espacio en memoria para el primer proceso de la cola
	respuestaMemoria := PedirEspacioAMemoria(*cola[0])

	if !respuestaMemoria {
		MutexHayEspacioEnMemoria.Lock()
		HayEspacioEnMemoria = false
		MutexHayEspacioEnMemoria.Unlock()
		return false
	}

	// Mueve el proceso a la cola Ready y notifica al planificador de corto plazo
	MoverProcesoACola(cola[0], &ColaReady)
	CortoNotifier <- struct{}{}
	return true
}

var MutexPlanificadorCorto sync.Mutex // Mutex para evitar doble planificador de corto plazo
var CortoNotifier = make(chan struct{})

func PlanificadorCortoPlazo() {
	algoritmo := Config_Kernel.ReadyIngressAlgorithm

	fmt.Println("Planificador de corto plazo iniciado (algoritmo:", algoritmo, ")")
	Logger.Debug("Planificador de corto plazo iniciado", "algoritmo", algoritmo)

	for {
		<-CortoNotifier
		MutexPlanificadorCorto.Lock()

		for len(ColaReady) != 0 {
			if algoritmo == "FIFO" && CpuLibres {

				// Sabemos que el proceso que acabamos de mover a Exec es el último de la cola
				primerProcesoEnReady := ColaReady[0]

				// Intentamos enviar el proceso a la CPU
				if ElegirCpuYMandarProceso(*primerProcesoEnReady) {
					// No se pudo enviar el proceso a la CPU, lo devolvemos a la cola Ready
					Logger.Debug("P.CP: Encontramos CPU libre, moviendo proceso a Exec", "PID", primerProcesoEnReady.PID)
					fmt.Println("P.CP: Encontramos CPU libre, moviendo proceso a Exec", "PID", primerProcesoEnReady.PID)

					break // Salimos del for para esperar un nuevo proceso en Ready
				} else {
					Logger.Debug("P.CP: No se pudo enviar el proceso a la CPU porque no hay CPUs libres del", "PID", primerProcesoEnReady.PID)
					fmt.Println("P.CP: No se pudo enviar el proceso a la CPU porque no hay CPUs libres")

					// Indicamos que la CPU no está libre
					MutexCpuLibres.Lock()
					CpuLibres = false
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
						Logger.Debug("Estimacion de rafaga del ", "PID", ColaReady[i].PID, "calculada:", ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga)
						fmt.Println("Estimacion de rafaga del ", "PID", ColaReady[i].PID, "calculada:", ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga)
						// Marcamos que ya calculamos la estimación de ráfaga
						ColaReady[i].EstimacionDeRafaga.YaCalculado = true
					}
				}

				// Una vez que calculamos las estimaciones de ráfaga, elegimos el proceso con la estimación más pequeña
				pcbElegido := elegirPcbConEstimacionMasChica()

				Logger.Debug("Proceso elegido ", "PID", pcbElegido.PID, "con estimación de ráfaga:", pcbElegido.EstimacionDeRafaga.TiempoDeRafaga)
				fmt.Println("Proceso elegido ", "PID", pcbElegido.PID, "con estimación de ráfaga:", pcbElegido.EstimacionDeRafaga.TiempoDeRafaga)

				// Cambiamos el boolean de YaCalculado a false para que se vuelva a calcular en la próxima iteración
				pcbElegido.EstimacionDeRafaga.YaCalculado = false

				if ElegirCpuYMandarProceso(*pcbElegido) {
					break // Salimos del for para esperar un nuevo proceso en Ready
				} else {
					// No se pudo enviar el proceso a la CPU porque no habia CPUs libres
					MutexCpuLibres.Lock()
					CpuLibres = false // Indicamos que la CPU no está libre
					MutexCpuLibres.Unlock()
				}
			}
			if algoritmo == "SRT" && CpuLibres {
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
					break // Salimos del for para esperar un nuevo proceso en Ready
				} else {
					// No se pudo enviar el proceso a la CPU porque no habia CPUs libres
					MutexCpuLibres.Lock()
					CpuLibres = false // Indicamos que la CPU no está libre
					MutexCpuLibres.Unlock()
				}
			} else if algoritmo == "SRT" && !CpuLibres {
				// Si no hay cpu libres, elegir a victima de SRT, el que tenga mayor tiempo restante en la CPU
				pcbVictima := buscarTiempoRestanteEnCpuMasAlto()
				// Agarro el proceso que generó la comparación
				ultimoProcesoEnReady := ColaReady[len(ColaReady)-1]

				if ultimoProcesoEnReady.EstimacionDeRafaga.TiempoDeRafaga < tiempoRestanteEnCpu(*pcbVictima) {
					// Pido el desalojo a la CPU del proceso víctima
					PeticionDesalojo(pcbVictima.PID, "Planificador")

					// Muevo el proceso víctima a la cola Ready
					MoverProcesoACola(pcbVictima, &ColaReady)

					// Intentamos enviar el proceso a la CPU
					if ElegirCpuYMandarProceso(*ultimoProcesoEnReady) {
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
		MutexPlanificadorCorto.Unlock()
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

	Logger.Debug("## BUSCANDO PCB CON MENOR ESTIMACION DE RAFAGA ##")
	fmt.Println("## BUSCANDO PCB CON MENOR ESTIMACION DE RAFAGA ##")

	if len(ColaReady) == 0 {
		Logger.Debug("No hay procesos en la cola Ready")
		fmt.Println("No hay procesos en la cola Ready")
		return nil
	}
	minIdx := 0
	for i := 1; i < len(ColaReady); i++ {

		Logger.Debug("Estimacion de rafaga del ", "PID", ColaReady[i].PID, ":", ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga)
		fmt.Println("Estimacion de rafaga del ", "PID", ColaReady[i].PID, ":", ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga)

		if ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga < ColaReady[minIdx].EstimacionDeRafaga.TiempoDeRafaga {
			minIdx = i
		}
	}
	return ColaReady[minIdx]

}

func buscarTiempoRestanteEnCpuMasAlto() *globals.PCB {
	if len(ColaExec) == 0 {
		return nil
	}
	maxIdx := 0
	for i := 1; i < len(ColaExec); i++ {
		if tiempoRestanteEnCpu(*ColaExec[i]) > tiempoRestanteEnCpu(*ColaExec[maxIdx]) {
			maxIdx = i
		}
	}
	return ColaExec[maxIdx]
}

func tiempoRestanteEnCpu(pcb globals.PCB) float64 {
	//& Esta función calcula el tiempo restante en la CPU para un PCB dado.

	return pcb.EstimacionDeRafaga.TiempoDeRafaga - float64(time.Since(pcb.InicioEstadoActual).Milliseconds())

}
