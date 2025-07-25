package cpu_internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"globals"
	"net/http"
	"strconv"
)

// &--------------------------------------------Funciones de Cliente-------------------------------------------------------------

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
	url := fmt.Sprintf("http://%s:%d/cpu/instrucciones", Config_CPU.IPMemory, Config_CPU.PortMemory)

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

	mutexProcesoEjecutando.Lock()
	ProcesoEjecutando.InstruccionActual = respuestaMemoria.InstruccionAEjecutar
	mutexProcesoEjecutando.Unlock()

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

func PeticionIOKernel(pid int, nombreDispositivo string, tiempo string) {

	tiempoInt, err := strconv.Atoi(tiempo)
	if err != nil {
		Logger.Debug("Error al convertir direccionLogica a int", "error", err)
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
		Logger.Debug("Error al convertir tamanio a int", "error", err)
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
	// Declaro la URL a la que me voy a conectar (handler de desalojo con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/cpu/desalojo", Config_CPU.IPKernel, Config_CPU.PortKernel)

	// Declaro el body de la petición
	mutexProcesoEjecutando.Lock()
	pedidoBody := globals.CPUtoKernelDesalojoRequest{
		PID:    ProcesoEjecutando.PID,
		PC:     ProcesoEjecutando.PC,
		Motivo: ProcesoEjecutando.MotivoDesalojo,
		CPUID:  CPUId,
	}
	mutexProcesoEjecutando.Unlock()

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
		mutexProcesoEjecutando.Lock()
		Logger.Debug("Se desalojo a ", "pid", ProcesoEjecutando.PID, "pc", ProcesoEjecutando.PC, "de la cpuid", CPUId, "por el motivo", ProcesoEjecutando.MotivoDesalojo)
		fmt.Println("Se desalojo a ", "pid", ProcesoEjecutando.PID, "pc", ProcesoEjecutando.PC, "de la cpuid", CPUId, "por el motivo", ProcesoEjecutando.MotivoDesalojo)
		mutexProcesoEjecutando.Unlock()
	} else {
		mutexProcesoEjecutando.Lock()
		Logger.Debug("No se pudo desalojar el proceso", "pid", ProcesoEjecutando.PID, "pc", ProcesoEjecutando.PC, "de la cpuid", CPUId)
		fmt.Println("No se pudo desalojar el proceso", ProcesoEjecutando.PID, "por el Kernel", "pc", ProcesoEjecutando.PC, "de la CPU", CPUId)
		mutexProcesoEjecutando.Unlock()
	}
}

func PedirPaginaAMemoria(pid int, direccionFisica int, numeroDePagina int) *EntradaCache {
	// Declaro la URL a la que me voy a conectar (handler de petición de página con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/cpu/pagina/pedir", Config_CPU.IPMemory, Config_CPU.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.CPUtoMemoriaPageRequest{
		PID:            pid,
		NumeroDePagina: numeroDePagina,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return nil
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Memoria", "error", err)
		return nil
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	var respuestaMemoria globals.MemoriaToCPUPageResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return nil
	}

	// Agregamos la página que nos devolvió Memoria a la Cache
	return AgregarPaginaEnCache(numeroDePagina, respuestaMemoria.ContenidoPagina, direccionFisica)
}

func EscribirEnPaginaMemoria(pid int, direccionFisica int, valor string) {

	valorEnBytes := []byte(valor)

	fmt.Println("Escribiendo en memoria: PID:", pid, "Dirección Física:", direccionFisica, "Valor:", valorEnBytes)
	Logger.Debug("Escribiendo en memoria", "pid", pid, "direccion_fisica", direccionFisica, "valor", valorEnBytes)

	// Declaro la URL a la que me voy a conectar (handler de petición de página con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/cpu/pagina/escribir", Config_CPU.IPMemory, Config_CPU.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.CPUWriteAMemoriaRequest{
		PID:             pid,
		DireccionFisica: direccionFisica,
		Data:            valorEnBytes,
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
		LogLecturaEscrituraMemoria(pid, "WRITE", direccionFisica, valor)
	} else {
		Logger.Debug("Error al escribir en memoria", "pid", pid, "direccion_fisica", direccionFisica, "data", valor)
	}

}

func LeerDePaginaMemoria(pid int, direccionFisica int, tamanio string) {

	tamanioInt, err := strconv.Atoi(tamanio)
	if err != nil {
		Logger.Debug("Error al convertir tamanio a int", "error", err)
		return
	}
	// Declaro la URL a la que me voy a conectar (handler de petición de página con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/cpu/pagina/leer", Config_CPU.IPMemory, Config_CPU.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.CPUReadAMemoriaRequest{
		PID:             pid,
		DireccionFisica: direccionFisica,
		Tamanio:         tamanioInt,
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
		LogLecturaEscrituraMemoria(pid, "READ", direccionFisica, string(respuestaMemoria.Data))
		fmt.Printf("PID: %d - Acción: READ - Dirección Física: %d - Valor: %s\n", pid, direccionFisica, string(respuestaMemoria.Data))
	} else {
		Logger.Debug("Error al leer de memoria", "pid", pid, "direccion_fisica", direccionFisica, "tamanio", tamanio)
	}
}

func ActualizarPaginaEnMemoria(pid int, numeroDePagina int, data []byte) {

	// Declaro la URL a la que me voy a conectar (handler de actualización de página con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/cpu/pagina/actualizar", Config_CPU.IPMemory, Config_CPU.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.CPUActualizarPaginaEnMemoriaRequest{
		PID:            pid,
		NumeroDePagina: numeroDePagina,
		Data:           data,
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

	var respuestaMemoria globals.CPUActualizarPaginaEnMemoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	if respuestaMemoria.Respuesta {
		LogPaginaActualizadaDeCacheAMemoria(pid, numeroDePagina, respuestaMemoria.Frame)
	} else {
		Logger.Debug("Error al actualizar página en memoria", "pid", pid, "numeroDePagina", numeroDePagina)
	}
}
