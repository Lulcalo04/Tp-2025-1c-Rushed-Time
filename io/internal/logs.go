package io_internal

import "fmt"

/*
Inicio de IO: “## PID: <PID> - Inicio de IO - Tiempo: <TIEMPO_IO>”.
Finalización de IO: “## PID: <PID> - Fin de IO”.
*/

func LogInicioIO(pid int, ioTime int) {
	Logger.Info(fmt.Sprintf("## PID: %d - Inicio de IO - Tiempo: %d", pid, ioTime))
}

func LogFinalizacionIO(pid int) {
	Logger.Info(fmt.Sprintf("## PID: %d - Fin de IO", pid))
}
