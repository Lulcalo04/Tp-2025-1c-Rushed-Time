package cpu_internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"globals"
	"net/http"
	"strconv"
)

func HandshakeConKernel(cpuid string) bool {

	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/handshake/cpu", Config_CPU.IPKernel, Config_CPU.PortKernel)

	// Declaro el body de la petición
	pedidoBody := globals.CPUToKernelHandshakeRequest{
		CPUID:  cpuid,
		Puerto: Config_CPU.PortCPU,
		Ip:     Config_CPU.IpCpu,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return false
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Kernel", "error", err)
		return false
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuestaKernel globals.CPUToKernelHandshakeResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaKernel); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return false
	}

	if !respuestaKernel.Respuesta {
		Logger.Debug("Error en el handshake con Kernel, Motivo: ", "mensaje", respuestaKernel.Mensaje)
	}
	//Devolvemos el bool para confirmar si el handshake fue exitoso
	return respuestaKernel.Respuesta
}

func HandshakeConMemoria(cpuid string) bool {
	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/handshake/cpu", Config_CPU.IPMemory, Config_CPU.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.CPUToMemoriaHandshakeRequest{
		CPUID: cpuid,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return false
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Memoria", "error", err)
		return false
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuestaMemoria globals.CPUToMemoriaHandshakeResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return false
	}

	EstructuraMemoriaDeCPU.EntradasPorTabla = respuestaMemoria.EntradasPorTabla
	EstructuraMemoriaDeCPU.NivelesDeTabla = respuestaMemoria.NivelesDeTabla
	EstructuraMemoriaDeCPU.TamanioMemoria = respuestaMemoria.TamanioMemoria
	EstructuraMemoriaDeCPU.TamanioPagina = respuestaMemoria.TamanioPagina
	//Devolvemos el bool para confirmar si el handshake fue exitoso
	return true

}

func SolicitarSiguienteInstruccionMemoria(pid int, pc int) {

	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/cpu/instruccion", Config_CPU.IPMemory, Config_CPU.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.InstruccionAMemoriaRequest{
		PID: pid,
		PC:  pc,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Memoria", "error", err)
		return
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuestaMemoria globals.InstruccionAMemoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	ProcesoEjecutando.InstruccionActual = respuestaMemoria.InstruccionAEjecutar
}

func PeticionFrameAMemoria(entradasPorNivel []int, pid int) int {

	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/cpu/frame", Config_CPU.IPMemory, Config_CPU.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.SolicitudFrameRequest{
		PID:              pid,
		EntradasPorNivel: entradasPorNivel,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return -1
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Memoria", "error", err)
		return -1
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	var respuestaMemoria globals.SolicitudFrameResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return -1
	}

	return respuestaMemoria.Frame

}

func PeticionWriteAMemoria(direccionFisica int, instruccion string, data string, pid int) {

	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/cpu/write", Config_CPU.IPMemory, Config_CPU.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.CPUWriteAMemoriaRequest{
		PID:             pid,
		Instruccion:     instruccion,
		DireccionFisica: direccionFisica,
		Data:            data,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Memoria", "error", err)
		return
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	var respuestaMemoria globals.CPUWriteAMemoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	if respuestaMemoria.Respuesta {
		LogLecturaEscrituraMemoria(pid, instruccion, direccionFisica, data)
	} else {
		Logger.Debug("Error al escribir en memoria", "pid", pid, "instruccion", instruccion, "direccion_fisica", direccionFisica, "data", data)
	}

}

func PeticionReadAMemoria(direccionFisica int, instruccion string, data string, pid int) {
	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/cpu/read", Config_CPU.IPMemory, Config_CPU.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.CPUReadAMemoriaRequest{
		PID:             pid,
		Instruccion:     instruccion,
		DireccionFisica: direccionFisica,
		Data:            data,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Memoria", "error", err)
		return
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	var respuestaMemoria globals.CPUReadAMemoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	if respuestaMemoria.Respuesta {
		//Logeo e imprimo lo leido de memoria
		LogLecturaEscrituraMemoria(pid, instruccion, direccionFisica, data)
		fmt.Printf("PID: %d - Acción: %s - Dirección Física: %d - Valor: %s", pid, instruccion, direccionFisica, data)
	} else {
		Logger.Debug("Error al leer de memoria", "pid", pid, "instruccion", instruccion, "direccion_fisica", direccionFisica, "data", data)
	}

}

