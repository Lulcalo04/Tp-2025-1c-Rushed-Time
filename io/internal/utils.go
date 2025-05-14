package io_internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
	"utils/globals"
)

type ConfigIO struct {
	IPKernel   string `json:"ip_kernel"`
	PortKernel int    `json:"port_kernel"`
	PortIO     int    `json:"port_io"`
	IPIo       string `json:"ip_io"`
	LogLevel   string `json:"log_level"`
}

var Config_IO *ConfigIO

var Logger *slog.Logger

type IORequest struct { // Estructura que representa el response que recibe el IO desde el kernel
	PID  int `json:"pid"`
	Time int `json:"time"`
}

type Paquete struct {
	IPio   string `json:"ip_io"`
	PortIO int    `json:"port_io"`
	Nombre string `json:"nombre"`
}

type DispositivoIO struct {
	NombreIO     string
	InstanciasIO int
}

var ListaDispositivosIO []DispositivoIO

//-------------------------------------------------------------------------------------------------------------//

func InicializarIO() string {
	if len(os.Args) < 2 {
		fmt.Println("Error, mal escrito usa: ./io.go [nombreio]")
		os.Exit(1)
	}
	ioName := os.Args[1]

	return ioName
}

//--------------------------------Server de conexion IO-Kernel-------------------------------------//

func IniciarServerIO(puerto int) {

	//Transformo el puerto a string
	stringPuerto := fmt.Sprintf(":%d", puerto)

	//Declaro el server
	mux := http.NewServeMux()

	//Declaro los handlers para el server
	mux.HandleFunc("/io/request", RecibirIOpaquete)

	//Escucha el puerto y espera conexiones
	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

//-------------------------------------------------------------------------------------------------------------//

func RecibirIOpaquete(w http.ResponseWriter, r *http.Request) {

	// Verificar que el metodo sea POST ya que es el unico valido que nos puede llegar
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	//Declaro la variable request de tipo IORequest
	var request IORequest

	// Decodifico el request en la variable request, si no puedo decodificarlo, devuelvo un error
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error en el formato del request", http.StatusBadRequest)
		return
	}

	LogInicioIO(request.PID, request.Time)

	time.Sleep(time.Millisecond * time.Duration(request.Time))

	LogFinalizacionIO(request.PID)

	// Envio la respuesta al cliente, en este caso el kernel diciendole que termino el IO
	w.Header().Set("Content-Type", "application/json")

	// Envio el PID y el tiempo que duro el IO
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"pid": %d}`, request.PID)

	//escribo la informacion en el body http de la respuesta
	w.Write([]byte(response))

}

//------------------------------------------------------------------------------------------------------------------//

func HandshakeKernel(ipKernel string, puertoKernel int, nombreIO string) {

	//Aca estamos armando un paquete con el nombre del IO y la ip y puerto del IO
	paquete := globals.IoHandshakeRequest{
		IPio:   Config_IO.IPIo,
		PortIO: Config_IO.PortIO,
		Nombre: nombreIO,
	}

	body, err := json.Marshal(paquete)
	if err != nil {
		Logger.Debug("Error codificando mensajes", "error", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/handshake/io", ipKernel, puertoKernel)

	//Modificamos la url para que sea /handshake/io
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		Logger.Debug("Error enviando mensajes", "ip", ipKernel, "puerto", puertoKernel)
	}
	Logger.Debug("Respuesta del servidor", "status", resp.Status)

}

//-----------------------------------------------------------------------------------------------------------------//
