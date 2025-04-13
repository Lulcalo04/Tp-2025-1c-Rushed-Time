package cpu_internal

/* Fetch Instrucción: “## PID: <PID> - FETCH - Program Counter: <PROGRAM_COUNTER>”.
Interrupción Recibida: “## Llega interrupción al puerto Interrupt”.
Instrucción Ejecutada: “## PID: <PID> - Ejecutando: <INSTRUCCION> - <PARAMETROS>”.
Lectura/Escritura Memoria: “PID: <PID> - Acción: <LEER / ESCRIBIR> - Dirección Física: <DIRECCION_FISICA> - Valor: <VALOR LEIDO / ESCRITO>”.
Obtener Marco: “PID: <PID> - OBTENER MARCO - Página: <NUMERO_PAGINA> - Marco: <NUMERO_MARCO>”.
TLB Hit: “PID: <PID> - TLB HIT - Pagina: <NUMERO_PAGINA>”
TLB Miss: “PID: <PID> - TLB MISS - Pagina: <NUMERO_PAGINA>”
Página encontrada en Caché: “PID: <PID> - Cache Hit - Pagina: <NUMERO_PAGINA>”
Página faltante en Caché: “PID: <PID> - Cache Miss - Pagina: <NUMERO_PAGINA>”
Página ingresada en Caché: “PID: <PID> - Cache Add - Pagina: <NUMERO_PAGINA>”
Página Actualizada de Caché a Memoria: “PID: <PID> - Memory Update - Página: <NUMERO_PAGINA> - Frame: <FRAME_EN_MEMORIA_PRINCIPAL>”*/

import "log"

func LogFetchInstruccion(pid int, programCounter int) {
	log.Printf("## PID: %d - FETCH - Program Counter: %d\n", pid, programCounter)
}

func LogInterrupcionRecibida() {
	log.Printf("## Llega interrupción al puerto Interrupt\n")
}

func LogInstruccionEjecutada(pid int, instruccion string, parametros string) {
	log.Printf("## PID: %d - Ejecutando: %s - %s\n", pid, instruccion, parametros)
}

func LogLecturaEscrituraMemoria(pid int, accion string, direccionFisica int, valor string) {
	log.Printf("PID: %d - Acción: %s - Dirección Física: %d - Valor: %s\n", pid, accion, direccionFisica, valor)
}

func LogObtenerMarco(pid int, numeroPagina int, numeroMarco int) {
	log.Printf("PID: %d - OBTENER MARCO - Página: %d - Marco: %d\n", pid, numeroPagina, numeroMarco)
}

func LogTLBHit(pid int, numeroPagina int) {
	log.Printf("PID: %d - TLB HIT - Pagina: %d\n", pid, numeroPagina)
}

func LogTLBMiss(pid int, numeroPagina int) {
	log.Printf("PID: %d - TLB MISS - Pagina: %d\n", pid, numeroPagina)
}

func LogPaginaEncontradaEnCache(pid int, numeroPagina int) {
	log.Printf("PID: %d - Cache Hit - Pagina: %d\n", pid, numeroPagina)
}

func LogPaginaFaltanteEnCache(pid int, numeroPagina int) {
	log.Printf("PID: %d - Cache Miss - Pagina: %d\n", pid, numeroPagina)
}

func LogPaginaIngresadaEnCache(pid int, numeroPagina int) {
	log.Printf("PID: %d - Cache Add - Pagina: %d\n", pid, numeroPagina)
}

func LogPaginaActualizadaDeCacheAMemoria(pid int, numeroPagina int, frameEnMemoriaPrincipal int) {
	log.Printf("PID: %d - Memory Update - Página: %d - Frame: %d\n", pid, numeroPagina, frameEnMemoriaPrincipal)
}
