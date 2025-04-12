package main

import (
	"memoria_utils"
	"utils/globals"
	"utils/server"
)

// "fmt"
// fmt.Println("Configuracion de memoria: ", globals.Config_Memoria)

func main() {

	//Crea el archivo donde se logea cpu
	globals.ConfigurarLogger("memoria")
	globals.IniciarConfiguracion("memoria/config.json", &memoria_utils.Config_Memoria)

	server.IniciarServer(memoria_utils.Config_Memoria.PortMemory)
}
