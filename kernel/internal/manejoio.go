package kernel_internal

import (
	"fmt"
	"globals"
	"sync"
)

type DispositivoIO struct {
	InstanciasDispositivo []InstanciaIO
	ColaEsperaProcesos    []ProcesoEsperando
}

type ProcesoEsperando struct {
	Proceso globals.PCB
	Tiempo  int
}

type InstanciaIO struct {
	NombreIO string
	IpIO     string
	PortIO   int
	Estado   string
	PID      int // PID del proceso que está usando la instancia, -1 si está libre
}

var ListaDispositivosIO map[string]*DispositivoIO = make(map[string]*DispositivoIO)

var MutexIdentificadoresIO sync.Mutex

var mutexProcesoEsperando sync.Mutex

// &-------------------------------------------Funciones de Administración de IO-------------------------------------------------------------

func RegistrarInstanciaIO(nombre string, puerto int, ip string) {
	instancia := InstanciaIO{NombreIO: nombre, IpIO: ip, PortIO: puerto, Estado: "Libre", PID: -1}

	// Si ya existe el dispositivo IO, agregamos la instancia
	if disp, ok := ListaDispositivosIO[nombre]; ok {
		disp.InstanciasDispositivo = append(disp.InstanciasDispositivo, instancia)

		//Al tener una nueva instancia libre, si hay un proceso en espera, la ocupamos
		if len(disp.ColaEsperaProcesos) != 0 {

			MutexIdentificadoresIO.Lock()
			mutexProcesoEsperando.Lock()

			procesoEsperando := ListaDispositivosIO[nombre].ColaEsperaProcesos[0]
			// Lo eliminamos de la cola de espera
			ListaDispositivosIO[nombre].ColaEsperaProcesos = ListaDispositivosIO[nombre].ColaEsperaProcesos[1:] // Eliminamos el primer proceso de la cola de espera

			mutexProcesoEsperando.Unlock()
			MutexIdentificadoresIO.Unlock()

			UsarDispositivoDeIO(nombre, procesoEsperando.Proceso.PID, procesoEsperando.Tiempo)
		}

		//Si no existe el dispositivo IO, lo creamos
	} else {
		MutexIdentificadoresIO.Lock()
		ListaDispositivosIO[nombre] = &DispositivoIO{
			ColaEsperaProcesos:    []ProcesoEsperando{},
			InstanciasDispositivo: []InstanciaIO{instancia},
		}
		MutexIdentificadoresIO.Unlock()
	}

	mensajeRegistro := fmt.Sprintf("Dispositivo IO registrado: Nombre %s, IP %s, Puerto %d", nombre, ip, puerto)
	fmt.Println(mensajeRegistro)
	Logger.Debug(mensajeRegistro)

	mensajeNumeroInstancias := fmt.Sprintf("Número de instancias del dispositivo %s: %d", nombre, len(ListaDispositivosIO[nombre].InstanciasDispositivo))
	fmt.Println(mensajeNumeroInstancias)
	Logger.Debug(mensajeNumeroInstancias)

}

