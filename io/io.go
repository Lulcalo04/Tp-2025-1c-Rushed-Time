package main

import (
	"fmt"
	"globals"
	io_internal "inputoutput/internal"
)

func main() {

	// Verificación del nombre del dispositivo IO
	ioName := io_internal.InicializarIO()

	// Inicializar configuración
	fmt.Println("Inicializando configuración de IO...")
	globals.IniciarConfiguracion("io/config.json", &io_internal.Config_IO)

	// Crear el logger
	fmt.Println("Inicializando logger de IO...")
	io_internal.Logger = globals.ConfigurarLogger("io", io_internal.Config_IO.LogLevel)

	// PRIMERO levantar el servidor HTTP de IO en un hilo aparte para que no se bloquee el main
	fmt.Println("Iniciando servidor IO, en el puerto:", io_internal.Config_IO.PortIO)
	go io_internal.IniciarServerIO(io_internal.Config_IO.PortIO)

	// Hacemos handshake con el kernel
	fmt.Println("Haciendo handshake con el kernel...")
	io_internal.HandshakeConKernel(io_internal.Config_IO.IPKernel, io_internal.Config_IO.PortKernel, ioName)

	go io_internal.EscucharSeñalDesconexion(ioName)

	select {}
}
