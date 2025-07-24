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
	mensajeHandshakeMemoria := fmt.Sprintf("Iniciando Handshake con Memoria en %s:%d", ip, puerto)
	fmt.Println(mensajeHandshakeMemoria)
	Logger.Debug(mensajeHandshakeMemoria)

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

	mensajeHandshakeMemoriaExitoso := fmt.Sprintf("Handshake exitoso con Memoria: IP %s, Puerto %d, Respuesta: %t", respuesta.Ip, respuesta.Port, respuesta.Respuesta)
	fmt.Println(mensajeHandshakeMemoriaExitoso)
	Logger.Debug(mensajeHandshakeMemoriaExitoso)
}

func PingCon(nombre, ip string, puerto int) bool {
	url := fmt.Sprintf("http://%s:%d/ping", ip, puerto)

	// Realizamos la petición GET al servidor
	resp, err := http.Get(url)
	if err != nil {
		Logger.Debug(fmt.Sprintf("Error conectando con %s", nombre), "error", err)
		return false // Devuelve false si hay un error de conexión
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Validar el código de estado antes de intentar decodificar
	if resp.StatusCode != http.StatusOK {
		Logger.Debug(fmt.Sprintf("Error: StatusCode no es 200 OK para %s", nombre), "status_code", resp.StatusCode)
		return false
	}

	// Decodifico la respuesta JSON del servidor
	var respuesta globals.PingResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		Logger.Debug(fmt.Sprintf("Error decodificando respuesta JSON de %s", nombre), "error", err)
		return false
	}

	mensajePingExitoso := fmt.Sprintf("Ping exitoso a %s, respuesta: %s", nombre, respuesta.Mensaje)
	fmt.Println(mensajePingExitoso)
	Logger.Debug(mensajePingExitoso)

	return true
}

func PedirEspacioAMemoria(pcbDelProceso globals.PCB) bool {
	url := fmt.Sprintf("http://%s:%d/kernel/espacio/pedir", Config_Kernel.IPMemory, Config_Kernel.PortMemory)

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

	if respuestaMemoria.Respuesta {
		mensajeEspacioConcedido := fmt.Sprintf("Espacio en memoria concedido para el PID %d: %s", pcbDelProceso.PID, respuestaMemoria.Mensaje)
		fmt.Println(mensajeEspacioConcedido)
		Logger.Debug(mensajeEspacioConcedido)
		return true
	} else {
		fmt.Printf("Espacio en memoria NO concedido para el PID %d: %s\n", pcbDelProceso.PID, respuestaMemoria.Mensaje)
		Logger.Debug("Espacio en memoria NO concedido", "mensaje", respuestaMemoria.Mensaje, "pid", pcbDelProceso.PID)
		return false
	}

}

func LiberarProcesoEnMemoria(pid int) bool {

	// Declaro la URL a la que me voy a conectar (handler de liberación de memoria con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/kernel/espacio/liberar", Config_Kernel.IPMemory, Config_Kernel.PortMemory)

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
		mensajeLiberacionExitosa := fmt.Sprintf("Liberación de memoria exitosa para el PID %d", pid)
		Logger.Debug(mensajeLiberacionExitosa)
		fmt.Println(mensajeLiberacionExitosa)

		return true
	} else {
		mensajeLiberacionFallida := fmt.Sprintf("Liberación de memoria fallida para el PID %d", pid)
		Logger.Debug(mensajeLiberacionFallida)
		fmt.Println(mensajeLiberacionFallida)

		return false
	}
}

func PedirDumpMemory(pid int) bool {
	// Declaro la URL a la que me voy a conectar (handler de liberación de memoria con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/kernel/dumpMemory", Config_Kernel.IPMemory, Config_Kernel.PortMemory)

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

func PedirSwapping(pid int) bool {
	// Declaro la URL a la que me voy a conectar (handler de swappeo con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/kernel/espacio/entrarASwap", Config_Kernel.IPMemory, Config_Kernel.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.SwappingRequest{
		PID: pid,
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

	var respuestaMemoria globals.SwappingResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return false
	}

	if respuestaMemoria.Respuesta {
		Logger.Debug("Swapping Exitoso", "PID", pid)
		return true
	} else {
		Logger.Debug("No se pudo hacer el Swapping", "PID", pid)
		return false
	}
}

