package kernel_internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"globals"
	"net/http"
	"os"
)

// &--------------------------------------------Funciones de Cliente-------------------------------------------------------------

func HandshakeConMemoria(ip string, puerto int) {
	Logger.Debug("Iniciando Handshake con Memoria")
	fmt.Println("Iniciando Handshake con Memoria")

	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/handshake", ip, puerto)

	// Declaro el body de la petición
	pedidoBody := globals.HandshakeRequest{
		Modulo:    "Kernel",
		Ip:        Config_Kernel.IPKernel,
		Port:      Config_Kernel.PortKernel,
		Respuesta: false,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		os.Exit(1) // Termina el programa si hay un error al serializar
	}
	// Hago la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Memoria", "error", err)
		os.Exit(1) // Termina el programa si hay un error de conexión
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuesta globals.HandshakeRequest
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		os.Exit(1)
	}
	fmt.Println("Handshake exitoso")
	Logger.Debug("Handshake exitoso",
		"modulo", "Memoria",
		"ip", respuesta.Ip,
		"puerto", respuesta.Port,
		"respuesta", respuesta.Respuesta)
}

func PingCon(nombre string, ip string, puerto int) (respuestaPing bool) {
	// Variable para guardar la respuesta del ping
	respuestaPing = false

	// Declaro la URL a la que me voy a conectar (handler de ping con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/ping", ip, puerto)

	// Hacemos la petición GET al server
	resp, err := http.Get(url)
	if err != nil {
		Logger.Debug("Error conectando con %s: %v", nombre, err)
		return respuestaPing // Devuelve false si hay un error de conexión
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuesta globals.PingResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return respuestaPing // Devuelve false si hay un error al decodificar
	}

	Logger.Debug("Conexión exitosa",
		"nombre", nombre,
		"status", resp.Status,
		"mensaje", respuesta.Mensaje)

	respuestaPing = true
	return respuestaPing
}

func PedirEspacioAMemoria(pcbDelProceso globals.PCB) bool {

	// Declaro la URL a la que me voy a conectar (handler de Petición de memoria con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/espacio/pedir", Config_Kernel.IPMemory, Config_Kernel.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.PeticionMemoriaRequest{
		Modulo:     "Kernel",
		ProcesoPCB: pcbDelProceso,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return false
	}

	Logger.Debug("Enviando petición de espacio a Memoria", "url", url, "proceso", pcbDelProceso.PID)

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Memoria", "error", err, "url", url)
		return false
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	Logger.Debug("Respuesta de Memoria recibida", "status_code", resp.StatusCode, "proceso", pcbDelProceso.PID)
	// Decodifico la respuesta JSON del server
	var respuestaMemoria globals.PeticionMemoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err, "proceso", pcbDelProceso.PID)
		return false
	}
	Logger.Debug("Respuesta de Memoria procesada", "respuesta", respuestaMemoria.Respuesta, "mensaje", respuestaMemoria.Mensaje, "proceso", pcbDelProceso.PID)

	Logger.Debug("Espacio en Memoria concedido",
		"modulo", respuestaMemoria.Modulo,
		"respuesta", respuestaMemoria.Respuesta,
		"mensaje", respuestaMemoria.Mensaje)

	return true

}

func LiberarProcesoEnMemoria(pid int) bool {

	// Declaro la URL a la que me voy a conectar (handler de liberación de memoria con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/espacio/liberar", Config_Kernel.IPMemory, Config_Kernel.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.LiberacionMemoriaRequest{
		Modulo: "Kernel",
		PID:    pid,
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
	defer resp.Body.Close()

	// Validar el StatusCode ANTES de intentar leer el body
	if resp.StatusCode != http.StatusOK {
		Logger.Debug("Error: StatusCode no es 200 OK", "status_code", resp.StatusCode)
		return false
	}

	// Decodifico la respuesta JSON del server
	var respuestaMemoria globals.LiberacionMemoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return false
	}

	// Verificar el campo Respuesta en la respuesta
	if respuestaMemoria.Respuesta {
		Logger.Debug("Liberación de proceso en memoria exitosa", "PID", pid)
		return true
	} else {
		Logger.Debug("No se pudo liberar el proceso en memoria", "PID", pid)
		return false
	}
}

func PedirDumpMemory(pid int) bool {
	// Declaro la URL a la que me voy a conectar (handler de liberación de memoria con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/dump", Config_Kernel.IPMemory, Config_Kernel.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.DumpMemoryRequest{
		Modulo: "Kernel",
		PID:    pid,
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		Logger.Debug("Error: StatusCode no es 200 OK", "status_code", resp.StatusCode)
		return false
	}

	// Decodifico la respuesta JSON del server
	var respuestaMemoria globals.DumpMemoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return false
	}

	// Verificar el campo Respuesta en la respuesta
	if respuestaMemoria.Respuesta {
		Logger.Debug("Dump Memory Exitoso", "PID", pid)
		return true
	} else {
		Logger.Debug("No se pudo hacer el Dump Memory", "PID", pid)
		return false
	}

}

func EnviarProcesoAIO(instanciaDeIO InstanciaIO, pid int, milisegundosDeUso int) {
	url := fmt.Sprintf("http://%s:%d/io/request", instanciaDeIO.IpIO, instanciaDeIO.PortIO)

	// Declaro el body de la petición
	pedidoBody := globals.IORequest{
		NombreDispositivo: instanciaDeIO.NombreIO,
		PID:               pid,
		Tiempo:            milisegundosDeUso,
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
		Logger.Debug("Error conectando con IO", "error", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		Logger.Debug("Error: StatusCode no es 200 OK", "status_code", resp.StatusCode)
		return
	}

	Logger.Debug("Petición de IO enviada",
		"nombre_dispositivo", instanciaDeIO.NombreIO,
		"pid", pid,
		"tiempo", milisegundosDeUso)
}

func EnviarProcesoACPU(cpuIp string, cpuPuerto int, procesoPID int, procesoPC int) {

	// Declaro la URL a la que me voy a conectar (handler de Petición de memoria con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/dispatch", cpuIp, cpuPuerto)

	// Declaro el body de la petición
	pedidoBody := globals.ProcesoAEjecutarRequest{
		PID: procesoPID,
		PC:  procesoPC,
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
}

func PeticionDesalojo(pid int, motivoDesalojo string) {

	cpuDelPID := BuscarCPUporPID(pid)
	if cpuDelPID == nil {
		Logger.Debug("No se encontró el CPU que tenga este PID: ", "PID", pid)
		return
	}

	url := fmt.Sprintf("http://%s:%d/desalojo", cpuDelPID.Ip, cpuDelPID.Puerto)

	// Declaro el body de la petición
	pedidoBody := globals.KerneltoCPUDesalojoRequest{
		PID:    pid,
		Motivo: motivoDesalojo,
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
		Logger.Debug("Error conectando con CPU", "error", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		Logger.Debug("Error: StatusCode no es 200 OK", "status_code", resp.StatusCode)
		return
	}

	// Decodifico la respuesta JSON del server
	var respuestaDesalojo globals.KerneltoCPUDesalojoResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaDesalojo); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	if respuestaDesalojo.Respuesta {
		Logger.Debug("Desalojo exitoso")
	} else {
		Logger.Debug("Desalojo fallido")
	}

}
