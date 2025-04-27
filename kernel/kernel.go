package main

import (
	kernel_internal "kernel/internal"
	"utils/client"
	"utils/globals"
)

func main() {

	//Crea el archivo donde se logea kernel
	globals.ConfigurarLogger("kernel")

	//Inicializa la config de kernel
	globals.IniciarConfiguracion("kernel/config.json", &kernel_internal.Config_Kernel)

	//Prende el server de kernel en un hilo aparte
	go kernel_internal.IniciarServerKernel(kernel_internal.Config_Kernel.PortKernel)

	//Realiza el handshake con memoria
	client.HandshakeCon("Memoria", kernel_internal.Config_Kernel.IPMemory, kernel_internal.Config_Kernel.PortMemory)

	kernel_internal.Prueba()

}
