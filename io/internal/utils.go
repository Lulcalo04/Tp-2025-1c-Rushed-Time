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
	mux.HandleFunc("/io/request", RecibirSolicitudIO)

	//Escucha el puerto y espera conexiones
	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

//-------------------------------------------------------------------------------------------------------------//

func RecibirSolicitudIO(w http.ResponseWriter, r *http.Request) {

	// Verificar que el metodo sea POST ya que es el unico valido que nos puede llegar
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	//Declaro la variable request de tipo IORequest
	var paqueteKernel globals.IORequest

	// Decodifico el request en la variable request, si no puedo decodificarlo, devuelvo un error
	if err := json.NewDecoder(r.Body).Decode(&paqueteKernel); err != nil {
		http.Error(w, "Error en el formato del request", http.StatusBadRequest)
		return
	}
	Logger.Debug("Recibiendo paquete desde Kernel", "nombre_dispositivo", paqueteKernel.NombreDispositivo, "pid", paqueteKernel.PID, "tiempo", paqueteKernel.Tiempo)

	LogInicioIO(paqueteKernel.PID, paqueteKernel.Tiempo)

	time.Sleep(time.Millisecond * time.Duration(paqueteKernel.Tiempo))

	LogFinalizacionIO(paqueteKernel.PID)

	notificarFinalizacionIO(paqueteKernel.PID, paqueteKernel.NombreDispositivo)
}

// & -----------------------------------------------Peticiones------------------------------------------------------------------//

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

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		Logger.Debug("Error enviando mensajes", "ip", ipKernel, "puerto", puertoKernel)
	}
	Logger.Debug("Respuesta del servidor", "status", resp.Status)

}

func notificarFinalizacionIO(pid int, nombreDispositivo string) {

	respuestaIO := globals.IOResponse{
		NombreDispositivo: nombreDispositivo,
		PID:               pid,
		Respuesta:         true,
	}

	body, err := json.Marshal(respuestaIO)
	if err != nil {
		Logger.Debug("Error codificando mensajes", "error", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/fin/io", Config_IO.IPKernel, Config_IO.PortKernel)

	Logger.Debug("Enviando respuesta al kernel", "nombre_dispositivo", respuestaIO.NombreDispositivo, "pid", respuestaIO.PID, "respuesta", respuestaIO.Respuesta)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		Logger.Debug("Error enviando mensajes", "ip", Config_IO.IPKernel, "puerto", Config_IO.PortKernel)
	}

	Logger.Debug("Respuesta del servidor", "status", resp.Status)
}
