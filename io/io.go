package main

import (
	io_internal "inputoutput/internal"
	"utils/globals"
)

func main() {

	globals.ConfigurarLogger("io")
	globals.IniciarConfiguracion("io/config.json", &io_internal.Config_IO)

}
