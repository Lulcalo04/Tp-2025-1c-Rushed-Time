package main

import (
	"globals"
	memoria_internal "memoria/internal"
)

func main() {

	memoria_internal.InicializarMemoria()
	//Inicializa la config de memoria
	globals.IniciarConfiguracion("memoria/config.json", &memoria_internal.Config_Memoria)

	//Crea el archivo donde se logea memoria
	memoria_internal.Logger = globals.ConfigurarLogger("memoria", memoria_internal.Config_Memoria.LogLevel)

	memoria_internal.MemoriaGlobal = memoria_internal.NuevaMemoria(memoria_internal.Config_Memoria.MemorySize, memoria_internal.Config_Memoria.PageSize)

	//Prende el server de memoria
	memoria_internal.IniciarServerMemoria(memoria_internal.Config_Memoria.PortMemory)

}