func PedirLiberacionDeSwap(pid int) bool {
	// Declaro la URL a la que me voy a conectar (handler de liberación de swap con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/kernel/espacio/volverDeSwap", Config_Kernel.IPMemory, Config_Kernel.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.SwappingRequest{
		PID: pid,
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

	var respuestaMemoria globals.SwappingResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return false
	}

	if respuestaMemoria.Respuesta {
		Logger.Debug("Liberación de Swap Exitosa", "PID", pid)
		fmt.Println("Liberación de Swap Exitosa", "PID", pid)
		return true
	} else {
		Logger.Debug("No se pudo liberar el Swap", "PID", pid)
		fmt.Println("No se pudo liberar el Swap", "PID", pid)
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
	bodyPeticion, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return
	}

	// Realizamos la petición en un goroutine para no bloquear
	go func() {
		respuestaIo, err := http.Post(url, "application/json", bytes.NewBuffer(bodyPeticion))
		if err != nil {
			Logger.Debug("Error conectando con IO", "error", err)
			return
		}
		defer respuestaIo.Body.Close()

		// Validar el código de estado
		if respuestaIo.StatusCode != http.StatusOK {
			Logger.Debug("Error: StatusCode no es 200 OK", "status_code", respuestaIo.StatusCode)
			return
		}

		mensajePeticionEnviada := fmt.Sprintf("Petición enviada a IO con éxito para el PID %d en el dispositivo %s con puerto %d", pid, instanciaDeIO.NombreIO, instanciaDeIO.PortIO)
		Logger.Debug(mensajePeticionEnviada)
		fmt.Println(mensajePeticionEnviada)

	}()
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

func PeticionDesalojo(pid int, motivoDesalojo string) bool {

	MutexIdentificadoresCPU.Lock()
	cpuDelPID := BuscarCPUporPID(pid)
	if cpuDelPID == nil {
		mensajeCpuNoEncontrada := fmt.Sprintf("No se encontró la CPU que tiene el PID %d, no se puede desalojar", pid)
		Logger.Debug(mensajeCpuNoEncontrada)
		fmt.Println(mensajeCpuNoEncontrada)
		return false
	}
	MutexIdentificadoresCPU.Unlock()

	if !cpuDelPID.DesalojoSolicitado {
		cpuDelPID.DesalojoSolicitado = true
	} else {
		Logger.Debug("Ya se ha solicitado un desalojo para el PID", "pid", pid, "cpu_id", cpuDelPID.CPUID)
		return false
	}

	mensajeIntentoDesalojo := fmt.Sprintf("Intentando desalojar el PID %d de la CPU %s por motivo: %s", pid, cpuDelPID.CPUID, motivoDesalojo)
	Logger.Debug(mensajeIntentoDesalojo)
	fmt.Println(mensajeIntentoDesalojo)

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
		return false
	}
	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con CPU", "error", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		Logger.Debug("Error: StatusCode no es 200 OK", "status_code", resp.StatusCode)
		return false
	}

	// Decodifico la respuesta JSON del server
	var respuestaDesalojo globals.KerneltoCPUDesalojoResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaDesalojo); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return false
	}

	if respuestaDesalojo.Respuesta {
		mensajeDesalojoExitoso := fmt.Sprintf("Desalojo exitoso del PID %d de la CPU %s", pid, cpuDelPID.CPUID)
		Logger.Debug(mensajeDesalojoExitoso)
		fmt.Println(mensajeDesalojoExitoso)
		CortoNotifier <- struct{}{} // Notificamos al planificador de corto plazo que se ha desalojado un proceso
		return true
	} else {
		mensajeDesalojoFallido := fmt.Sprintf("Desalojo fallido del PID %d de la CPU %s", pid, cpuDelPID.CPUID)
		Logger.Debug(mensajeDesalojoFallido)
		fmt.Println(mensajeDesalojoFallido)
		return false
	}

}
