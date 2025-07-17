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

	// Guardo el PID y PC del proceso que se va a ejecutar y reseteo iterrupt
	Logger.Debug("Empiezo a ejecutar el proceso", "PID", ProcesoRequest.PID, "PC", ProcesoRequest.PC)
	fmt.Printf("Empiezo a ejecutar el proceso PID %d en la CPU %s\n", ProcesoRequest.PID, CPUId)

	mutexProcesoEjecutando.Lock()
	ProcesoEjecutando.PID = ProcesoRequest.PID
	ProcesoEjecutando.PC = ProcesoRequest.PC
	ProcesoEjecutando.Interrupt = false
	mutexProcesoEjecutando.Unlock()

	// Levanto ciclo de instruccion

	mutexCicloDeInstruccion.Lock()
	go CicloDeInstruccion()

}

func DesalojoHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var ProcesoRequest globals.KerneltoCPUDesalojoRequest
	if err := json.NewDecoder(r.Body).Decode(&ProcesoRequest); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	// Si el PID es el correcto, se marca que hay q interrumpir el ciclo
	if ProcesoRequest.PID == ProcesoEjecutando.PID {
		Logger.Debug("Recibí una solicitud de desalojo", "PID", ProcesoRequest.PID, "Motivo", ProcesoRequest.Motivo)
		fmt.Printf("Recibí una solicitud de desalojo para el PID %d con motivo: %s\n", ProcesoRequest.PID, ProcesoRequest.Motivo)

		
		

		mutexProcesoEjecutando.Lock()
		ProcesoEjecutando.Interrupt = true
		ProcesoEjecutando.MotivoDesalojo = ProcesoRequest.Motivo
		mutexProcesoEjecutando.Unlock()

		LogInterrupcionRecibida()
	}

	var respuestaDesalojo = globals.KerneltoCPUDesalojoResponse{
		Respuesta: true,
	}
	json.NewEncoder(w).Encode(respuestaDesalojo)

}
