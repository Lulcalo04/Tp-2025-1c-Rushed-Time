package io_internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

type IORequest struct { // Estructura que representa el response que recibe el IO desde el kernel
	PID  int `json:"pid"`
	Time int `json:"time"`
}

type Paquete struct {
	IPio   string `json:"ip_io"`
	PortIO int    `json:"port_io"`
	Nombre string `json:"nombre"`
}

//-------------------------------------------------------------------------------------------------------------//

func VerificarNombreIO() string {

	if len(os.Args) < 2 {
		fmt.Println("Error, mal escrito usa: ./io.go [nombreio]")
		os.Exit(1)
	}
	ioName := os.Args[1]

	return ioName
}

//--------------------------------Server de conexion IO-Kernel-------------------------------------//

func IniciarServerIO(puerto int) {
	stringPuerto := fmt.Sprintf(":%d", puerto)

	mux := http.NewServeMux()

	mux.HandleFunc("/io/request", RecibirIOpaquete)

	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

//-------------------------------------------------------------------------------------------------------------//

func RecibirIOpaquete(w http.ResponseWriter, r *http.Request) {

	// Verificar que el mÃ©todo sea POST ya que es el unico valido que nos puede llegar
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

	// Envio la respuesta al cliente, en este caso el kernel diciendole que termino el IO S
	w.Header().Set("Content-Type", "application/json")

	// Envio el PID y el tiempo que duro el IO
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"pid": %d}`, request.PID)

	//escribo la informacion en el body http de la respuesta
	w.Write([]byte(response))

}

//------------------------------------------------------------------------------------------------------------------//

//------------------------------------------------------------------------------------------------------------------//

func Conexion_Kernel(ip_kernel string, puerto_kernel int, nombre string, ip_IO string, puertoIO int) {

	paquete := Paquete{
		IPio:   ip_IO,
		PortIO: puertoIO,
		Nombre: nombre,
	}
	//Aca estamos armando un paquete con el nombre del IO y la ip y puerto del IO

	log.Printf("Paquete a enviar: %+v", paquete)

	EnviarPaqueteKernel(ip_kernel, puerto_kernel, paquete)

}

//-----------------------------------------------------------------------------------------------------------------//

func EnviarPaqueteKernel(ip string, puerto int, paquete Paquete) {

	body, err := json.Marshal(paquete)
	if err != nil {
		log.Printf("Error codificando mensajes: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/ConexionIOKernel", ip, puerto)

	//Modificamos la url para que sea /COnexionIOKernel
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error enviando mensajes a ip:%s puerto:%d", ip, puerto)
	}
	log.Printf("Respuesta del servidor: %s", resp.Status)
}