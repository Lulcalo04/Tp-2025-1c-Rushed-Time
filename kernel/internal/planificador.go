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

var MutexCpuLibres sync.Mutex
var CpuLibres bool = true

var mejorTiempoDeRafaga float64 = 1.7976931348623157e+30 // valor máximo para float64
var posicionDeLaMejorRafaga = -1

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
var AlgoritmoLargoPlazo string

func PlanificadorLargoPlazo() {
	AlgoritmoLargoPlazo = Config_Kernel.ReadyIngressAlgorithm

	fmt.Println("Planificador de largo plazo iniciado (algoritmo:", AlgoritmoLargoPlazo, ")")
	Logger.Debug("Planificador de largo plazo iniciado", "algoritmo", AlgoritmoLargoPlazo)

	for {
		time.Sleep(1000 * time.Millisecond) // Espera 1000ms antes de volver a ejecutar el planificador
		<-LargoNotifier
		MutexPlanificadorLargo.Lock()

		fmt.Println("P.LP: Planificador de largo plazo ejecutando y entro al mutex, Procesos en New:", len(ColaNew), "Procesos en SuspReady:", len(ColaSuspReady))
		Logger.Debug("Planificador de largo plazo ejecutando y entro al mutex", "ProcesosEnNew", len(ColaNew), "ProcesosEnSuspReady", len(ColaSuspReady))
		// Mientras hay espacio en memoria y procesos en las colas

		for len(ColaNew) != 0 || len(ColaSuspReady) != 0 {

			if AlgoritmoLargoPlazo == "FIFO" {
				Logger.Debug("P.LP: Procesando cola New")
				fmt.Println("P.LP: Procesando cola New")
				// Verifica conexión con memoria
				if !PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory) {
					Logger.Debug("No se puede conectar con memoria (Ping no devuelto)")
					fmt.Println("No se puede conectar con memoria (Ping no devuelto)")
					break // Salimos del for para esperar un nuevo proceso en New
				}

				// Procesa las colas SuspReady o New
				if len(ColaSuspReady) != 0 {

					// Pedimos la liberación de swap del primer proceso de la cola SuspReady
					respuestaMemoria := PedirLiberacionDeSwap(ColaSuspReady[0].PID)

					if !respuestaMemoria {
						fmt.Println("P.LP: No hay espacio en memoria, rompiendo el for")
						Logger.Debug("No hay espacio en memoria, rompiendo el for")
						break // Salimos del for para esperar un nuevo proceso en New
					}

					// Mueve el proceso a la cola Ready y notifica al planificador de corto plazo
					MoverProcesoACola(ColaSuspReady[0], &ColaReady)
					// En vez de CortoNotifier <- struct{}{}, usamos el patrón select para evitar señales acumuladas
					select {
					case CortoNotifier <- struct{}{}:
						// señal enviada
					default:
						// ya hay una señal pendiente, no enviar otra
					}
				} else {
					// Guardamos referencia al proceso antes de moverlo
					procesoAProcesar := ColaNew[0]

					respuestaMemoria := PedirEspacioAMemoria(*procesoAProcesar)

					if !respuestaMemoria {
						fmt.Println("P.LP: No hay espacio en memoria, rompiendo el for")
						Logger.Debug("No hay espacio en memoria, rompiendo el for")
						break
					}

					// Mueve el proceso a la cola Ready y notifica al planificador de corto plazo
					MoverProcesoACola(procesoAProcesar, &ColaReady)
					// En vez de CortoNotifier <- struct{}{}, usamos el patrón select para evitar señales acumuladas
					select {
					case CortoNotifier <- struct{}{}:
						// señal enviada
					default:
						// ya hay una señal pendiente, no enviar otra
					}
				}
			}

			if AlgoritmoLargoPlazo == "PMCP" {
				// Si memoria responde...
				if !PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory) {
					Logger.Debug("No se puede conectar con memoria (Ping no devuelto)")
					break // Salimos del for para esperar un nuevo proceso en New
				}

				//Verifico si hay procesos en la cola SuspReady
				if len(ColaSuspReady) != 0 {

					// Pido la liberación de swap del proceso mas chico de la cola SuspReady
					respuestaMemoria := PedirLiberacionDeSwap(ColaSuspReady[pcbMasChico()].PID)

					if !respuestaMemoria {
						fmt.Println("P.LP: No hay espacio en memoria, rompiendo el for")
						Logger.Debug("No hay espacio en memoria, rompiendo el for")
						break // Salimos del for para esperar un nuevo proceso en New
					}

					MoverProcesoACola(ColaSuspReady[pcbMasChico()], &ColaReady)
					// En vez de CortoNotifier <- struct{}{}, usamos el patrón select para evitar señales acumuladas
					select {
					case CortoNotifier <- struct{}{}:
						// señal enviada
					default:
						// ya hay una señal pendiente, no enviar otra
					} // Notifico que hay un proceso listo para ejecutar
				} else {
					// Pido espacio en memoria para el primer proceso de la cola New

					fmt.Println("P.LP: Procesando cola New")
					Logger.Debug("Procesando cola New")

					// Guardamos el índice y referencia al proceso antes de procesarlo
					indiceProceso := pcbMasChico()
					procesoAProcesar := ColaNew[indiceProceso]

					respuestaMemoria := PedirEspacioAMemoria(*procesoAProcesar)

					// Si memoria responde que no hay espacio...
					if !respuestaMemoria {
						fmt.Println("P.LP: No hay espacio en memoria, rompiendo el for")
						Logger.Debug("No hay espacio en memoria, rompiendo el for")
						break // Salimos del for para esperar un nuevo proceso en New
					}

					MoverProcesoACola(procesoAProcesar, &ColaReady)
					// En vez de CortoNotifier <- struct{}{}, usamos el patrón select para evitar señales acumuladas
					select {
					case CortoNotifier <- struct{}{}:
						// señal enviada
					default:
						// ya hay una señal pendiente, no enviar otra
					} // Notifico que hay un proceso listo para ejecutar
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

var MutexPlanificadorCorto sync.Mutex // Mutex para evitar doble planificador de corto plazo
var CortoNotifier = make(chan struct{}, 999)
var AlgoritmoCortoPlazo string

func PlanificadorCortoPlazo() {
	AlgoritmoCortoPlazo = Config_Kernel.SchedulerAlgorithm

	fmt.Println("Planificador de corto plazo iniciado (algoritmo:", AlgoritmoCortoPlazo, ")")
	Logger.Debug("Planificador de corto plazo iniciado", "algoritmo", AlgoritmoCortoPlazo)

	for {
		<-CortoNotifier
		MutexPlanificadorCorto.Lock()

		for len(ColaReady) != 0 {
			if AlgoritmoCortoPlazo == "FIFO" && CpuLibres {

				Logger.Debug("P.CP: ENTRANDO EN FIFO")
				fmt.Println("P.CP: ENTRANDO EN FIFO")

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
			if AlgoritmoCortoPlazo == "SJF" && CpuLibres {
				// Una vez que calculamos las estimaciones de ráfaga, elegimos el proceso con la estimación más pequeña
				pcbElegido := elegirPcbConEstimacionMasChica()

				Logger.Debug("Proceso elegido ", "PID", pcbElegido.PID, "con estimación de ráfaga:", pcbElegido.EstimacionDeRafaga.TiempoDeRafaga)
				fmt.Println("Proceso elegido ", "PID", pcbElegido.PID, "con estimación de ráfaga:", pcbElegido.EstimacionDeRafaga.TiempoDeRafaga)

				ElegirCpuYMandarProceso(*pcbElegido)
				break // Salimos del for para esperar un nuevo proceso en Ready
			}
			if AlgoritmoCortoPlazo == "SRT" && CpuLibres {

				Logger.Debug("P.CP: Hay CPU libre, buscando proceso con menor estimación de ráfaga")
				fmt.Println("P.CP: Hay CPU libre, buscando proceso con menor estimación de ráfaga")

				// Una vez que calculamos las estimaciones de ráfaga, elegimos el proceso con la estimación más pequeña
				pcbElegido := elegirPcbConEstimacionMasChica()

				if !ElegirCpuYMandarProceso(*pcbElegido) {
					continue
				}
				break
			} else if AlgoritmoCortoPlazo == "SRT" && !CpuLibres {
				// Si no hay cpu libres, elegir a victima de SRT, el que tenga mayor tiempo restante en la CPU
				Logger.Debug("P.CP: No hay CPU libre, buscando víctima de SRT")
				fmt.Println("P.CP: No hay CPU libre, buscando víctima de SRT")

				MutexReady.Lock()

				// Verificar que ColaReady no esté vacía antes de calcular PosicionAMandar
				if len(ColaReady) == 0 {
					Logger.Debug("P.CP: ColaReady está vacía durante SRT")
					fmt.Println("P.CP: ColaReady está vacía durante SRT")
					MutexReady.Unlock()
					break
				}

				PosicionAMandar := len(ColaReady) - 1 // Posición del último proceso en Ready

				if ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga >= mejorTiempoDeRafaga {
					Logger.Debug("P.CP: Mejor estimacion", "ME", mejorTiempoDeRafaga, "PID", ColaReady[PosicionAMandar].PID, "Estimación de ráfaga", ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga)
					fmt.Println("P.CP: Mejor estimacion", "ME", mejorTiempoDeRafaga, "PID", ColaReady[PosicionAMandar].PID, "Estimación de ráfaga", ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga)

					if posicionDeLaMejorRafaga != -1 && posicionDeLaMejorRafaga < len(ColaReady) {
						PosicionAMandar = posicionDeLaMejorRafaga
					} else {
						Logger.Debug("P.CP: Mejor estimacion no establecida o índice inválido, cortando el ciclo")
						MutexReady.Unlock()
						break
					}
				}

				// Verificar nuevamente que el índice sea válido antes de usar PosicionAMandar
				if PosicionAMandar >= len(ColaReady) {
					Logger.Debug("P.CP: Índice PosicionAMandar fuera de rango", "indice", PosicionAMandar, "longitud", len(ColaReady))
					fmt.Println("P.CP: Índice PosicionAMandar fuera de rango", "indice", PosicionAMandar, "longitud", len(ColaReady))
					MutexReady.Unlock()
					break
				}

				// Buscamos el proceso que este ejecutando con mayor tiempo restante en CPU
				pcbVictima, tiempoRestanteVictima := buscarTiempoRestanteEnCpuMasAlto()

				if pcbVictima == nil {
					Logger.Debug("No hay procesos en la cola Exec para comparar con SRT")
					fmt.Println("No hay procesos en la cola Exec para comparar con SRT")
					MutexReady.Unlock()
					break
				}

				fmt.Println("P.CP: Buscando víctima de SRT, PID:", pcbVictima.PID, "con tiempo restante en CPU:", tiempoRestanteVictima, "y estimación de ráfaga del último en Ready:", ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga)
				Logger.Debug("P.CP: Buscando víctima de SRT", "PID", pcbVictima.PID, "TiempoRestanteEnCPU", tiempoRestanteVictima, "Estimación de ráfaga del último en Ready", ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga)

				if tiempoRestanteVictima > ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga && pcbVictima.PID != ColaReady[PosicionAMandar].PID {

					Logger.Debug("Quiero desalojar a la victima de SRT", "PID", pcbVictima.PID, "TiempoRestanteEnCPU", tiempoRestanteVictima, "Estimación de ráfaga del último en Ready", ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga)
					fmt.Println("Quiero desalojar a la victima de SRT", "PID", pcbVictima.PID, "TiempoRestanteEnCPU", tiempoRestanteVictima, "Estimación de ráfaga del último en Ready", ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga)

					// Guardamos una referencia al proceso antes de liberar el mutex
					if PosicionAMandar >= len(ColaReady) {
						Logger.Debug("P.CP: Índice inválido antes de enviar a CPU", "indice", PosicionAMandar, "longitud", len(ColaReady))
						MutexReady.Unlock()
						break
					}
					procesoAEnviar := ColaReady[PosicionAMandar]

					MutexCpuLiberada.Lock()
					CpuLiberada = false
					MutexCpuLiberada.Unlock()

					if PeticionDesalojo(pcbVictima.PID, "Planificador") {
						fmt.Println("P.CP: Desalojo exitoso de la víctima de SRT", "PID", pcbVictima.PID)

						//Le damos el lock de la cola READY para que pueda mover el proceso victima
						MutexReady.Unlock()

						MoverProcesoACola(pcbVictima, &ColaReady)
						fmt.Println("P.CP: Proceso víctima de SRT movido a Ready", "PID", pcbVictima.PID)

						for {
							if CpuLiberada {
								break
							}
						}

						// Usamos la referencia guardada en lugar del índice
						if ElegirCpuYMandarProceso(*procesoAEnviar) {
							//Si se pudo mandar el proceso a la CPU, actualizamos las variables de mejor estimación
							MutexReady.Lock()
							var PosibleMejor float64 = 999999999999999
							var posicionDelMejor int
							for i := range ColaReady {
								if ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga < PosibleMejor {
									PosibleMejor = ColaReady[i].EstimacionDeRafaga.TiempoDeRafaga
									posicionDelMejor = i
								}
							}
							mejorTiempoDeRafaga = PosibleMejor
							posicionDeLaMejorRafaga = posicionDelMejor
							MutexReady.Unlock()
						}
						break

					} else {
						Logger.Debug("No se pudo desalojar a la víctima de SRT", "PID", pcbVictima.PID)
						fmt.Println("No se pudo desalojar a la víctima de SRT", "PID", pcbVictima.PID)
						mejorTiempoDeRafaga = ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga
						posicionDeLaMejorRafaga = PosicionAMandar
						MutexReady.Unlock()
						break
					}
				} else {
					mejorTiempoDeRafaga = ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga
					posicionDeLaMejorRafaga = PosicionAMandar
					fmt.Println("P.CP: No se encontró una víctima de SRT con mejor estimación de ráfaga, manteniendo el último en Ready", "PID", ColaReady[PosicionAMandar].PID, "Estimación de ráfaga", ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga)
					Logger.Debug("P.CP: No se encontró una víctima de SRT con mejor estimación de ráfaga", "PID", ColaReady[PosicionAMandar].PID, "Estimación de ráfaga", ColaReady[PosicionAMandar].EstimacionDeRafaga.TiempoDeRafaga)
					MutexReady.Unlock()
				}
				break
			}
		}
		MutexPlanificadorCorto.Unlock()
		// Drenar señales adicionales en el canal para evitar interrupciones
		select {
		case <-CortoNotifier:
			Logger.Debug("P.CP: señal recibida, acumulada para una proxima iteración")
			fmt.Println("P.CP: señal recibida, acumulada para una proxima iteración")
		default:
			// No hay más señales, continuar normalmente
		}

	}
}

func pcbMasChico() int {
	// Encontrar el PCB más chico
	minIndex := 0
	for i := 1; i < len(ColaNew); i++ {
		if ColaNew[i].TamanioEnMemoria < ColaNew[minIndex].TamanioEnMemoria {
			minIndex = i
		}
	}
	return minIndex
}

func reestimarProceso(pcb *globals.PCB, motivoDesalojo string) {
	switch motivoDesalojo {
	case "Planificador":
		// Cuando desalojo un proceso la estimación de ráfaga se actualiza siendo la anterior menos lo que ejecutó
		tiempoRestante := pcb.EstimacionDeRafaga.TiempoDeRafaga - pcb.TiempoDeUltimaRafaga
		pcb.EstimacionDeRafaga.TiempoDeRafaga = tiempoRestante

		pcb.EstimacionDeRafaga.YaCalculado = true // Marcar como ya calculado

		Logger.Debug("Proceso reestimado, desalojo por planificador", "pid", pcb.PID, "nueva_estimacion", pcb.EstimacionDeRafaga.TiempoDeRafaga)
	default:
		// Caso en el que desalojo por IO o por DUMP (EXEC->BLOCKED/SUSP.READY->READY)

		if !pcb.EstimacionDeRafaga.YaCalculado {

			pcb.EstimacionDeRafaga.TiempoDeRafaga = (Config_Kernel.Alpha * pcb.TiempoDeUltimaRafaga) +
				(1-Config_Kernel.Alpha)*pcb.EstimacionDeRafaga.TiempoDeRafaga

			pcb.EstimacionDeRafaga.YaCalculado = true // Marcar como ya calculado
		}
		Logger.Debug("Proceso reestimado, desalojo por IO/DUMP", "pid", pcb.PID, "nueva_estimacion", pcb.EstimacionDeRafaga.TiempoDeRafaga)
	}

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

func buscarTiempoRestanteEnCpuMasAlto() (*globals.PCB, float64) {
	if len(ColaExec) == 0 {
		return nil, -1
	}

	maxIdx := 0
	maxTiempoRestante := tiempoRestanteEnCpu(*ColaExec[maxIdx])

	for i := 1; i < len(ColaExec); i++ {
		nuevoTiempoRestante := tiempoRestanteEnCpu(*ColaExec[i])
		if nuevoTiempoRestante > maxTiempoRestante {
			maxIdx = i
			maxTiempoRestante = nuevoTiempoRestante
		}
	}
	return ColaExec[maxIdx], maxTiempoRestante
}

func tiempoRestanteEnCpu(pcb globals.PCB) float64 {
	//& Esta función calcula el tiempo restante en la CPU para un PCB dado.

	return pcb.EstimacionDeRafaga.TiempoDeRafaga - float64(time.Since(pcb.InicioEstadoActual).Milliseconds())
}
