package main

import (
	"utils/client"
	/* "utils/client/globals" */
	"net/http"
	"utils/server"
)

func main() {
	client.ConfigurarLogger("cpu")

	//FUNCIONES DE SERVER
	mux := http.NewServeMux()

	mux.HandleFunc("/paquetes", server.RecibirPaquetes)
	mux.HandleFunc("/mensaje", server.RecibirMensaje)

	err := http.ListenAndServe(":8004", mux)
	if err != nil {
		panic(err)
	}
}
