package kernel_internal

import (
	"fmt"
	"globals"
)

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

		// Movemos el proceso a la cola Exec
		MoverProcesoACola(&proceso, &ColaExec)
		Logger.Debug("Proceso movido a cola Exec", "proceso_pid", proceso.PID, "Inicio de ejecución", proceso.InicioEjecucion)
		fmt.Println("Proceso", proceso.PID, "movido a la cola Exec", "Inicio de ejecución:", proceso.InicioEjecucion)

		Logger.Debug("CPU elegida: ", "cpu_id", cpu.CPUID, ", Mandando proceso_pid: ", proceso.PID)
		EnviarProcesoACPU(cpu.Ip, cpu.Puerto, proceso.PID, proceso.PC)

		fmt.Println("Proceso", proceso.PID, "enviado a la CPU", cpu.CPUID)

		return true
	} else {
		Logger.Debug("No hay CPU disponible para el proceso ", "proceso_pid", proceso.PID)
		fmt.Println("No hay CPU disponible para el proceso", proceso.PID)
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
