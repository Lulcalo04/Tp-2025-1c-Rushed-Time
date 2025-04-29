package main

import (
	io_internal "inputoutput/internal"
	"time"
	"utils/globals"
	//"fmt"
	//"net"
	//"time"
)

func main() {

	// Verificación del nombre del dispositivo IO
	ioName := io_internal.VerificarNombreIO()

	// Crear el logger
	globals.ConfigurarLogger("io")

	// Inicializar configuración
	globals.IniciarConfiguracion("io/config.json", &io_internal.Config_IO)

	// PRIMERO levantar el servidor HTTP de IO en un hilo aparte para que no se bloquee el main
	go io_internal.IniciarServerIO(io_internal.Config_IO.PortIO)

	// Pequeño delay para asegurarnos que el server esté arriba
	time.Sleep(500 * time.Millisecond)

	// Ahora sí, mandar el paquete a Kernel
	io_internal.Conexion_Kernel(io_internal.Config_IO.IPKernel, io_internal.Config_IO.PortKernel, ioName, io_internal.Config_IO.IPIo, io_internal.Config_IO.PortIO)
}