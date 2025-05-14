package cpu_internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"utils/globals"
)

func IniciarServerCPU(puerto int) {
	//Transformo el puerto a string
	stringPuerto := fmt.Sprintf(":%d", puerto)

	//Declaro el server
	mux := http.NewServeMux()

	//Declaro los handlers para el server
	mux.HandleFunc("/dispatch", RecibirProceso)

	//Escucha el puerto y espera conexiones
	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

func RecibirProceso(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var ProcesoRequest globals.ProcesoAEjecutarRequest
	if err := json.NewDecoder(r.Body).Decode(&ProcesoRequest); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	ProcesoEjecutando.PID = ProcesoRequest.PID
	ProcesoEjecutando.PC = ProcesoRequest.PC

}
