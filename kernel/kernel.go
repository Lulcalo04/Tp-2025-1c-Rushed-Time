package main

import (
	kernel_internal "kernel/internal"
	"utils/client"
	"utils/globals"
	"utils/server"
)

func main() {

	//Crea el archivo donde se logea kernel
	globals.ConfigurarLogger("kernel")
	//Inicializa la config de kernel
	globals.IniciarConfiguracion("kernel/config.json", &kernel_internal.Config_Kernel)

	//Mandar paquete a Memoria
	client.GenerarYEnviarPaquete(kernel_internal.Config_Kernel.IPMemory, kernel_internal.Config_Kernel.PortMemory)
	//Prende el server de kernel
	server.IniciarServer(kernel_internal.Config_Kernel.PortKernel)
}
