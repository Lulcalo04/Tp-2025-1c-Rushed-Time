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
	//Desalojo el proceso de la CPU
	PeticionDesalojo(pid, "EXIT")
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

	//Desalojo el proceso de la CPU
	PeticionDesalojo(pid, "DUMP_MEMORY")

	//Bloque el proceso
	MoverProcesoABlocked(pid)

	//Pido el dump a memoria y espero la respuesta
	respuestaDelDump := PedirDumpMemory(pid)

	//Analizo la respuesta
	if respuestaDelDump {
		MoverProcesoDeBlockedAReady(pid)
	} else {
		MoverProcesoDeBlockedAExit(pid)
	}
}

func SyscallEntradaSalida(pid int, nombreDispositivo string, milisegundosDeUso int) {
	LogSyscall(pid, "IO")

	if VerificarDispositivo(nombreDispositivo) {
		//& SI EL DISPOSITIVO EXISTE, PERO ESTA EN USO, BLOQUEAR EL PROCESO EN LA COLA DEL DISPOSITIVO
		MoverProcesoABlocked(pid)
		if VerificarInstanciaDeIO(nombreDispositivo) {
			//Si hay instancias de IO disponibles, se bloquea el proceso por estar usando la IO
			UsarDispositivoDeIO(nombreDispositivo, pid, milisegundosDeUso)
			MoverProcesoDeBlockedAReady(pid)
		} else {
			//Si no hay instancias de IO disponibles, se bloquea el proceso en la cola del dispositivo
			BloquearProcesoPorIO(nombreDispositivo, pid, milisegundosDeUso) //* FUNCION A DESARROLLAR
		}
	} else {
		//& NO EXISTE EL DISPOSITIVO, ENTONCES SE MANDA EL PROCESO A EXIT
		//! REALIZAR PEDIDO DE DESALOJO EN LA CPU

		// Busco el PCB en la lista de Exec
		pcbDelProceso := BuscarProcesoEnCola(pid, &ColaExec)
		// Mando el proceso a Exit
		MoverProcesoACola(*pcbDelProceso, &ColaExit)
	}
}
