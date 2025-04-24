package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// &------------------------------------------------------------------------------------------------------------------------
type HandshakeResponse struct {
	Modulo  string `json:"modulo"`
	Mensaje string `json:"mensaje"`
}

type PingResponse struct {
	Modulo  string `json:"modulo"`
	Mensaje string `json:"mensaje"`
}

func HandshakeCon(nombre string, ip string, puerto int) {
	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/handshake", ip, puerto)

	// Hacemos la petición GET al server
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error conectando con %s: %v", nombre, err)
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuesta HandshakeResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		log.Fatalf("Error decodificando respuesta JSON: %v", err)
	}

	log.Printf("Handshake con %s exitoso: módulo = %s, mensaje = %s", nombre, respuesta.Modulo, respuesta.Mensaje)
}

func PingCon(nombre string, ip string, puerto int) (respuestaPing bool) {
	// Variable para guardar la respuesta del ping
	respuestaPing = false

	// Declaro la URL a la que me voy a conectar (handler de ping con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/ping", ip, puerto)

	// Hacemos la petición GET al server
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error conectando con %s: %v", nombre, err)
		return respuestaPing // Devuelve false si hay un error de conexión
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuesta PingResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		log.Printf("Error decodificando respuesta JSON: %v", err)
		return respuestaPing // Devuelve false si hay un error al decodificar
	}

	log.Printf("Conexión con %s exitosa: %s / %s", nombre, resp.Status, respuesta.Mensaje)
	respuestaPing = true
	return respuestaPing
}

// &-------------------------------------------Funciones del TP0-----------------------------------------------------------------------------
type Mensaje struct {
	Mensaje string `json:"mensaje"`
}

type Paquete struct {
	Valores []string `json:"valores"`
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
	log.Printf("Paquete a enviar: %+v", paquete)

	EnviarPaquete(ip, puerto, paquete)
}

func EnviarMensaje(ip string, puerto int, mensajeTxt string) {
	mensaje := Mensaje{Mensaje: mensajeTxt}
	body, err := json.Marshal(mensaje)
	if err != nil {
		log.Printf("Error codificando mensaje: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/mensaje", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error enviando mensaje a ip:%s puerto:%d", ip, puerto)
	}

	log.Printf("Respuesta del servidor: %s", resp.Status)
}

func EnviarPaquete(ip string, puerto int, paquete Paquete) {
	body, err := json.Marshal(paquete)
	if err != nil {
		log.Printf("Error codificando mensajes: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/paquetes", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error enviando mensajes a ip:%s puerto:%d", ip, puerto)
	}

	log.Printf("Respuesta del servidor: %s", resp.Status)
}
