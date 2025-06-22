package cpu_internal

import (
	"encoding/json"
	"fmt"
	"globals"
	"net/http"
)

func IniciarServerCPU(puerto int) {
	//Transformo el puerto a string
	stringPuerto := fmt.Sprintf(":%d", puerto)

	//Declaro el server
	mux := http.NewServeMux()

	//Declaro los handlers para el server
	mux.HandleFunc("/dispatch", DispatchHandler)
	mux.HandleFunc("/desalojo", DesalojoHandler)

	//Escucha el puerto y espera conexiones
	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

func DispatchHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var ProcesoRequest globals.ProcesoAEjecutarRequest
	if err := json.NewDecoder(r.Body).Decode(&ProcesoRequest); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	ProcesoEjecutando.PID = ProcesoRequest.PID
	ProcesoEjecutando.PC = ProcesoRequest.PC
	ProcesoEjecutando.Interrupt = false

	go CicloDeInstruccion()

}

func DesalojoHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var ProcesoRequest globals.DesalojoRequest
	if err := json.NewDecoder(r.Body).Decode(&ProcesoRequest); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	if ProcesoRequest.PID == ProcesoEjecutando.PID {
		ProcesoEjecutando.Interrupt = true
	}

	for ProcesoEjecutando.Interrupt {
		if InterrupcionAtendida {
			InterrupcionAtendida = false
			break
		}
	}

	var respuestaDesalojo = globals.DesalojoResponse{
		PID: ProcesoEjecutando.PID,
		PC:  ProcesoEjecutando.PC,
	}
	json.NewEncoder(w).Encode(respuestaDesalojo)

}
