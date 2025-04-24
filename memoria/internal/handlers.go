package memoria_internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"utils/client"
	"utils/globals"
)

// &-------------------------------------------Funcion para iniciar Server de Memoria-------------------------------------------------------------

func IniciarServerMemoria(puerto int) {
	//Transformo el puerto a string
	stringPuerto := fmt.Sprintf(":%d", puerto)

	//Declaro el server
	mux := http.NewServeMux()

	//Declaro los handlers para el server
	mux.HandleFunc("/handshake", HandshakeHandler)
	mux.HandleFunc("/ping", PingHandler)
	mux.HandleFunc("/espacio", EspacioHandler)

	//Escucha el puerto y espera conexiones
	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

// &-------------------------------------------Endpoints de Memoria-------------------------------------------------------------

// * Endpoint de handshake = /handshake
func HandshakeHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	//!Falta validar en el cliente si la es un JSON o no
	w.Header().Set("Content-Type", "application/json")

	//Se utiliza el encoder para enviar la respuesta en formato JSON
	json.NewEncoder(w).Encode(
		map[string]string{
			"modulo":  "Memoria",
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
		Modulo:  "Memoria",
		Mensaje: "Pong",
	}

	if err := json.NewEncoder(w).Encode(respuestaPing); err != nil {
		http.Error(w, "Error al codificar JSON", http.StatusInternalServerError)
	}
}

// * Endpoint de espacio = /espacio
func EspacioHandler(w http.ResponseWriter, r *http.Request) {
	var req globals.PeticionMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Lógica para verificar espacio y reservarlo
	log.Printf("Solicitud a Memoria aceptada: PID=%d, Tamanio=%d", req.ProcesoPCB.PID, req.ProcesoPCB.TamanioEnMemoria)

	// Preparar respuesta
	resp := globals.PeticionMemoriaResponse{
		Modulo:    "Memoria",
		Respuesta: true, // Simulamos que se concede el espacio
		Mensaje:   fmt.Sprintf("Espacio concedido para PID %d con %d de espacio", req.ProcesoPCB.PID, req.ProcesoPCB.TamanioEnMemoria),
	}

	json.NewEncoder(w).Encode(resp)
}
