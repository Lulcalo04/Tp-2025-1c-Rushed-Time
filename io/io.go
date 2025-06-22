package main

import (
	"globals"
	io_internal "inputoutput/internal"
)

func main() {

	// Verificación del nombre del dispositivo IO
	ioName := io_internal.InicializarIO()

	// Inicializar configuración
	globals.IniciarConfiguracion("io/config.json", &io_internal.Config_IO)

	// Crear el logger
	io_internal.Logger = globals.ConfigurarLogger("io", io_internal.Config_IO.LogLevel)

	// PRIMERO levantar el servidor HTTP de IO en un hilo aparte para que no se bloquee el main
	go io_internal.IniciarServerIO(io_internal.Config_IO.PortIO)

	// Hacemos handshake con el kernel
	io_internal.HandshakeKernel(io_internal.Config_IO.IPKernel, io_internal.Config_IO.PortKernel, ioName)

	go io_internal.EscucharSeñalDesconexion(ioName)

	select {}
}
