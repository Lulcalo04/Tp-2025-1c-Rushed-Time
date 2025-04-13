package io_internal

/*Inicio de IO: “## PID: <PID> - Inicio de IO - Tiempo: <TIEMPO_IO>”.
Finalización de IO: “## PID: <PID> - Fin de IO”.*/

import (
	"log"
)

// LogIOStart logs the start of an IO operation
func LogInicioIO(pid int, ioTime int) {
	log.Printf("## PID: %d - Inicio de IO - Tiempo: %d\n", pid, ioTime)
}

// LogIOFinish logs the completion of an IO operation
func LogFinalizacionIO(pid int) {
	log.Printf("## PID: %d - Fin de IO\n", pid)
}
