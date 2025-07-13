package kernel_internal

import (
	"fmt"
	"globals"
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
	message := fmt.Sprintf("## (%d) - Solicitó syscall: %s", pid, syscallName)
	Logger.Info(message)
	fmt.Println(message)
}

func LogCreacionDeProceso(pid int) {
	message := fmt.Sprintf("## (%d) Se crea el proceso - Estado: NEW", pid)
	Logger.Info(message)
	fmt.Println(message)
}

func LogCambioDeEstado(pid int, estadoAnterior, estadoActual string) {
	message := fmt.Sprintf("## (%d) Pasa del estado %s al estado %s", pid, estadoAnterior, estadoActual)
	Logger.Info(message)
	fmt.Println(message)
}

func LogMotivoDeBloqueo(pid int, ioDevice string) {
	message := fmt.Sprintf("## (%d) - Bloqueado por IO: %s", pid, ioDevice)
	Logger.Info(message)
	fmt.Println(message)
}

func LogFinDeIO(pid int) {
	message := fmt.Sprintf("## (%d) finalizó IO y pasa a READY", pid)
	Logger.Info(message)
	fmt.Println(message)
}

func LogDesalojoPorSJF_SRT(pid int) {
	message := fmt.Sprintf("## (%d) - Desalojado por fin de SJF/SRT", pid)
	Logger.Info(message)
	fmt.Println(message)
}

func LogFinDeProceso(pid int) {
	message := fmt.Sprintf("## (%d) - Finaliza el proceso", pid)
	Logger.Info(message)
	fmt.Println(message)
}

func LogMetricasDeEstado(pcb globals.PCB) {
	message := fmt.Sprintf(
		"## (%d) - Métricas de estado: NEW (%d) (%dms), READY (%d) (%dms), EXEC (%d) (%dms), BLOCKED (%d) (%dms), SUSP_BLOCKED (%d) (%dms), SUSP_READY (%dms) (%d), EXIT (%d) (%dms)",
		pcb.PID,
		pcb.MetricasDeEstados[globals.New], pcb.MetricasDeTiempos[globals.New].Milliseconds(),
		pcb.MetricasDeEstados[globals.Ready], pcb.MetricasDeTiempos[globals.Ready].Milliseconds(),
		pcb.MetricasDeEstados[globals.Exec], pcb.MetricasDeTiempos[globals.Exec].Milliseconds(),
		pcb.MetricasDeEstados[globals.Blocked], pcb.MetricasDeTiempos[globals.Blocked].Milliseconds(),
		pcb.MetricasDeEstados[globals.SuspBlocked], pcb.MetricasDeTiempos[globals.SuspBlocked].Milliseconds(),
		pcb.MetricasDeEstados[globals.SuspReady], pcb.MetricasDeTiempos[globals.SuspReady].Milliseconds(),
		pcb.MetricasDeEstados[globals.Exit], pcb.MetricasDeTiempos[globals.Exit].Milliseconds(),
	)
	Logger.Info(message)
	fmt.Println(message)
}
