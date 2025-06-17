package kernel_internal

import (
	"log/slog"
	"os"
	"strconv"
	"time"
	"utils/client"
	"utils/globals"
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

type IdentificadorCPU struct {
	CPUID   string
	Puerto  int
	Ip      string
	Ocupado bool
	PID     int
}

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

var ListaDispositivosIO map[string]*DispositivoIO

var ListaIdentificadoresCPU []IdentificadorCPU

var Config_Kernel *ConfigKernel

var Logger *slog.Logger

var ContadorPID int

var canceladoresBlocked = make(map[int]chan struct{})

// &-------------------------------------------Funciones de Kernel-------------------------------------------------------------

func IniciarKernel() {
	//Inicializa la config de kernel
	globals.IniciarConfiguracion("kernel/config.json", &Config_Kernel)

	//Crea el archivo donde se logea kernel
	Logger = globals.ConfigurarLogger("kernel", Config_Kernel.LogLevel)

	//Prende el server de kernel en un hilo aparte
	go IniciarServerKernel(Config_Kernel.PortKernel)

	//Realiza el handshake con memoria
	//client.HandshakeCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory, Logger)

	//Inicia los planificadores
	IniciarPlanificadores()
}

func InicializarProcesoCero() (string, int) {
	if len(os.Args) < 3 {
		Logger.Debug("Error, mal escrito usa: .kernel/kernel.go [archivo_pseudocodigo] [tamanio_proceso]")
		os.Exit(1)
	}

	// Leer el nombre del archivo de pseudocódigo
	nombreArchivoPseudocodigo := os.Args[1]

	// Leer y convertir el tamaño del proceso a entero
	tamanioProceso, err := strconv.Atoi(os.Args[2])
	if err != nil {
		Logger.Debug("Error: el tamaño del proceso debe ser un número entero.", "valor_recibido", os.Args[2])
		os.Exit(1)
	}

	// !VAMOS A TENER QUE METER EL PROCESO CERO EN LA COLA DE NEW
	return nombreArchivoPseudocodigo, tamanioProceso
}

func BuscarProcesoEnCola(pid int, cola *[]globals.PCB) *globals.PCB {
	for i, proceso := range *cola {
		if proceso.PID == pid {
			return &(*cola)[i]
		}
	}
	return nil
}

func InicializarPCB(tamanioEnMemoria int) {

	ContadorPID++

	pcb := globals.PCB{
		PID:               ContadorPID,
		PC:                0,
		Estado:            globals.New,
		MetricasDeEstados: make(map[globals.Estado]int),
		MetricasDeTiempos: make(map[globals.Estado]int),
		TamanioEnMemoria:  tamanioEnMemoria,
	}

	LogCreacionDeProceso(ContadorPID)
	MoverProcesoACola(pcb, &ColaNew)
	PlanificadorLargoPlazo(Config_Kernel.SchedulerAlgorithm)

}

func TerminarProceso(proceso globals.PCB) {

	if !client.PingCon("Memoria", Config_Kernel.IPMemory, Config_Kernel.PortMemory, Logger) {
		Logger.Debug("No se puede conectar con memoria (Ping no devuelto)")
		return
	}

	respuestaMemoria := LiberarProcesoEnMemoria(proceso.PID)
	if respuestaMemoria {
		MoverProcesoACola(proceso, &ColaExit)
		hayEspacioEnMemoria = true
		PlanificadorLargoPlazo(Config_Kernel.SchedulerAlgorithm)
	}

	LogFinDeProceso(proceso.PID)
}

func AnalizarDesalojo(pid int, pc int, motivoDesalojo string) {

	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaExec)
	pcbDelProceso.PC = pc

	if motivoDesalojo == "Planificador" {
		if pcbDelProceso != nil {
			MoverProcesoACola(*pcbDelProceso, &ColaReady)
			LogDesalojoPorSJF_SRT(pid)
		}
	} else if motivoDesalojo == "IO" {
		if pcbDelProceso != nil {
			Logger.Debug("Desalojo por IO", "pid", pid)
		}
	} else if motivoDesalojo == "DUMP_MEMORY" {
		if pcbDelProceso != nil {
			Logger.Debug("Desalojo por DUMP_MEMORY", "pid", pid)
		}
	} else if motivoDesalojo == "EXIT" {
		if pcbDelProceso != nil {
			TerminarProceso(*pcbDelProceso)
		}
	} else {
		Logger.Debug("Error, motivo de desalojo no válido", "motivo", motivoDesalojo)
	}
}

// &-------------------------------------------Funciones de Administración de IO-------------------------------------------------------------

