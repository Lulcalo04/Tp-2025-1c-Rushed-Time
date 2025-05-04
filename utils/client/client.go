package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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

func HandshakeCon(nombre string, ip string, puerto int, Logger *slog.Logger) {
	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/handshake", ip, puerto)

	// Hacemos la petición GET al server
	resp, err := http.Get(url)
	if err != nil {
		Logger.Debug("Error conectando", "nombre", nombre, "error", err)
		os.Exit(1)
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuesta HandshakeResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		os.Exit(1)
	}

	Logger.Debug("Handshake exitoso",
		"nombre", nombre,
		"modulo", respuesta.Modulo,
		"mensaje", respuesta.Mensaje)
}

func PingCon(nombre string, ip string, puerto int, Logger *slog.Logger) (respuestaPing bool) {
	// Variable para guardar la respuesta del ping
	respuestaPing = false

	// Declaro la URL a la que me voy a conectar (handler de ping con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/ping", ip, puerto)

	// Hacemos la petición GET al server
	resp, err := http.Get(url)
	if err != nil {
		Logger.Debug("Error conectando con %s: %v", nombre, err)
		return respuestaPing // Devuelve false si hay un error de conexión
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuesta PingResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return respuestaPing // Devuelve false si hay un error al decodificar
	}

	Logger.Debug("Conexión exitosa",
		"nombre", nombre,
		"status", resp.Status,
		"mensaje", respuesta.Mensaje)

	respuestaPing = true
	return respuestaPing
}
