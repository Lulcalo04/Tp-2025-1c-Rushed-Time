package memoria_internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"utils/client"
	"utils/globals"
)

// -------------------------------------------Funcion para iniciar Server de Memoria-------------------------------------------------------------

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
	mux.HandleFunc("/instrucciones", InstruccionesHandler) //Escucha el puerto y espera conexiones
	mux.HandleFunc("/dump", DumpMemoryHandler)

	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

// -------------------------------------------Endpoints de Memoria-------------------------------------------------------------

// * Endpoint de handshake = /handshake
func HandshakeHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	//Se utiliza el encoder para enviar la respuesta en formato JSON
	json.NewEncoder(w).Encode(
		map[string]string{
			"modulo":  "Memoria",
			"mensaje": "Conexion aceptada desde " + r.RemoteAddr,
		})
}

// * Endpoint de ping = /ping
func PingHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
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
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// Logica para verificar espacio y reservarlo

	//! HAY QUE DESARROLLAR LA LOGICA DE RESERVA DE ESPACIO EN MEMORIA MAS ADELANTE

	pedidoEnMemoria := true //! Simulamos que se concede el espacio (checkpoint 2)

	if pedidoEnMemoria {
		// Si el pedido es valido, se hace la concesion de espacio
		Logger.Debug("Solicitud a Memoria aceptada", "PID", pedidoRecibido.ProcesoPCB.PID, "Tamanio", pedidoRecibido.ProcesoPCB.TamanioEnMemoria)

		// Preparar respuesta y codificarla como JSON (se envia automaticamente a traves del encode)
		resp := globals.PeticionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: true, // Simulamos que se concede el espacio
			Mensaje:   fmt.Sprintf("Espacio concedido para PID %d con %d de espacio", pedidoRecibido.ProcesoPCB.PID, pedidoRecibido.ProcesoPCB.TamanioEnMemoria),
		}
		json.NewEncoder(w).Encode(resp)
	} else {
		// Si el pedido no es valido, se envia un mensaje de error
		Logger.Debug("Solicitud a Memoria rechazada", "PID", pedidoRecibido.ProcesoPCB.PID, "Tamanio", pedidoRecibido.ProcesoPCB.TamanioEnMemoria)

		// Preparar respuesta y codificarla como JSON (se envia automaticamente a traves del encode)
		resp := globals.PeticionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: false,
			Mensaje:   fmt.Sprintf("Espacio no concedido para PID %d con %d de espacio", pedidoRecibido.ProcesoPCB.PID, pedidoRecibido.ProcesoPCB.TamanioEnMemoria),
		}
		json.NewEncoder(w).Encode(resp)
	}

}

// * Endpoint de liberacion de espacio = /espacio/liberar
func LiberarEspacioHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	var pedidoRecibido globals.LiberacionMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&pedidoRecibido); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	// Logica para liberar espacio en memoria

	liberacionDeMemoria := true // Simulacion

	w.Header().Set("Content-Type", "application/json") // Setear Content-Type JSON (osea que la respuesta sera del tipo JSON)

	if liberacionDeMemoria {
		Logger.Debug("Liberacion de espacio aceptada", "PID", pedidoRecibido.PID)

		respuestaMemoria := globals.LiberacionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: true,
			Mensaje:   fmt.Sprintf("Liberacion de espacio en memoria aceptada para PID %d", pedidoRecibido.PID),
		}

		w.WriteHeader(http.StatusOK) //Confirmar que la respuesta es 200 OK
		json.NewEncoder(w).Encode(respuestaMemoria)
	} else {
		Logger.Debug("Liberacion de espacio fallida", "PID", pedidoRecibido.PID)

		respuestaMemoria := globals.LiberacionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: false,
			Mensaje:   fmt.Sprintf("Liberacion de espacio en memoria para PID %d rechazada", pedidoRecibido.PID),
		}

		w.WriteHeader(http.StatusOK) // Siempre responder OK con JSON aunque falle la liberacion
		json.NewEncoder(w).Encode(respuestaMemoria)
	}
}

// * Endpoint de pedido de espacio = /dump
func DumpMemoryHandler(w http.ResponseWriter, r *http.Request) {
	var pedidoRecibido globals.DumpMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&pedidoRecibido); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// Logica para verificar espacio y reservarlo

	//! HAY QUE DESARROLLAR LA LOGICA DE DUMPEO DEL PROECESO

	dumpMemoria := true //! Simulamos que dumpea (checkpoint 2)

	if dumpMemoria {
		// Si el pedido es valido, se hace la concesion de espacio
		Logger.Debug("Solicitud de Dump Memory aceptada", "PID", pedidoRecibido.PID)

		// Preparar respuesta y codificarla como JSON (se envia automaticamente a traves del encode)
		resp := globals.DumpMemoryResponse{
			Modulo:    "Memoria",
			Respuesta: true, // Simulamos que se concede el espacio
			Mensaje:   fmt.Sprintf("Dump Memory concedido para PID %d", pedidoRecibido.PID),
		}
		json.NewEncoder(w).Encode(resp)
	} else {
		// Si el pedido no es valido, se envia un mensaje de error
		Logger.Debug("Solicitud de Dump Memory rechazada", "PID", pedidoRecibido.PID)

		// Preparar respuesta y codificarla como JSON (se envia automaticamente a traves del encode)
		resp := globals.DumpMemoryResponse{
			Modulo:    "Memoria",
			Respuesta: false,
			Mensaje:   fmt.Sprintf("Espacio no concedido para PID %d", pedidoRecibido.PID),
		}
		json.NewEncoder(w).Encode(resp)
	}

}

// -----------------------------------------------Funcion para instrucciones------------------------------------------------

// * Endpoint de instrucciones = /instrucciones
func InstruccionesHandler(w http.ResponseWriter, r *http.Request) {

	//verifica que el metodo sea POST ya que es el unico valido que nos puede llegar
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	//Declaro la variable request de tipo InstruccionesRequest para almacenar los datos enviadoos
	var req globals.InstruccionesRequest

	//Verifica que le hallamos mandado un JSON valido y decodifica el contenido en la variable request, si no puedo decodificarlo, devuelvo un error
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// Armamos el nombre del archivo: ejemplo "proceso_1.txt"
	filename := fmt.Sprintf("memoria/proceso_%d.txt", req.PID)

	//Verificamos si el archivo existe o se puede leer , si se puede content contendra los bytes del archivo
	content, err := os.ReadFile(filename)
	if err != nil {
		http.Error(w, "No se encontro el archivo del proceso", http.StatusNotFound)
		return
	}

	// Suponemos que el proceso1.txt contiene las instrucciones una por linea

	// string(content) convierte el proceso (que es un slice de bytes) a string
	// strings.TrimSpace elimina los espacios en blanco al principio y al final de la cadena para evitar saltos de linea
	// strings.Split divide la cadena en un slice de strings
	// lineas me devolvera, por ejemplo : []string{"Instruccion 1", "Instruccion 2", "Instruccion 3"}

	lineas := strings.Split(strings.TrimSpace(string(content)), "\n")

	// Preparamos la respuesta
	resp := globals.InstruccionesResponse{
		PID:           req.PID,
		Instrucciones: lineas,
	}

	//Establece el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	//Codifica el objeto resp en formato JSON y lo envia al cliente
	json.NewEncoder(w).Encode(resp)
}
