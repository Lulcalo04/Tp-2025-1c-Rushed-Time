package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Mensaje struct {
	Mensaje string `json:"mensaje"`
}

type Paquete struct {
	Valores []string `json:"valores"`
}

func IniciarConfiguracion(filePath string, config interface{}) {
	//var config *globals.Config

	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(config) // Se pasa el puntero directamente
	if err != nil {
		log.Fatalf("Error al decodificar la configuración: %s", err.Error())
	}
	if config == nil {
		log.Fatalf("No se pudo cargar la configuración")
	}

}

func LeerConsola() []string {
	// Leer de la consola hasta que se ingrese una linea vacia
	reader := bufio.NewReader(os.Stdin)
	log.Println("Ingrese los mensajes")

	var valores []string

	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text) //Le saca \n y espacios al texto

		if text == "" {
			log.Println("Fin de lectura por consola")
			break
		}

		valores = append(valores, text)
		log.Printf("Usted ingreso: %s", text)
	}

	return valores
}

func GenerarYEnviarPaquete(ip string, puerto int) {
	valores := LeerConsola()

	if len(valores) == 0 {
		log.Println("No se ingresaron valores. No se envía nada")
		return
	}

	paquete := Paquete{Valores: valores}
	log.Printf("paqute a enviar: %+v", paquete)

	EnviarPaquete(ip, puerto, paquete)
}

func EnviarMensaje(ip string, puerto int, mensajeTxt string) {
	mensaje := Mensaje{Mensaje: mensajeTxt}
	body, err := json.Marshal(mensaje)
	if err != nil {
		log.Printf("error codificando mensaje: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/mensaje", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando mensaje a ip:%s puerto:%d", ip, puerto)
	}

	log.Printf("respuesta del servidor: %s", resp.Status)
}

func EnviarPaquete(ip string, puerto int, paquete Paquete) {
	body, err := json.Marshal(paquete)
	if err != nil {
		log.Printf("error codificando mensajes: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/paquetes", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando mensajes a ip:%s puerto:%d", ip, puerto)
	}

	log.Printf("respuesta del servidor: %s", resp.Status)
}

func ConfigurarLogger(nombreDelModulo string) {

	rutaDelLog := nombreDelModulo + "/" + nombreDelModulo + ".log"

	logFile, err := os.OpenFile(rutaDelLog, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}