func PeticionGotoAMemoria(direccionFisica int, instruccion string, pid int) {

	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/cpu/goto", Config_CPU.IPMemory, Config_CPU.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.CPUGotoAMemoriaRequest{
		PID:             pid,
		Instruccion:     instruccion,
		DireccionFisica: direccionFisica,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Memoria", "error", err)
		return
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	var respuestaMemoria globals.CPUGotoAMemoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	if respuestaMemoria.Respuesta {
		Logger.Debug("Goto exitoso", "pid", pid, "instruccion", instruccion, "direccion_fisica", direccionFisica)
	} else {
		Logger.Debug("Error Goto", "pid", pid, "instruccion", instruccion, "direccion_fisica", direccionFisica)
	}

}

func PeticionIOKernel(pid int, nombreDispositivo string, tiempo string) {

	tiempoInt, err := strconv.Atoi(tiempo)
	if err != nil {
		Logger.Error("Error al convertir direccionLogica a int", "error", err)
		return
	}

	// Declaro la URL a la que me voy a conectar (handler de IO con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/syscall/io", Config_CPU.IPKernel, Config_CPU.PortKernel)

	// Declaro el body de la petición
	pedidoBody := globals.IoSyscallRequest{
		PID:               pid,
		NombreDispositivo: nombreDispositivo,
		Tiempo:            tiempoInt,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Kernel", "error", err)
		return
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	var respuestaKernel globals.IoSyscallResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaKernel); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	if respuestaKernel.Respuesta {
		Logger.Debug("Petición IO exitosa", "pid", pid, "nombreDispositivo", nombreDispositivo, "tiempo", tiempo)
	} else {
		Logger.Debug("Error en la petición IO", "pid", pid, "nombreDispositivo", nombreDispositivo, "tiempo", tiempo)
	}

}

func PeticionInitProcKernel(pid int, nombreArchivo string, tamanio string) {

	tamanioInt, err := strconv.Atoi(tamanio)
	if err != nil {
		Logger.Error("Error al convertir tamanio a int", "error", err)
		return
	}

	// Declaro la URL a la que me voy a conectar (handler de IO con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/syscall/init_proc", Config_CPU.IPKernel, Config_CPU.PortKernel)

	// Declaro el body de la petición
	pedidoBody := globals.InitProcSyscallRequest{
		PID:           pid,
		NombreArchivo: nombreArchivo,
		Tamanio:       tamanioInt,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Kernel", "error", err)
		return
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	var respuestaKernel globals.InitProcSyscallResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaKernel); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	if respuestaKernel.Respuesta {
		Logger.Debug("Petición InitProc exitosa", "pid", pid, "nombre", nombreArchivo, "tamanio", tamanio)
	} else {
		Logger.Debug("Error en la petición InitProc", "pid", pid, "nombre", nombreArchivo, "tamanio", tamanio)
	}

}

func PeticionDumpMemoryKernel(pid int) {

	// Declaro la URL a la que me voy a conectar (handler de IO con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/syscall/dump_memory", Config_CPU.IPKernel, Config_CPU.PortKernel)

	// Declaro el body de la petición
	pedidoBody := globals.DumpMemorySyscallRequest{
		PID: pid,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Kernel", "error", err)
		return
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	var respuestaKernel globals.DumpMemorySyscallResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaKernel); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	if respuestaKernel.Respuesta {
		Logger.Debug("Petición DumpMemory exitosa", "pid", pid)
	} else {
		Logger.Debug("Error en la petición DumpMemory", "pid", pid)
	}

}

func PeticionExitKernel(pid int) {
	// Declaro la URL a la que me voy a conectar (handler de IO con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/syscall/exit", Config_CPU.IPKernel, Config_CPU.PortKernel)

	// Declaro el body de la petición
	pedidoBody := globals.ExitSyscallRequest{
		PID: pid,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Kernel", "error", err)
		return
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	var respuestaKernel globals.ExitSyscallResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaKernel); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	if respuestaKernel.Respuesta {
		Logger.Debug("Petición Exit exitosa", "pid", pid)
	} else {
		Logger.Debug("Error en la petición Exit", "pid", pid)
	}

}

func PeticionDesalojoKernel() {

	// Declaro la URL a la que me voy a conectar (handler de IO con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/cpu/desalojo", Config_CPU.IPKernel, Config_CPU.PortKernel)

	// Declaro el body de la petición
	pedidoBody := globals.CPUtoKernelDesalojoRequest{
		PID:    ProcesoEjecutando.PID,
		PC:     ProcesoEjecutando.PC,
		Motivo: ProcesoEjecutando.MotivoDesalojo,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Kernel", "error", err)
		return
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	var respuestaKernel globals.CPUtoKernelDesalojoResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaKernel); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	if respuestaKernel.Respuesta {
		Logger.Debug("Kernel aceptó el desalojo", "pid", ProcesoEjecutando.PID, "pc", ProcesoEjecutando.PC)
	} else {
		Logger.Debug("No se pudo desalojar el proceso", "pid", ProcesoEjecutando.PID, "pc", ProcesoEjecutando.PC)
	}
}
