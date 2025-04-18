package main

import (
	memoria_internal "memoria/internal"
	"utils/globals"
	"utils/server"
)

func main() {

	//Crea el archivo donde se logea memoria
	globals.ConfigurarLogger("memoria")
	//Inicializa la config de memoria
	globals.IniciarConfiguracion("memoria/config.json", &memoria_internal.Config_Memoria)

	//Prende el server de memoria
	server.IniciarServer(memoria_internal.Config_Memoria.PortMemory)
}