func DesconectarInstanciaIO(nombreDispositivo string, ipInstancia string, puertoInstancia int) {
	// Recorro la lista de instancias del dispositivo IO en búsqueda de la instancia a eliminar
	for pos, instanciaBuscada := range ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo {

		// Busco que coincida el IP y el puerto de la instancia
		if instanciaBuscada.IpIO == ipInstancia && instanciaBuscada.PortIO == puertoInstancia {

			mensajeDesconectandoInstancia := fmt.Sprintf("Desconectando instancia de IO Dispositivo %s, IP %s, Puerto %d", nombreDispositivo, ipInstancia, puertoInstancia)
			fmt.Println(mensajeDesconectandoInstancia)
			Logger.Debug(mensajeDesconectandoInstancia)

			// Mandamos al proceso que estaba usando la instancia a Exit
			if instanciaBuscada.Estado == "Ocupada" {
				mensajeDesconexionDuranteEjecucion := fmt.Sprintf("Desconexion de IO durante ejecucion, se envia proceso a Exit: PID %d, Dispositivo %s", instanciaBuscada.PID, nombreDispositivo)
				fmt.Println(mensajeDesconexionDuranteEjecucion)
				Logger.Debug(mensajeDesconexionDuranteEjecucion)

				MoverProcesoDeBlockedAExit(instanciaBuscada.PID)
			}

			// Borramos la instancia de IO de la lista de instancias del dispositivo IO
			MutexIdentificadoresIO.Lock()
			ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo = append(ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo[:pos], ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo[pos+1:]...)
			mensajeInstanciasRestantes := fmt.Sprintf("Instancias del dispositivo %s: %d", nombreDispositivo, len(ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo))
			MutexIdentificadoresIO.Unlock()

			fmt.Println(mensajeInstanciasRestantes)
			Logger.Debug(mensajeInstanciasRestantes)

			// Si la instancia desconectada era la única que quedaba...
			if len(ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo) == 0 {

				// Y además hay procesos en espera...
				if len(ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos) != 0 {

					// Recorro la lista de procesos bloqueados por la IO y los mando a Exit
					for i, procesoEnEspera := range ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos {
						mensajeDesconexionDuranteEjecucion := fmt.Sprintf("Desconexion de IO durante ejecucion, se envia proceso esperando al dispositivo a Exit: PID %d, Dispositivo %s", procesoEnEspera.Proceso.PID, nombreDispositivo)
						fmt.Println(mensajeDesconexionDuranteEjecucion)
						Logger.Debug(mensajeDesconexionDuranteEjecucion)

						MoverProcesoDeBlockedAExit(ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos[i].Proceso.PID)
					}
				}

				// Borramos del mapa de dispositivos IO el dispositivo que ya no tiene instancias
				MutexIdentificadoresIO.Lock()
				delete(ListaDispositivosIO, nombreDispositivo)
				MutexIdentificadoresIO.Unlock()

			} else {
				mensajeInstanciasRestantes := fmt.Sprintf("Instancias del dispositivo %s: %d", nombreDispositivo, len(ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo))
				fmt.Println(mensajeInstanciasRestantes)
				Logger.Debug(mensajeInstanciasRestantes)
			}

			mensajeInstanciaDesconectada := fmt.Sprintf("Instancia de IO desconectada: Dispositivo %s, IP %s, Puerto %d", nombreDispositivo, ipInstancia, puertoInstancia)
			fmt.Println(mensajeInstanciaDesconectada)
			Logger.Debug(mensajeInstanciaDesconectada)

		} else {
			mensajeInstanciaNoEncontrada := fmt.Sprintf("Instancia de IO no encontrada: Dispositivo %s, IP %s, Puerto %d", nombreDispositivo, ipInstancia, puertoInstancia)
			fmt.Println(mensajeInstanciaNoEncontrada)
			Logger.Debug(mensajeInstanciaNoEncontrada)
		}
	}

}

func VerificarDispositivo(ioName string) bool {
	// Verificar si el dispositivo existe en el mapa ListaDispositivosIO
	if _, existeDispositivo := ListaDispositivosIO[ioName]; existeDispositivo {
		return true
	}
	return false
}

func VerificarInstanciaDeIO(nombreDispositivo string) bool {
	for _, instancia := range ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo {
		if instancia.Estado == "Libre" {
			return true
		}
	}
	return false
}

func UsarDispositivoDeIO(nombreDispositivo string, pid int, milisegundosDeUso int) {
	// Buscamos una instancia de IO libre para el proceso
	instanciaDeIOLibre, estaLibre := BuscarPrimerInstanciaLibre(nombreDispositivo)

	if !estaLibre {
		Logger.Debug("No hay instancias libres del dispositivo IO", "nombre", nombreDispositivo)
		BloquearProcesoPorIO(nombreDispositivo, pid, milisegundosDeUso)
		return
	}

	// Ocupamos la instancia de IO libre con el PID del proceso
	OcuparInstanciaDeIO(nombreDispositivo, instanciaDeIOLibre, pid)
	// Enviamos el proceso a la instancia de IO
	EnviarProcesoAIO(instanciaDeIOLibre, pid, milisegundosDeUso)
}

