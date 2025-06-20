package kernel_internal

import (
	"fmt"
	"utils/globals"
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

func LogMetricasDeEstado(pcb globals.PCB) {
	Logger.Info(fmt.Sprintf(
		"## (%d) - Métricas de estado: NEW (%d) (%d), READY (%d) (%d), EXEC (%d) (%d), BLOCKED (%d) (%d), SUSP_BLOCKED (%d) (%d), SUSP_READY (%d) (%d), EXIT (%d) (%d)",
		pcb.PID,
		pcb.MetricasDeEstados[globals.New], pcb.MetricasDeTiempos[globals.New],
		pcb.MetricasDeEstados[globals.Ready], pcb.MetricasDeTiempos[globals.Ready],
		pcb.MetricasDeEstados[globals.Exec], pcb.MetricasDeTiempos[globals.Exec],
		pcb.MetricasDeEstados[globals.Blocked], pcb.MetricasDeTiempos[globals.Blocked],
		pcb.MetricasDeEstados[globals.SuspBlocked], pcb.MetricasDeTiempos[globals.SuspBlocked],
		pcb.MetricasDeEstados[globals.SuspReady], pcb.MetricasDeTiempos[globals.SuspReady],
		pcb.MetricasDeEstados[globals.Exit], pcb.MetricasDeTiempos[globals.Exit],
	))
}
