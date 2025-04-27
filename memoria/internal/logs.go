package memoria_internal

import (
	"fmt"
	"log"
)

/*
Creación de Proceso: “## PID: <PID> - Proceso Creado - Tamaño: <TAMAÑO>”
Destrucción de Proceso: “## PID: <PID> - Proceso Destruido - Métricas - Acc.T.Pag: <ATP>; Inst.Sol.: <Inst.Sol.>; SWAP: <SWAP>; Mem.Prin.: <Mem.Prin.>; Lec.Mem.: <Lec.Mem.>; Esc.Mem.: <Esc.Mem.>”
Obtener instrucción: “## PID: <PID> - Obtener instrucción: <PC> - Instrucción: <INSTRUCCIÓN> <...ARGS>”
Escritura / lectura en espacio de usuario: “## PID: <PID> - <Escritura/Lectura> - Dir. Física: <DIRECCIÓN_FÍSICA> - Tamaño: <TAMAÑO>”
Memory Dump: “## PID: <PID> - Memory Dump solicitado”
*/

// LogCreacionDeProceso logs de creación de un proceso con su tamaño
func LogCreacionDeProceso(pid int, size int) {
	log.Printf("## PID: %d - Proceso Creado - Tamaño: %d\n", pid, size)
}

// LogDestruccionDeProceso logs de destrucción de un proceso con sus métricas
func LogDestruccionDeProceso(pid int, atp int, instSol int, swap int, memPrin int, lecMem int, escMem int) {
	log.Printf("## PID: %d - Proceso Destruido - Métricas - Acc.T.Pag: %d; Inst.Sol.: %d; SWAP: %d; Mem.Prin.: %d; Lec.Mem.: %d; Esc.Mem.: %d\n",
		pid, atp, instSol, swap, memPrin, lecMem, escMem)
}

// LogObtenerInstruccion logs de obtener instrucción
func LogObtenerInstruccion(pid int, pc int, instruccion string, args ...string) {
	log.Printf("## PID: %d - Obtener instrucción: %d - Instrucción: %s %s\n", pid, pc, instruccion, fmt.Sprint(args))
}

// LogEspacioUsuario logs de escritura/lectura en espacio de usuario
func LogOperacionEnEspacioUsuario(pid int, operacion string, direccionFisica int, size int) {
	log.Printf("## PID: %d - %s - Dir. Física: %d - Tamaño: %d\n", pid, operacion, direccionFisica, size)
}

// LogMemoryDump logs de memory dump solicitado
func LogMemoryDump(pid int) {
	log.Printf("## PID: %d - Memory Dump solicitado\n", pid)
}
