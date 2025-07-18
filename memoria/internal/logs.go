package memoria_internal

import "fmt"

/*
Creación de Proceso: “## PID: <PID> - Proceso Creado - Tamaño: <TAMAÑO>”
Destrucción de Proceso: “## PID: <PID> - Proceso Destruido - Métricas - Acc.T.Pag: <ATP>; Inst.Sol.: <Inst.Sol.>; SWAP: <SWAP>; Mem.Prin.: <Mem.Prin.>; Lec.Mem.: <Lec.Mem.>; Esc.Mem.: <Esc.Mem.>”
Obtener instrucción: “## PID: <PID> - Obtener instrucción: <PC> - Instrucción: <INSTRUCCIÓN> <...ARGS>”
Escritura / lectura en espacio de usuario: “## PID: <PID> - <Escritura/Lectura> - Dir. Física: <DIRECCIÓN_FÍSICA> - Tamaño: <TAMAÑO>”
Memory Dump: “## PID: <PID> - Memory Dump solicitado”
*/

func LogCreacionDeProceso(pid int, size int) {
	Logger.Info(fmt.Sprintf("## PID: %d - Proceso Creado - Tamaño: %d", pid, size))
	fmt.Printf("\n## PID: %d - Proceso Creado - Tamaño: %d\n", pid, size)
}

func LogDestruccionDeProceso(pid int, metricas MetricasPorProceso) {
	Logger.Info(fmt.Sprintf("## PID: %d - Proceso Destruido - Métricas - Acc.T.Pag: %d; Inst.Sol.: %d; SWAP: %d; Mem.Prin.: %d; Lec.Mem.: %d; Esc.Mem.: %d",
		pid, metricas.AccesoATablaDePaginas, metricas.InstruccionesSolicitadas, metricas.BajadasASwap, metricas.SubidasAMemoriaPrincipal, metricas.LecturasDeMemoria, metricas.EscriturasDeMemoria))
	fmt.Printf("\n## PID: %d - Proceso Destruido - Métricas - Acc.T.Pag: %d; Inst.Sol.: %d; SWAP: %d; Mem.Prin.: %d; Lec.Mem.: %d; Esc.Mem.: %d",
		pid, metricas.AccesoATablaDePaginas, metricas.InstruccionesSolicitadas, metricas.BajadasASwap, metricas.SubidasAMemoriaPrincipal, metricas.LecturasDeMemoria, metricas.EscriturasDeMemoria)
}

func LogObtenerInstruccion(pid int, pc int, instruccion string, args ...string) {
	Logger.Info(fmt.Sprintf("## PID: %d - Obtener instrucción: %d - Instrucción: %s %s",
		pid, pc, instruccion, fmt.Sprint(args))) // no pasamos el args, porque no sabemos que es
	fmt.Printf("\n## PID: %d - Obtener instrucción: %d - Instrucción: %s %s", pid, pc, instruccion, fmt.Sprint(args))
}

func LogOperacionEnEspacioUsuario(pid int, operacion string, direccionFisica int, size int) {
	Logger.Info(fmt.Sprintf("## PID: %d - %s - Dir. Física: %d - Tamaño: %d", pid, operacion, direccionFisica, size))
	fmt.Printf("\n## PID: %d - %s - Dir. Física: %d - Tamaño: %d", pid, operacion, direccionFisica, size)
}

func LogMemoryDump(pid int) {
	Logger.Info(fmt.Sprintf("## PID: %d - Memory Dump solicitado", pid))
	fmt.Printf("\n## PID: %d - Memory Dump solicitado", pid)
}
