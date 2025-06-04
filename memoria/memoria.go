package main

import (
	memoria_internal "memoria/internal"
	"utils/globals"
)

func main() {

	memoria_internal.InicializarMemoria()
	//Inicializa la config de memoria
	globals.IniciarConfiguracion("memoria/config.json", &memoria_internal.Config_Memoria)

	//Crea el archivo donde se logea memoria
	memoria_internal.Logger = globals.ConfigurarLogger("memoria", memoria_internal.Config_Memoria.LogLevel)

	//Prende el server de memoria
	memoria_internal.IniciarServerMemoria(memoria_internal.Config_Memoria.PortMemory)

}
