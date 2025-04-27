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
	mux.HandleFunc("/espacio/pedir", PidenEspacioHandler)
	mux.HandleFunc("/espacio/liberar", LiberarEspacioHandler)

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

// * Endpoint de pedido de espacio = /espacio/pedir
func PidenEspacioHandler(w http.ResponseWriter, r *http.Request) {
	var pedidoRecibido globals.PeticionMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&pedidoRecibido); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Lógica para verificar espacio y reservarlo
	//! HAY QUE DESARROLLAR LA LOGICA DE RESERVA DE ESPACIO EN MEMORIA
	pedidoEnMemoria := true // Simulamos que se concede el espacio

	if pedidoEnMemoria {
		// Si el pedido es válido, se hace la concesión de espacio
		log.Printf("Solicitud a Memoria aceptada: PID=%d, Tamanio=%d", pedidoRecibido.ProcesoPCB.PID, pedidoRecibido.ProcesoPCB.TamanioEnMemoria)

		// Preparar respuesta y codificarla como JSON (se envia automaticamente a través del encode)
		resp := globals.PeticionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: true, // Simulamos que se concede el espacio
			Mensaje:   fmt.Sprintf("Espacio concedido para PID %d con %d de espacio", pedidoRecibido.ProcesoPCB.PID, pedidoRecibido.ProcesoPCB.TamanioEnMemoria),
		}
		json.NewEncoder(w).Encode(resp)
	} else {
		// Si el pedido no es válido, se envía un mensaje de error
		log.Printf("Solicitud a Memoria rechazada: PID=%d, Tamanio=%d", pedidoRecibido.ProcesoPCB.PID, pedidoRecibido.ProcesoPCB.TamanioEnMemoria)

		// Preparar respuesta y codificarla como JSON (se envia automaticamente a través del encode)
		resp := globals.PeticionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: false,
			Mensaje:   fmt.Sprintf("Espacio no concedido para PID %d con %d de espacio", pedidoRecibido.ProcesoPCB.PID, pedidoRecibido.ProcesoPCB.TamanioEnMemoria),
		}
		json.NewEncoder(w).Encode(resp)
	}

}

// * Endpoint de liberación de espacio = /espacio/liberar
func LiberarEspacioHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var pedidoRecibido globals.LiberacionMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&pedidoRecibido); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	liberacionDeMemoria := true // Simulación

	w.Header().Set("Content-Type", "application/json") // Setear Content-Type JSON

	if liberacionDeMemoria {
		log.Printf("Liberación de espacio aceptada: PID=%d", pedidoRecibido.PID)

		respuestaMemoria := globals.LiberacionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: true,
			Mensaje:   fmt.Sprintf("Liberación de espacio en memoria aceptada para PID %d", pedidoRecibido.PID),
		}

		w.WriteHeader(http.StatusOK) //Confirmar que la respuesta es 200 OK
		json.NewEncoder(w).Encode(respuestaMemoria)
	} else {
		log.Printf("Liberación de espacio fallida: PID=%d", pedidoRecibido.PID)

		respuestaMemoria := globals.LiberacionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: false,
			Mensaje:   fmt.Sprintf("Liberación de espacio en memoria para PID %d rechazada", pedidoRecibido.PID),
		}

		w.WriteHeader(http.StatusOK) // Siempre responder OK con JSON aunque falle la liberación
		json.NewEncoder(w).Encode(respuestaMemoria)
	}
}
