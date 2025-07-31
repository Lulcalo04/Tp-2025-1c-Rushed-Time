package cpu_internal

import (
	"fmt"
	"globals"
	"log/slog"
	"os"
)

/*
Fetch Instrucción: “## PID: <PID> - FETCH - Program Counter: <PROGRAM_COUNTER>”.
Interrupción Recibida: “## Llega interrupción al puerto Interrupt”.
Instrucción Ejecutada: “## PID: <PID> - Ejecutando: <INSTRUCCION> - <PARAMETROS>”.
Lectura/Escritura Memoria: “PID: <PID> - Acción: <LEER / ESCRIBIR> - Dirección Física: <DIRECCION_FISICA> - Valor: <VALOR LEIDO / ESCRITO>”.
Obtener Marco: “PID: <PID> - OBTENER MARCO - Página: <NUMERO_PAGINA> - Marco: <NUMERO_MARCO>”.
TLB Hit: “PID: <PID> - TLB HIT - Pagina: <NUMERO_PAGINA>”
TLB Miss: “PID: <PID> - TLB MISS - Pagina: <NUMERO_PAGINA>”
Página encontrada en Caché: “PID: <PID> - Cache Hit - Pagina: <NUMERO_PAGINA>”
Página faltante en Caché: “PID: <PID> - Cache Miss - Pagina: <NUMERO_PAGINA>”
Página ingresada en Caché: “PID: <PID> - Cache Add - Pagina: <NUMERO_PAGINA>”
Página Actualizada de Caché a Memoria: “PID: <PID> - Memory Update - Página: <NUMERO_PAGINA> - Frame: <FRAME_EN_MEMORIA_PRINCIPAL>”
*/

func ConfigurarLoggerCPU(cpuId string, logLevelModulo string) *slog.Logger {

	// Definimos la ruta del log
	rutaDelLog := "cpu" + "/" + cpuId + ".log"
	// Definimos el nivel de log
	nivel := globals.PasarStringALogLevel(logLevelModulo)

	logFile, err := os.OpenFile(rutaDelLog, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	handler := slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: nivel})

	// Guardamos la instancia del logger en una variable global o accesible por el módulo
	Logger := slog.New(handler)

	return Logger
}

func LogFetchInstruccion(pid int, programCounter int) {
	mensaje := fmt.Sprintf("## PID: %d - FETCH - Program Counter: %d", pid, programCounter)
	Logger.Info(mensaje)
	fmt.Println(mensaje)
}

func LogInterrupcionRecibida() {
	mensaje := "## Llega interrupción al puerto Interrupt"
	Logger.Info(mensaje)
	fmt.Println(mensaje)
}

func LogInstruccionEjecutada(pid int, instruccion string, parametros string) {
	mensaje := fmt.Sprintf("## PID: %d - Ejecutando: %s - %s", pid, instruccion, parametros)
	Logger.Info(mensaje)
	fmt.Println(mensaje)
}

func LogLecturaEscrituraMemoria(pid int, accion string, direccionFisica int, valor string) {
	mensaje := fmt.Sprintf("PID: %d - Acción: %s - Dirección Física: %d - Valor: %s", pid, accion, direccionFisica, valor)
	Logger.Info(mensaje)
	fmt.Println(mensaje)
}

func LogObtenerMarco(pid int, numeroPagina int, numeroMarco int) {
	mensaje := fmt.Sprintf("PID: %d - OBTENER MARCO - Página: %d - Marco: %d", pid, numeroPagina, numeroMarco)
	Logger.Info(mensaje)
	fmt.Println(mensaje)
}

func LogTLBHit(pid int, numeroPagina int) {
	mensaje := fmt.Sprintf("PID: %d - TLB HIT - Pagina: %d", pid, numeroPagina)
	Logger.Info(mensaje)
	fmt.Println(mensaje)
}

func LogTLBMiss(pid int, numeroPagina int) {
	mensaje := fmt.Sprintf("PID: %d - TLB MISS - Pagina: %d", pid, numeroPagina)
	Logger.Info(mensaje)
	fmt.Println(mensaje)
}

func LogPaginaEncontradaEnCache(pid int, numeroPagina int) {
	mensaje := fmt.Sprintf("PID: %d - Cache Hit - Pagina: %d", pid, numeroPagina)
	Logger.Info(mensaje)
	fmt.Println(mensaje)
}

func LogPaginaFaltanteEnCache(pid int, numeroPagina int) {
	mensaje := fmt.Sprintf("PID: %d - Cache Miss - Pagina: %d", pid, numeroPagina)
	Logger.Info(mensaje)
	fmt.Println(mensaje)
}

func LogPaginaIngresadaEnCache(pid int, numeroPagina int) {
	mensaje := fmt.Sprintf("PID: %d - Cache Add - Pagina: %d", pid, numeroPagina)
	Logger.Info(mensaje)
	fmt.Println(mensaje)
}

func LogPaginaActualizadaDeCacheAMemoria(pid int, numeroPagina int, frameEnMemoriaPrincipal int) {
	mensaje := fmt.Sprintf("PID: %d - Memory Update - Página: %d - Frame: %d", pid, numeroPagina, frameEnMemoriaPrincipal)
	Logger.Info(mensaje)
	fmt.Println(mensaje)
}
