package main

import (
	"fmt"
	"utils/client"
	"utils/client/globals"
)

func main() {

	//Crea el archivo donde se logea memoria
	client.ConfigurarLogger("memoria")

	var Config_Memoria *globals.ConfigMemoria
	client.IniciarConfiguracion("memoria/config.json", &Config_Memoria)

	fmt.Println("Configuracion de memoria: ", Config_Memoria)

	//Enviamos el paquete

	client.GenerarYEnviarPaquete("127.0.0.1", 8004)

}
