package kernel_internal

/* INIT_PROC, esta syscall recibirá 2 parámetros de la CPU,
el primero será el nombre del archivo de pseudocódigo que deberá ejecutar el proceso
y el segundo parámetro es el tamaño del proceso en Memoria.
El Kernel creará un nuevo PCB y lo dejará en estado NEW, esta syscall no implica cambio de estado,
por lo que el proceso que llamó a esta syscall, inmediatamente volverá a ejecutar en la CPU. */

func SyscallInitProc(pid int, nombreArchivo string, tamanioProcesoEnMemoria int) {
	LogSyscall(pid, "INIT_PROC")

	// ! FALTA IMPLEMENTAR DENTRO DE PCB EL ARCHIVO DE PSEUDOCODIGO
	InicializarPCB(tamanioProcesoEnMemoria)
}

/* EXIT, esta syscall no recibirá parámetros y
se encargará de finalizar el proceso que la invocó,
siguiendo lo descrito anteriormente para Finalización de procesos. */

func SyscallExit(pid int) {
	LogSyscall(pid, "EXIT")

	//Busco el PCB en la lista de Exit
	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaExit)
	//Termino el proceso, avisandole memoria que libere el espacio y buscando
	TerminarProceso(*pcbDelProceso)
}

/* DUMP_MEMORY, esta syscall le solicita a la memoria, junto al PID que lo solicitó, que haga un Dump del proceso.
Esta syscall bloqueará al proceso que la invocó hasta que el módulo memoria confirme la finalización de la operación,
en caso de error, el proceso se enviará a EXIT.
Caso contrario, se desbloquea normalmente pasando a READY.
*/

func SyscallDumpMemory(pid int) {
	LogSyscall(pid, "DUMP_MEMORY")

	pcbDelProceso := BuscarProcesoEnCola(pid, &ColaExec)
	MoverProcesoACola(*pcbDelProceso, &ColaBlocked)
	respuestaDelDump := PedirDumpMemory(pid)
	if respuestaDelDump {
		MoverProcesoACola(*pcbDelProceso, &ColaReady)
	} else {
		MoverProcesoACola(*pcbDelProceso, &ColaExit)
	}
}

func SyscallEntradaSalida(pid int) {
	LogSyscall(pid, "IO")

}
