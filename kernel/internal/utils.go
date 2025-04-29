package kernel_internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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

var IsKernelRunning bool = false

func IniciarKernel() {
	IsKernelRunning = true
	IniciarPlanificador()
}

func IniciarServerKernel(puerto int) {
	stringPuerto := fmt.Sprintf(":%d", puerto)

	mux := http.NewServeMux()

	mux.HandleFunc("/ConexionIOKernel", ConexionIOKernelHandler)

	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}

}

type PaqueteIO struct {
	IPio   string `json:"ip_io"`
	PortIO int    `json:"port_io"`
	Nombre string `json:"nombre"`
}

type IORequest struct {
	PID  int `json:"pid"`
	Time int `json:"time"`
}

func ConexionIOKernelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var paquete PaqueteIO
	if err := json.NewDecoder(r.Body).Decode(&paquete); err != nil {
		http.Error(w, "Error en el formato del JSON recibido", http.StatusBadRequest)
		return
	}

	log.Printf("Se registró un IO llamado '%s' en %s:%d", paquete.Nombre, paquete.IPio, paquete.PortIO)

	// Ahora Kernel se conecta al IO para enviarle el tiempo
	enviarTiempoAlIO(paquete.IPio, paquete.PortIO)

	// Le respondemos al IO simplemente para confirmar que lo registramos
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensaje": "IO registrado exitosamente"}`))
}

func enviarTiempoAlIO(ip string, puerto int) {
	url := fmt.Sprintf("http://%s:%d/io/request", ip, puerto)

	// Podés setear el tiempo como quieras. Por ahora ponemos un ejemplo fijo (1000ms = 1 segundo)
	request := IORequest{
		PID:  0,    // Podrías enviar un PID válido más adelante
		Time: 3000, // Ej: 3000 milisegundos de sleep
	}

	body, err := json.Marshal(request)
	if err != nil {
		log.Printf("Error serializando JSON de solicitud a IO: %v", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error conectando al IO para enviarle el tiempo: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Tiempo enviado al IO: %d ms - Estado respuesta: %s", request.Time, resp.Status)
}