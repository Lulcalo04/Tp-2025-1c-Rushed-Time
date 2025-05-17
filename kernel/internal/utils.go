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
	InstanciasDispositivo []InstanciaIO
	ColaEsperaProcesos    []globals.PCB
}

type InstanciaIO struct {
	NombreIO string
	IpIO     string
	PortIO   int
	Estado   string
}

var ListaDispositivosIO map[string]*DispositivoIO

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

func RegistrarInstanciaIO(nombre string, puerto int, ip string) {
	instancia := InstanciaIO{NombreIO: nombre, IpIO: ip, PortIO: puerto, Estado: "Libre"}

	if disp, ok := ListaDispositivosIO[nombre]; ok {
		disp.InstanciasDispositivo = append(disp.InstanciasDispositivo, instancia)
	} else {
		ListaDispositivosIO[nombre] = &DispositivoIO{
			ColaEsperaProcesos:    []globals.PCB{},
			InstanciasDispositivo: []InstanciaIO{instancia},
		}
	}
	Logger.Debug("Dispositivo nuevo", "nombre", nombre, "instancias", len(ListaDispositivosIO[nombre].InstanciasDispositivo))
}

func VerificarDispositivo(ioName string) bool {
	for nombreDispositivo, _ := range ListaDispositivosIO {
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
	//Buscamos el PCB en la cola de blocked
	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaBlocked)
	//Lo agregamos a la cola de espera del dispositivo
	ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos = append(ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos, *pcbDelProceso)

	//Ejecuto el primer proceso de la cola de espera (en bucle)
	for len(ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos) != 0 {
		instanciaDeIO, hayInstancias := BuscarPrimerInstanciaLibre(nombreDispositivo)
		if !hayInstancias {
			BloquearProcesoPorIO(nombreDispositivo, pid)
			return
		}
		EnviarProcesoAIO(instanciaDeIO, ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos[0].PID, milisegundosDeUso)
	}

	Logger.Debug("Instancias de IO", "nombre", nombreDispositivo, "instancias", len(ListaDispositivosIO[nombreDispositivo].InstanciasDispositivo))
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

func BloquearProcesoPorIO(nombreDispositivo string, pid int) {
	//Buscamos el PCB en la cola de blocked
	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaBlocked)
	//Agregar el proceso a la cola de espera del dispositivo
	ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos = append(ListaDispositivosIO[nombreDispositivo].ColaEsperaProcesos, *pcbDelProceso)

}
