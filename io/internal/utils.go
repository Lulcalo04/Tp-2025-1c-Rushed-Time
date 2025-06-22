package io_internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"globals"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

// & -------------------------------- Server de IO -------------------------------------//

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

// & ----------------------------------------------- Handlers ------------------------------------------------------------------//

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

	NotificarFinalizacionIO(paqueteKernel.PID, paqueteKernel.NombreDispositivo)
}

// & ----------------------------------------------- Peticiones ------------------------------------------------------------------//

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

func NotificarFinalizacionIO(pid int, nombreDispositivo string) {

	respuestaIO := globals.IOResponse{
		NombreDispositivo: nombreDispositivo,
		PID:               pid,
		Respuesta:         true,
	}

	body, err := json.Marshal(respuestaIO)
	if err != nil {
		Logger.Debug("Error codificando mensajes", "error", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/io/fin", Config_IO.IPKernel, Config_IO.PortKernel)

	Logger.Debug("Enviando respuesta al kernel", "nombre_dispositivo", respuestaIO.NombreDispositivo, "pid", respuestaIO.PID, "respuesta", respuestaIO.Respuesta)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		Logger.Debug("Error enviando mensajes", "ip", Config_IO.IPKernel, "puerto", Config_IO.PortKernel)
	}

	Logger.Debug("Respuesta del servidor", "status", resp.Status)
}

func NotificarDesconexionDispositivo(nombreDispositivo string, ipInstancia string, puertoInstancia int) {

	RequestIO := globals.IOtoKernelDesconexionRequest{
		NombreDispositivo: nombreDispositivo,
		IpInstancia:       ipInstancia,
		PuertoInstancia:   puertoInstancia,
	}

	body, err := json.Marshal(RequestIO)
	if err != nil {
		Logger.Debug("Error codificando mensajes", "error", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/io/desconexion", Config_IO.IPKernel, Config_IO.PortKernel)

	Logger.Debug("Enviando Request a kernel")

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		Logger.Debug("Error enviando mensajes", "ip", Config_IO.IPKernel, "puerto", Config_IO.PortKernel)
	}

	Logger.Debug("Request del servidor", "status", resp.Status)
}

func EscucharSeñalDesconexion(nombreDispositivo string) {
	canalDeEscucha := make(chan os.Signal, 1)                    // Creamos un canal para escuchar señales
	signal.Notify(canalDeEscucha, os.Interrupt, syscall.SIGTERM) // Escucha señales de interrupción
	<-canalDeEscucha                                             // Espera a recibir una señal de interrupción (Ctrl+C o SIGTERM)

	// Antes de salir, notificamos al kernel de la desconexión

	NotificarDesconexionDispositivo(nombreDispositivo, Config_IO.IPIo, Config_IO.PortIO)

	Logger.Debug("Recibido Ctrl+C, desconectando del kernel...")
	os.Exit(0)
}