func RegistrarInstanciaIO(nombre string, puerto int, ip string) {
	instancia := InstanciaIO{NombreIO: nombre, IpIO: ip, PortIO: puerto, Estado: "Libre", PID: -1}

	// Si ya existe el dispositivo IO, agregamos la instancia
	if disp, ok := ListaDispositivosIO[nombre]; ok {
		disp.InstanciasDispositivo = append(disp.InstanciasDispositivo, instancia)

		//Al tener una nueva instancia libre, si hay un proceso en espera, la ocupamos
		if len(disp.ColaEsperaProcesos) != 0 {
			OcuparInstanciaDeIO(nombre, instancia, disp.ColaEsperaProcesos[0].Proceso.PID)
			UsarDispositivoDeIO(nombre, disp.ColaEsperaProcesos[0].Proceso.PID, disp.ColaEsperaProcesos[0].Tiempo)
		}

		//Si no existe el dispositivo IO, lo creamos
	} else {
		ListaDispositivosIO[nombre] = &DispositivoIO{
			ColaEsperaProcesos:    []ProcesoEsperando{},
			InstanciasDispositivo: []InstanciaIO{instancia},
		}
	}

	Logger.Debug("Dispositivo nuevo", "nombre", nombre, "instancias", len(ListaDispositivosIO[nombre].InstanciasDispositivo))
}

func DesconectarInstanciaIO(nombreDispositivo string, ipInstancia string, puertoInstancia int) {
	// Recorro la lista de instancias del dispositivo IO en búsqueda de la instancia a eliminar
	for pos, instanciaBuscada := range ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo {

		// Busco que coincida el IP y el puerto de la instancia
		if instanciaBuscada.IpIO == ipInstancia && instanciaBuscada.PortIO == puertoInstancia {

			// Mandamos al proceso que estaba usando la instancia a Exit
			MoverProcesoDeBlockedAExit(instanciaBuscada.PID)
			Logger.Debug("Desconexion de IO, se envia proceso a Exit", "pid", instanciaBuscada.PID)

			// Borramos la instancia de IO de la lista de instancias del dispositivo IO
			ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo = append(ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo[:pos], ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo[pos+1:]...)

			// Si la instancia desconectada era la única que quedaba...
			if len(ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo) == 0 {

				// Y además hay procesos en espera...
				if len(ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos) != 0 {

					// Recorro la lista de procesos bloqueados por la IO y los mando a Exit
					for _, procesoEnEspera := range ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos {
						MoverProcesoDeBlockedAExit(procesoEnEspera.Proceso.PID)
						Logger.Debug("Desconexion de IO, se envia proceso a Exit", "pid", procesoEnEspera.Proceso.PID)
					}
				}

				// Borramos del mapa de dispositivos IO el dispositivo que ya no tiene instancias
				delete(ListaDispositivosIO, nombreDispositivo)

			} else {
				Logger.Debug("Instancias disponibles", "nombre", nombreDispositivo, "instancias", len(ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo))
			}

			Logger.Debug("Instancia de IO desconectada", "nombre", nombreDispositivo, "ip", ipInstancia, "puerto", puertoInstancia)

		} else {
			Logger.Debug("Instancia de IO no encontrada", "nombre", nombreDispositivo, "ip", ipInstancia, "puerto", puertoInstancia)
		}
	}

}

