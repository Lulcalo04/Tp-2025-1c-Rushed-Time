package main

import (
	io_internal "inputoutput/internal"
	"utils/client"
	"utils/globals"
)

func main() {

	//Crea el archivo donde se logea IO
	globals.ConfigurarLogger("io")
	//Inicializa la config de IO
	globals.IniciarConfiguracion("io/config.json", &io_internal.Config_IO)

	//Mandar paquete a Kernel
	client.GenerarYEnviarPaquete(io_internal.Config_IO.IPKernel, io_internal.Config_IO.PortKernel)
}
