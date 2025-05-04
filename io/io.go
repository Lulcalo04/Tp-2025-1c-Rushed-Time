package main

import (
	io_internal "inputoutput/internal"
	"utils/globals"
)

func main() {

	//Inicializa la config de IO
	globals.IniciarConfiguracion("io/config.json", &io_internal.Config_IO)

	//Crea el archivo donde se logea IO
	io_internal.Logger = globals.ConfigurarLogger("io", io_internal.Config_IO.LogLevel)

}
