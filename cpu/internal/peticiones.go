package cpu_internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"utils/globals"
)

func HandshakeConKernel(cpuid string) bool {

	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/handshake/cpu", Config_CPU.IPKernel, Config_CPU.PortKernel)

	// Declaro el body de la petición
	pedidoBody := globals.CPUHandshakeRequest{
		CPUID:  cpuid,
		Puerto: Config_CPU.PortCPU,
		Ip:     Config_CPU.IpCpu,
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
		Logger.Debug("Error conectando con Kernel", "error", err)
		return false
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuestaKernel globals.CPUHandshakeResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaKernel); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return false
	}

	if !respuestaKernel.Respuesta {
		Logger.Debug("Error en el handshake con Kernel, Motivo: ", "mensaje", respuestaKernel.Mensaje)
	}
	//Devolvemos el bool para confirmar si el handshake fue exitoso
	return respuestaKernel.Respuesta
}

func SolicitarSiguienteInstruccionMemoria(pid int, pc int) {

	// Declaro la URL a la que me voy a conectar (handler de handshake con el puerto del server)
	url := fmt.Sprintf("http://%s:%d/instruccion", Config_CPU.IPMemory, Config_CPU.PortMemory)

	// Declaro el body de la petición
	pedidoBody := globals.InstruccionAMemoriaRequest{
		PID: pid,
		PC:  pc,
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
		Logger.Debug("Error conectando con Memoria", "error", err)
		return
	}
	defer resp.Body.Close() // Cierra la conexión al finalizar la función

	// Decodifico la respuesta JSON del server
	var respuestaMemoria globals.InstruccionAMemoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respuestaMemoria); err != nil {
		Logger.Debug("Error decodificando respuesta JSON", "error", err)
		return
	}

	ProcesoEjecutando.InstruccionActual = respuestaMemoria.InstruccionAEjecutar
}

//! No estamos seguros de que es lo que tenemos que enviarle a memoria y que es lo que nos va a devolver
// func solicitarInfoMMU () {
//}
