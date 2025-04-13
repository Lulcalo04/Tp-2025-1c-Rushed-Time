package kernel_internal

import (
	"fmt"
	"log"
)

/*Syscall recibida: “## (<PID>) - Solicitó syscall: <NOMBRE_SYSCALL>”
Creación de Proceso: “## (<PID>) Se crea el proceso - Estado: NEW”
Motivo de Bloqueo: “## (<PID>) - Bloqueado por IO: <DISPOSITIVO_IO>”
Fin de IO: “## (<PID>) finalizó IO y pasa a READY”
Desalojo de SJF: “## (<PID>) - Desalojado por fin de SJF”
Fin de Proceso: “## (<PID>) - Finaliza el proceso”
Métricas de Estado: “## (<PID>) - Métricas de estado: NEW (NEW_COUNT) (NEW_TIME), READY (READY_COUNT) (READY_TIME), …”*/

// LogSyscall logs de syscall recibida
func LogSyscall(pid int, syscallName string) {
	log.Printf("## (%d) - Solicitó syscall: %s\n", pid, syscallName)
}

// LogProcessCreation logs de creacion de un proceso
func LogCreacionDeProceso(pid int) {
	log.Printf("## (%d) Se crea el proceso - Estado: NEW\n", pid)
}

// LogIOBlock logs de motivo de bloqueo
func LogBloqueoPorIO(pid int, ioDevice string) {
	log.Printf("## (%d) - Bloqueado por IO: %s\n", pid, ioDevice)
}

// LogIOCompletion logs de fin de I/O
func LogFinDeeIO(pid int) {
	log.Printf("## (%d) finalizó IO y pasa a READY\n", pid)
}

// LogSJFPreemption logs de desalojo de SJF
func LogDesalojoPorSJF(pid int) {
	log.Printf("## (%d) - Desalojado por fin de SJF\n", pid)
}

// LogProcessTermination logs de terminar un proceso
func LogFinDeProceso(pid int) {
	log.Printf("## (%d) - Finaliza el proceso\n", pid)
}

// LogStateMetrics logs de metricas de un
func LogMetricasDeEstado(pid int, metrics map[string][2]int) {
	// Example: metrics["NEW"] = [NEW_COUNT, NEW_TIME]
	logMessage := fmt.Sprintf("## (%d) - Métricas de estado:", pid)
	for state, values := range metrics {
		logMessage += fmt.Sprintf(" %s (%d) (%d),", state, values[0], values[1])
	}
	// Remove trailing comma and log the message
	log.Println(logMessage[:len(logMessage)-1])
}
