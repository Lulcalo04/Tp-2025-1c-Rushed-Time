package main

import (
	io_internal "inputoutput/internal"
	"utils/client"
	 "utils/server"
	"utils/globals"

	"fmt"
    "net"
    "os"
    "time"
)

func main() {

	//No necesito cargar os.args , esta se llena automaticamente con los argumentos que pasas cuando ejecutas desde la terminal

	//Verifica que se le pase el nombre del IO
	if len(os.Args) < 2 {
		fmt.Println("Error, mal escrito usa: ./bin/io [nombre]")
		os.Exit(1)
		}
	ioName := os.Args[1]

	//Crea el archivo donde se logea IO
	globals.ConfigurarLogger("io")
	//Inicializa la config de IO
	globals.IniciarConfiguracion("io/config.json", &io_internal.Config_IO)

	//Mandar paquete a Kernel
	client.HandshakeIo_Kernel(io_internal.Config_IO.IPKernel, io_internal.Config_IO.PortKernel, ioName, io_internal.Config_IO.IPIo, io_internal.Config_IO.PortIO)

	//Inicio un servidor en IO para esperar la respuesta de Kernel
	server.IniciarServer(io_internal.Config_IO.PortIO)
}

