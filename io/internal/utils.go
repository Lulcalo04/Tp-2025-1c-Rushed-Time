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

	// Declaro la variable request de tipo IORequest
	var paqueteKernel globals.IORequest

	// Decodifico el request en la variable request, si no puedo decodificarlo, devuelvo un error
	if err := json.NewDecoder(r.Body).Decode(&paqueteKernel); err != nil {
		http.Error(w, "Error en el formato del request", http.StatusBadRequest)
		return
	}

	mensajeSolicitud := fmt.Sprintf("Solicitud de IO recibida: PID %d, Dispositivo %s, Tiempo %d ms", paqueteKernel.PID, paqueteKernel.NombreDispositivo, paqueteKernel.Tiempo)
	fmt.Println(mensajeSolicitud)
	Logger.Debug(mensajeSolicitud)

	// Responder al Kernel inmediatamente para que pueda continuar
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Solicitud recibida y en proceso"))

	// Continuar con el procesamiento de IO en segundo plano
	LogInicioIO(paqueteKernel.PID, paqueteKernel.Tiempo)

	time.Sleep(time.Millisecond * time.Duration(paqueteKernel.Tiempo))

	mensajeFinIO := fmt.Sprintf("Finalizando IO: PID %d, Dispositivo %s, Tiempo: %d ms", paqueteKernel.PID, paqueteKernel.NombreDispositivo, paqueteKernel.Tiempo)
	fmt.Println(mensajeFinIO)
	Logger.Debug(mensajeFinIO)

	LogFinalizacionIO(paqueteKernel.PID)

	NotificarFinalizacionIO(paqueteKernel.PID, paqueteKernel.NombreDispositivo)
}

// & ----------------------------------------------- Peticiones ------------------------------------------------------------------//

func HandshakeConKernel(ipKernel string, puertoKernel int, nombreIO string) {

	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/handshake/io", ipKernel, puertoKernel)

	// Declaro el body de la petición
	pedidoBody := globals.IoHandshakeRequest{
		IPio:   Config_IO.IPIo,
		PortIO: Config_IO.PortIO,
		Nombre: nombreIO,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Kernel", "error", err)
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	// Log de éxito
	Logger.Debug("Handshake enviado al Kernel", "url", url, "body", string(bodyBytes))
	fmt.Println("Handshake con Kernel realizado con exito.")
}

func NotificarFinalizacionIO(pid int, nombreDispositivo string) {
	url := fmt.Sprintf("http://%s:%d/io/fin", Config_IO.IPKernel, Config_IO.PortKernel)

	mensajeFinIO := fmt.Sprintf("Notificando fin de IO: PID %d, Dispositivo %s", pid, nombreDispositivo)
	fmt.Println(mensajeFinIO)
	Logger.Debug(mensajeFinIO)

	respuestaIO := globals.IOResponse{
		NombreDispositivo: nombreDispositivo,
		PID:               pid,
		Respuesta:         true,
	}

	body, err := json.Marshal(respuestaIO)
	if err != nil {
		Logger.Debug("Error codificando mensajes", "error", err.Error())
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		Logger.Debug("Error enviando mensajes", "ip", Config_IO.IPKernel, "puerto", Config_IO.PortKernel)
	}

	Logger.Debug("Respuesta del servidor", "status", resp.Status)
}

func NotificarDesconexionDispositivo(nombreDispositivo string, ipInstancia string, puertoInstancia int) {
	url := fmt.Sprintf("http://%s:%d/io/desconexion", Config_IO.IPKernel, Config_IO.PortKernel)

	fmt.Println("Notificando desconexión de dispositivo al kernel.")
	RequestIO := globals.IOtoKernelDesconexionRequest{
		NombreDispositivo: nombreDispositivo,
		IpInstancia:       ipInstancia,
		PuertoInstancia:   puertoInstancia,
	}

	body, err := json.Marshal(RequestIO)
	if err != nil {
		Logger.Debug("Error codificando mensajes", "error", err.Error())
	}

	Logger.Debug("Enviando Request a kernel")

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		Logger.Debug("Error enviando mensajes", "ip", Config_IO.IPKernel, "puerto", Config_IO.PortKernel)
		return
	}
	if resp != nil {
		defer resp.Body.Close()
		Logger.Debug("Request del servidor", "status", resp.Status)
	} else {
		Logger.Debug("Error: respuesta del servidor es nil")
	}
}

func EscucharSeñalDesconexion(nombreDispositivo string) {

	canalDeEscucha := make(chan os.Signal, 1)                    // Creamos un canal para escuchar señales
	signal.Notify(canalDeEscucha, os.Interrupt, syscall.SIGTERM) // Escucha señales de interrupción
	<-canalDeEscucha                                             // Espera a recibir una señal de interrupción (Ctrl+C o SIGTERM)

	mensajeDesconexion := fmt.Sprintf("Recibido Ctrl+C, desconectando el dispositivo IO %s de KERNEL", nombreDispositivo)
	fmt.Println(mensajeDesconexion)
	Logger.Debug(mensajeDesconexion)

	NotificarDesconexionDispositivo(nombreDispositivo, Config_IO.IPIo, Config_IO.PortIO)

	os.Exit(0)
}
