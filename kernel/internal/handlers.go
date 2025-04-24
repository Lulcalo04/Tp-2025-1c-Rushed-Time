package kernel_internal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func IniciarServerKernel(puerto int) {
	//Transformo el puerto a string
	stringPuerto := fmt.Sprintf(":%d", puerto)

	//Declaro el server
	mux := http.NewServeMux()

	//Declaro los handlers para el server
	mux.HandleFunc("/handshake", HandshakeHandler)

	//Escucha el puerto y espera conexiones
	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

// Endpoint de handshake = /handshake
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