func ProcesarFinIO(pid int, nombreDispositivo string) {

	// Buscamos en Blocked / SuspBlocked el proceso que terminó su IO y lo mandamos a Ready
	MoverProcesoDeBlockedAReady(pid)

	// Buscamos la instancia de IO que estaba ocupada por el PID
	instanciaDeIO := BuscarInstanciaDeIOporPID(nombreDispositivo, pid)

	// Liberamos la instancia de IO que estaba ocupada por el PID
	LiberarInstanciaDeIO(nombreDispositivo, *instanciaDeIO)

	if len(ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos) != 0 {
		// Si hay procesos esperando en la cola de espera del dispositivo, ocupamos la instancia recientemente liberada

		MutexIdentificadoresIO.Lock()
		mutexProcesoEsperando.Lock()
		procesoEsperando := ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos[0]
		// Lo eliminamos de la cola de espera
		ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos = ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos[1:] // Eliminamos el primer proceso de la cola de espera

		mutexProcesoEsperando.Unlock()
		MutexIdentificadoresIO.Unlock()

		// Ocupamos la instancia de IO con el PID del proceso que estaba esperando
		UsarDispositivoDeIO(nombreDispositivo, procesoEsperando.Proceso.PID, procesoEsperando.Tiempo)
	}

	LogFinDeIO(pid)

}

func OcuparInstanciaDeIO(nombreDispositivo string, instancia InstanciaIO, pid int) {

	for i, instanciaIO := range ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo {

		if instanciaIO == instancia {
			MutexIdentificadoresIO.Lock()
			instancia := &ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo[i]
			instancia.Estado = "Ocupada"
			instancia.PID = pid

			mensajeOcupacionInstancia := fmt.Sprintf("Instancia de IO ocupada: Dispositivo %s, Estado: %s, PID %d", nombreDispositivo, instancia.Estado, pid)
			MutexIdentificadoresIO.Unlock()

			fmt.Println(mensajeOcupacionInstancia)
			Logger.Debug(mensajeOcupacionInstancia)

			return
		}
	}

	mensajeErrorOcupacion := fmt.Sprintf("Error al ocupar la instancia de IO: Dispositivo %s, PID %d", instancia.NombreIO, pid)
	fmt.Println(mensajeErrorOcupacion)
	Logger.Debug(mensajeErrorOcupacion)
}

func LiberarInstanciaDeIO(nombreDispositivo string, instancia InstanciaIO) {
	for i, instanciaIO := range ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo {
		if instanciaIO == instancia {
			MutexIdentificadoresIO.Lock()
			ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo[i].Estado = "Libre"
			ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo[i].PID = -1
			Logger.Debug("Instancia de IO liberada", "dispositivo", instancia.NombreIO)
			MutexIdentificadoresIO.Unlock()
			return
		}
	}
	Logger.Debug("Error al liberar la instancia de IO", "dispositivo", instancia.NombreIO)
}

func BuscarPrimerInstanciaLibre(nombreDispositivo string) (InstanciaIO, bool) {
	for _, instancia := range ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo {
		if instancia.Estado == "Libre" {
			mensajeInstanciaLibre := fmt.Sprintf("Instancia de IO libre encontrada: Dispositivo %s, IP %s, Puerto %d", instancia.NombreIO, instancia.IpIO, instancia.PortIO)
			fmt.Println(mensajeInstanciaLibre)
			Logger.Debug(mensajeInstanciaLibre)

			return instancia, true
		}
	}

	mensajeNoHayInstancias := fmt.Sprintf("No hay instancias libres del dispositivo IO: %s", nombreDispositivo)
	fmt.Println(mensajeNoHayInstancias)
	Logger.Debug(mensajeNoHayInstancias)

	return InstanciaIO{}, false
}

func BloquearProcesoPorIO(nombreDispositivo string, pid int, tiempoEspera int) {

	//Buscamos el PCB del proceso en la cola de blocked
	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaBlocked)

	ProcesoEsperando := ProcesoEsperando{
		Proceso: *pcbDelProceso,
		Tiempo:  tiempoEspera,
	}
	//Agregar el proceso a la cola de espera del dispositivo
	MutexIdentificadoresIO.Lock()
	ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos = append(ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos, ProcesoEsperando)
	MutexIdentificadoresIO.Unlock()

}

func BuscarInstanciaDeIOporPID(nombreDispositivo string, pid int) *InstanciaIO {
	for _, instancia := range ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo {
		if instancia.PID == pid {
			return &instancia
		}
	}
	Logger.Debug("No se encontró la instancia de IO para el PID", "nombre", nombreDispositivo, "pid", pid)
	return nil
}
