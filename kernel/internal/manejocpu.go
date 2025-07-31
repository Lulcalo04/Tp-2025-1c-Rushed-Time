package kernel_internal

import (
	"globals"
	"sync"
	"time"
)

type IdentificadorCPU struct {
	CPUID              string
	Puerto             int
	Ip                 string
	Ocupado            bool
	DesalojoSolicitado bool
	PID                int
}

var ListaIdentificadoresCPU []IdentificadorCPU = make([]IdentificadorCPU, 0)

var MutexIdentificadoresCPU sync.Mutex // ! Mutex para proteger el acceso a la lista de identificadores CPU

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
		CPUID:              cpuID,
		Puerto:             puerto,
		Ip:                 ip,
		Ocupado:            false,
		DesalojoSolicitado: false,
	}

	//Lo agregamos a la lista de identificadores CPU
	MutexIdentificadoresCPU.Lock()
	ListaIdentificadoresCPU = append(ListaIdentificadoresCPU, identificadorCPU)
	MutexIdentificadoresCPU.Unlock()

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

		if !proceso.DesalojoAnalizado {
			Logger.Debug("No se analizó desalojo, no se puede mandar a CPU todavía", "proceso_pid", proceso.PID)
			//fmt.Println("No se analizó desalojo, no se puede mandar a CPU todavía", "proceso PID:", proceso.PID)
			time.Sleep(time.Second * 1) // Esperamos un segundo antes de volver a intentar
			return false
		}

		MutexIdentificadoresCPU.Lock()
		cpu.Ocupado = true
		cpu.PID = proceso.PID
		MutexIdentificadoresCPU.Unlock()

		// Cambiamos el boolean de YaCalculado a false para que se vuelva a calcular en la próxima iteración
		proceso.EstimacionDeRafaga.YaCalculado = false

		// Movemos el proceso a la cola Exec
		MoverProcesoACola(&proceso, &ColaExec)
		Logger.Debug("Proceso movido a cola Exec", "proceso_pid", proceso.PID, "Inicio de ejecución", proceso.InicioEjecucion.Format("2006-01-02 15:04:05"))
		//fmt.Println("Proceso", proceso.PID, "movido a la cola Exec", "Inicio de ejecución:", proceso.InicioEjecucion.Format("2006-01-02 15:04:05"))

		Logger.Debug("CPU elegida: ", "cpu_id", cpu.CPUID, ", Mandando proceso_pid: ", proceso.PID)
		//fmt.Println("CPU elegida:", cpu.CPUID, ", Mandando proceso PID:", proceso.PID)
		EnviarProcesoACPU(cpu.Ip, cpu.Puerto, proceso.PID, proceso.PC)

		//Me fijo si hay alguna otra CPU libre y si no la hay marco que ya no queda ninguna libre
		MutexIdentificadoresCPU.Lock()
		for i := range ListaIdentificadoresCPU {
			if !ListaIdentificadoresCPU[i].Ocupado {
				MutexIdentificadoresCPU.Unlock()
				MutexCpuLibres.Lock()
				CpuLibres = true
				MutexCpuLibres.Unlock()
				return true
			}
		}
		MutexIdentificadoresCPU.Unlock()

		MutexCpuLibres.Lock()
		CpuLibres = false // Indicamos que no hay CPUs libres
		MutexCpuLibres.Unlock()

		Logger.Debug("No queda ninguna CPU libre")
		//fmt.Println("No queda ninguna CPU libre")

		return true
	} else {
		Logger.Debug("No hay CPU disponible para el proceso ", "proceso_pid", proceso.PID)
		//fmt.Println("No hay CPU disponible para el proceso", proceso.PID)
		MutexCpuLibres.Lock()
		CpuLibres = false // Indicamos que no hay CPUs libres
		MutexCpuLibres.Unlock()

		return false
	}
}

func BuscarCPUporPID(pid int) *IdentificadorCPU {
	for _, cpu := range ListaIdentificadoresCPU {
		if cpu.PID == pid {
			return &cpu
		}
	}
	return nil
}
