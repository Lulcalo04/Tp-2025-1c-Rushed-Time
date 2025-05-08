package kernel_internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"utils/client"
	"utils/globals"
)

// &-------------------------------------------Funcion para iniciar Server de Kernel-------------------------------------------------------------

func IniciarServerKernel(puerto int) {
	//Transformo el puerto a string
	stringPuerto := fmt.Sprintf(":%d", puerto)

	//Declaro el server
	mux := http.NewServeMux()

	//Declaro los handlers para el server
	mux.HandleFunc("/handshake", HandshakeHandler)
	mux.HandleFunc("/ping", PingHandler)
	mux.HandleFunc("/desalojo", DesalojoHandler)

	//Escucha el puerto y espera conexiones
	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

// &-------------------------------------------Endpoints de Kernel-------------------------------------------------------------

// * Endpoint de handshake = /handshake
func HandshakeHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	//!Falta validar en el cliente si la es un JSON o no
	w.Header().Set("Content-Type", "application/json")

	//Se utiliza el encoder para enviar la respuesta en formato JSON
	json.NewEncoder(w).Encode(
		map[string]string{
			"modulo":  "Kernel",
			"mensaje": "Conexión aceptada desde " + r.RemoteAddr,
		})
}

// * Endpoint de ping = /ping
func PingHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	//!Falta validar en el cliente si la es un JSON o no
	w.Header().Set("Content-Type", "application/json")

	//Se utiliza el encoder para enviar la respuesta en formato JSON
	var respuestaPing = client.PingResponse{
		Modulo:  "Kernel",
		Mensaje: "Pong",
	}

	json.NewEncoder(w).Encode(respuestaPing)
}

// * Endpoint de desalojo = /desalojo
func DesalojoHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var PeticionDesalojo globals.DesalojoRequest
	if err := json.NewDecoder(r.Body).Decode(&PeticionDesalojo); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	// Verifica si se desaloja por: Planificador (SJF CD), IO, o por fin de proceso
	// Dependiendo el motivo, se enviará el proceso a la cola correspondiente
	AnalizarDesalojo(PeticionDesalojo.PID, PeticionDesalojo.MotivoDesalojo)

	// Se replanifica para enviarle un proceso a CPU
	PlanificadorCortoPlazo(Config_Kernel.SchedulerAlgorithm)
}
