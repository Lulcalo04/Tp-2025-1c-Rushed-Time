package cpu_internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"utils/globals"
)

func HandshakeConKernel(nombre string, cpuid string) bool {

	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/handshake", Config_CPU.IPKernel, Config_CPU.PortKernel)

	// Declaro el body de la petición
	pedidoBody := globals.PeticionMemoriaRequest{
		Modulo:     "Kernel",
		ProcesoPCB: pcbDelProceso,
	}

	// Serializo el body a JSON
	bodyBytes, err := json.Marshal(pedidoBody)
	if err != nil {
		Logger.Debug("Error serializando JSON", "error", err)
		return false
	}

	// Hacemos la petición POST al server
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		Logger.Debug("Error conectando con Memoria", "error", err)
		return false
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuestaMemoria globals.PeticionMemoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return false
	}

	Logger.Debug("Espacio en Memoria concedido",
		"modulo", respuestaMemoria.Modulo,
		"respuesta", respuestaMemoria.Respuesta,
		"mensaje", respuestaMemoria.Mensaje)

	return true
}
