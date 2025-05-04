package kernel_internal

import (
	"fmt"
)

/*
Syscall recibida: “## (<PID>) - Solicitó syscall: <NOMBRE_SYSCALL>”
Creación de Proceso: “## (<PID>) Se crea el proceso - Estado: NEW”
Motivo de Bloqueo: “## (<PID>) - Bloqueado por IO: <DISPOSITIVO_IO>”
Fin de IO: “## (<PID>) finalizó IO y pasa a READY”
Desalojo de SJF: “## (<PID>) - Desalojado por fin de SJF”
Fin de Proceso: “## (<PID>) - Finaliza el proceso”
Métricas de Estado: “## (<PID>) - Métricas de estado: NEW (NEW_COUNT) (NEW_TIME), READY (READY_COUNT) (READY_TIME), …”
*/

// LogSyscall logs de syscall recibida
func LogSyscall(pid int, syscallName string) {
	Logger.Info(fmt.Sprintf("## (%d) - Solicitó syscall: %s", pid, syscallName))
}

func LogCreacionDeProceso(pid int) {
	Logger.Info(fmt.Sprintf("## (%d) Se crea el proceso - Estado: NEW", pid))
}

func LogCambioDeEstado(pid int, estadoAnterior, estadoActual string) {
	Logger.Info(fmt.Sprintf("## (%d) Pasa del estado %s al estado %s", pid, estadoAnterior, estadoActual))
}

func LogMotivoDeBloqueo(pid int, ioDevice string) {
	Logger.Info(fmt.Sprintf("## (%d) - Bloqueado por IO: %s", pid, ioDevice))
}

func LogFinDeIO(pid int) {
	Logger.Info(fmt.Sprintf("## (%d) finalizó IO y pasa a READY", pid))
}

func LogDesalojoPorSJF_SRT(pid int) {
	Logger.Info(fmt.Sprintf("## (%d) - Desalojado por fin de SJF/SRT", pid))
}

func LogFinDeProceso(pid int) {
	Logger.Info(fmt.Sprintf("## (%d) - Finaliza el proceso", pid))
}

func LogMetricasDeEstado(pid int, metrics map[string][2]int) {
	logMessage := fmt.Sprintf("## (%d) - Métricas de estado:", pid)
	for state, values := range metrics {
		logMessage += fmt.Sprintf(" %s (%d) (%d),", state, values[0], values[1])
	}
	Logger.Info(logMessage[:len(logMessage)-1]) // Elimina la coma final
}
