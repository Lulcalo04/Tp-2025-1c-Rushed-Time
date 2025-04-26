package main

import (
	io_internal "inputoutput/internal"
	"utils/client"
	 "utils/server"
	"utils/globals"
	//"fmt"
   //"net"
    //"time"
)

func main() {

	//verificacion del nombre IO
	ioName := io_internal.VerificarNombreIO()

	//Crea el archivo donde se logea IO
	globals.ConfigurarLogger("io")
	
	//Inicializa la config de IO
	globals.IniciarConfiguracion("io/config.json", &io_internal.Config_IO)

	//Mandar paquete a Kernel
	client.Conexion_Kernel(io_internal.Config_IO.IPKernel, io_internal.Config_IO.PortKernel, ioName, io_internal.Config_IO.IPIo, io_internal.Config_IO.PortIO)

	//Inicio un servidor en IO para esperar la respuesta de Kernel
	server.IniciarServerIO(io_internal.Config_IO.PortIO)


}