func VerificarDispositivo(ioName string) bool {
	for nombreDispositivo := range ListaDispositivosIO {
		if nombreDispositivo == ioName {
			return true
		}
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

	Logger.Debug("Instancias de IO", "nombre", nombreDispositivo, "instancias", len(ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo))
}

func OcuparInstanciaDeIO(nombreDispositivo string, instancia InstanciaIO, pid int) {

	for i, instanciaIO := range ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo {

		if instanciaIO == instancia {
			ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo[i].Estado = "Ocupada"
			ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo[i].PID = pid
			Logger.Debug("Instancia de IO ocupada", "nombre", nombreDispositivo, "instancia", instancia.NombreIO, "pid", pid)
			return
		}
	}

	Logger.Debug("Error al ocupar la instancia de IO", "nombre", nombreDispositivo, "instancia", instancia.NombreIO, "pid", pid)
}

func LiberarInstanciaDeIO(nombreDispositivo string, instancia InstanciaIO) {
	for i, instanciaIO := range ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo {
		if instanciaIO == instancia {
			ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo[i].Estado = "Libre"
			ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo[i].PID = -1
			Logger.Debug("Instancia de IO liberada", "nombre", nombreDispositivo, "instancia", instancia.NombreIO)
			return
		}
	}
	Logger.Debug("Error al liberar la instancia de IO", "nombre", nombreDispositivo, "instancia", instancia.NombreIO)
}

func BuscarPrimerInstanciaLibre(nombreDispositivo string) (InstanciaIO, bool) {
	for _, instancia := range ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo {
		if instancia.Estado == "Libre" {
			return instancia, true
		}
	}
	Logger.Debug("No hay instancias libres", "nombre", nombreDispositivo)
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
	ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos = append(ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos, ProcesoEsperando)

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

func ProcesarFinIO(pid int, nombreDispositivo string) {

	// Buscamos en Blocked / SuspBlocked el proceso que terminó su IO y lo mandamos a Ready
	MoverProcesoDeBlockedAReady(pid)

	// Buscamos la instancia de IO que estaba ocupada por el PID
	instanciaDeIO := BuscarInstanciaDeIOporPID(nombreDispositivo, pid)

	// Liberamos la instancia de IO que estaba ocupada por el PID
	LiberarInstanciaDeIO(nombreDispositivo, *instanciaDeIO)

	if len(ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos) != 0 {
		// Si hay procesos esperando en la cola de espera del dispositivo, ocupamos la instancia recientemente liberada
		OcuparInstanciaDeIO(nombreDispositivo, *instanciaDeIO, ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos[0].Proceso.PID)
		UsarDispositivoDeIO(nombreDispositivo, ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos[0].Proceso.PID, ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos[0].Tiempo)
	}

	LogFinDeIO(pid)

}

// &-------------------------------------------Funciones de CPU-------------------------------------------------------------

func VerificarIdentificadorCPU(cpuID string) bool {
	for _, dispositivoCPU := range ListaIdentificadoresCPU {
		if dispositivoCPU.CPUID == cpuID {
			return true
		}
	}
	return false
}

func RegistrarIdentificadorCPU(cpuID string, puerto int, ip string) globals.CPUToKernelHandshakeResponse {

	bodyRespuesta := globals.CPUToKernelHandshakeResponse{
		Modulo: "Kernel",
	}

	//Verificamos si existe el Identificador CPU, y retornamos su posicion en la lista
	if VerificarIdentificadorCPU(cpuID) {
		bodyRespuesta.Respuesta = false
		bodyRespuesta.Mensaje = "El identificador ya existe"
		return bodyRespuesta
	}

	//Creamos un nuevo identificador CPU
	identificadorCPU := IdentificadorCPU{
		CPUID:   cpuID,
		Puerto:  puerto,
		Ip:      ip,
		Ocupado: false,
	}
	//Lo agregamos a la lista de identificadores CPU
	ListaIdentificadoresCPU = append(ListaIdentificadoresCPU, identificadorCPU)
	Logger.Debug("Identificador CPU nuevo", "cpu_id", identificadorCPU.CPUID, "puerto", identificadorCPU.Puerto, "ip", identificadorCPU.Ip)

	bodyRespuesta.Respuesta = true
	bodyRespuesta.Mensaje = "Identificador CPU registrado correctamente"
	return bodyRespuesta
}

func ObtenerCpuDisponible() *IdentificadorCPU {
	for i := range ListaIdentificadoresCPU {
		if !ListaIdentificadoresCPU[i].Ocupado {
			return &ListaIdentificadoresCPU[i]
		}
	}
	return nil
}

func ElegirCpuYMandarProceso(proceso globals.PCB) bool {

	cpu := ObtenerCpuDisponible()
	if cpu != nil {
		cpu.Ocupado = true
		cpu.PID = proceso.PID
		Logger.Debug("CPU elegida: ", "cpu_id", cpu.CPUID, ", Mandando proceso_pid: ", proceso.PID)
		EnviarProcesoACPU(cpu.Ip, cpu.Puerto, proceso.PID, proceso.PC)
	} else {
		Logger.Debug("No hay CPU disponible para el proceso ", "proceso_pid", proceso.PID)
		return false
	}
	return true
}

func BuscarCPUporPID(pid int) *IdentificadorCPU {
	for _, cpu := range ListaIdentificadoresCPU {
		if cpu.PID == pid {
			return &cpu
		}
	}
	return nil
}

func IniciarContadorBlocked(pcb globals.PCB, milisegundos int) {
	cancel := make(chan struct{})
	canceladoresBlocked[pcb.PID] = cancel

	go func() {
		timer := time.NewTimer(time.Duration(milisegundos) * time.Millisecond)
		select {
		case <-timer.C:
			// Verifica que el proceso siga en Blocked
			if BuscarProcesoEnCola(pcb.PID, &ColaBlocked) != nil {
				MoverProcesoACola(pcb, &ColaSuspBlocked)
				//! PEDIR A MEMORIA QUE HAGA EL SWAP
			}
		case <-cancel:
			timer.Stop()
			// El proceso salio de Blocked antes de tiempo
		}
		delete(canceladoresBlocked, pcb.PID)
	}()
}

// Llama a esto cuando el proceso salga de Blocked por otro motivo
func CancelarContadorBlocked(pid int) {
	if cancel, ok := canceladoresBlocked[pid]; ok {
		close(cancel)
		delete(canceladoresBlocked, pid)
	}
}
