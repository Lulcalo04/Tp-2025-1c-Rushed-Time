package kernel_internal

import (
	"log/slog"
	"os"
	"strconv"
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
	SuspensionTime        int     `json:"suspension_time"`
	LogLevel              string  `json:"log_level"`
}

type DispositivoIO struct {
	NombreIO     string
	IpIO         string
	PortIO       int
	InstanciasIO int
}

type IdentificadorCPU struct {
	CPUID   string
	Puerto  int
	Ip      string
	Ocupado bool
}

var ListaDispositivosIO []DispositivoIO

var ListaIdentificadoresCPU []IdentificadorCPU

var Config_Kernel *ConfigKernel

var Logger *slog.Logger

var ContadorPID int

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
	//IniciarPlanificadores()
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

func AnalizarDesalojo(pid int, motivoDesalojo string) {

	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaExec)

	if motivoDesalojo == "Planificador" {
		if pcbDelProceso != nil {
			MoverProcesoACola(*pcbDelProceso, &ColaReady)
			LogDesalojoPorSJF_SRT(pid)
		}
	} else if motivoDesalojo == "IO" {
		if pcbDelProceso != nil {
			Logger.Debug("Desalojo por IO", "pid", pid)
		}
	} else if motivoDesalojo == "FinProceso" {
		if pcbDelProceso != nil {
			TerminarProceso(*pcbDelProceso)
		}
	} else {
		Logger.Debug("Error, motivo de desalojo no válido", "motivo", motivoDesalojo)
	}
}

// &-------------------------------------------Funciones de Administración de IO-------------------------------------------------------------

func RegistrarDispositivoIO(ipIO string, puertoIO int, ioName string) DispositivoIO {

	//Verificamos si existe el dispositivo IO, y retornamos su posicion en la lista
	existeDispositivo, posDispositivo := VerificarDispositivo(ioName)

	if existeDispositivo {
		//Si existe, le sumamos una instancia
		ListaDispositivosIO[posDispositivo].InstanciasIO++
		//Retornamos el dispositivo actualizado
		return ListaDispositivosIO[posDispositivo]

	} else {
		//Creamos un nuevo dispositivo IO
		dispositivoIO := DispositivoIO{
			NombreIO:     ioName,
			IpIO:         ipIO,
			PortIO:       puertoIO,
			InstanciasIO: 1,
		}
		//Lo agregamos a la lista de dispositivos IO
		ListaDispositivosIO = append(ListaDispositivosIO, dispositivoIO)

		Logger.Debug("Dispositivo nuevo", "nombre", dispositivoIO.NombreIO, "instancias", dispositivoIO.InstanciasIO)

		//Retornamos el nuevo dispositivo IO
		return dispositivoIO
	}
}

func VerificarDispositivo(ioName string) (bool, int) {
	for posDispositivo, dispositivoIO := range ListaDispositivosIO {
		if dispositivoIO.NombreIO == ioName {
			return true, posDispositivo
		}
	}
	return false, -1
}

func VerificarInstanciaDeIO(posDispositivo int) bool {
	if ListaDispositivosIO[posDispositivo].InstanciasIO >= 1 {
		ListaDispositivosIO[posDispositivo].InstanciasIO--
		Logger.Debug("Instancias de IO", "nombre", ListaDispositivosIO[posDispositivo].NombreIO, "instancias", ListaDispositivosIO[posDispositivo].InstanciasIO)
		return true
	} else {
		return false
	}
}

func UsarDispositivoDeIO(posDispositivo int, pid int, milisegundosDeUso int) {
	//Hacer peticion de uso de dispositivo IO
	EnviarProcesoAIO(ListaDispositivosIO[posDispositivo], pid, milisegundosDeUso)
	//Agregar instancia de IO
	ListaDispositivosIO[posDispositivo].InstanciasIO++
	Logger.Debug("Instancias de IO", "nombre", ListaDispositivosIO[posDispositivo].NombreIO, "instancias", ListaDispositivosIO[posDispositivo].InstanciasIO)
}

func VerificarIdentificadorCPU(cpuID string) bool {
	for _, dispositivoCPU := range ListaIdentificadoresCPU {
		if dispositivoCPU.CPUID == cpuID {
			return true
		}
	}
	return false
}

func RegistrarIdentificadorCPU(cpuID string, puerto int, ip string) globals.CPUHandshakeResponse {

	bodyRespuesta := globals.CPUHandshakeResponse{
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

func ElegirCpuYMandarProceso(proceso globals.PCB) {

	cpu := ObtenerCpuDisponible()
	if cpu != nil {
		cpu.Ocupado = true
		Logger.Debug("CPU elegida: ", "cpu_id", cpu.CPUID, ", Mandando proceso_pid: ", proceso.PID)
	} else {
		Logger.Debug("No hay CPU disponible para el proceso ", "proceso_pid", proceso.PID)
		return
	}

	// Enviar el proceso a la CPU elegida
	EnviarProcesoACPU(cpu.Ip, cpu.Puerto, proceso.PID, proceso.PC)

}
