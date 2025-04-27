package kernel_internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"utils/globals"
)

// &-------------------------------------------Config de Kernel-------------------------------------------------------------
type ConfigKernel struct {
	IPMemory           string `json:"ip_memory"`
	PortMemory         int    `json:"port_memory"`
	PortKernel         int    `json:"port_kernel"`
	SchedulerAlgorithm string `json:"scheduler_algorithm"`
	NewAlgorithm       string `json:"new_algorithm"`
	Alpha              string `json:"alpha"`
	SuspensionTime     int    `json:"suspension_time"`
	LogLevel           string `json:"log_level"`
}

var Config_Kernel *ConfigKernel

// &-------------------------------------------Funciones de Kernel-------------------------------------------------------------

var IsKernelRunning bool = false

func IniciarKernel() {
	IsKernelRunning = true
	IniciarPlanificador()
}

// &--------------------------------------------Funciones de Cliente-------------------------------------------------------------

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
		log.Printf("Error serializando JSON: %v", err)
		return false
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		log.Printf("Error conectando con Memoria: %v", err)
		return false
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuestaMemoria globals.PeticionMemoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		log.Printf("Error decodificando respuesta JSON: %v", err)
		return false
	}

	log.Printf(`Espacio en Memoria concedido: Petición de %s aceptada - Respuesta: %t - Mensaje: %s`,
		respuestaMemoria.Modulo, respuestaMemoria.Respuesta, respuestaMemoria.Mensaje)

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
		log.Printf("Error serializando JSON: %v", err)
		return false
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		log.Printf("Error conectando con Memoria: %v", err)
		return false
	}
	defer resp.Body.Close()

	// Validar el StatusCode ANTES de intentar leer el body
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error: StatusCode no es 200 OK, es %d", resp.StatusCode)
		return false
	}

	// Decodifico la respuesta JSON del server
	var respuestaMemoria globals.LiberacionMemoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		log.Printf("Error decodificando respuesta JSON: %v", err)
		return false
	}

	// Verificar el campo Respuesta en la respuesta
	if respuestaMemoria.Respuesta {
		log.Printf("Liberación de proceso en memoria exitosa para PID=%d", pid)
		return true
	} else {
		log.Printf("No se pudo liberar el proceso en memoria PID=%d", pid)
		return false
	}
}
