package main

import (
	kernel_internal "kernel/internal"
	"utils/globals"
	// "utils/client"
	// "utils/server"
)

func main() {

	//Crea el archivo donde se logea kernel
	globals.ConfigurarLogger("kernel")

	//Inicializa la config de kernel
	globals.IniciarConfiguracion("kernel/config.json", &kernel_internal.Config_Kernel)

	//Prende el server de kernel
	kernel_internal.IniciarServerKernel(kernel_internal.Config_Kernel.PortKernel)

	/*
		//kernel_internal.Prueba()

		// Espera un enter del usuario para iniciar el kernel
		fmt.Println("Presione Enter para iniciar el kernel...")
		fmt.Scanln()
		kernel_internal.IniciarKernel()
		fmt.Println("Kernel iniciado.")

		//Mandar paquete a Memoria
		client.GenerarYEnviarPaquete(kernel_internal.Config_Kernel.IPMemory, kernel_internal.Config_Kernel.PortMemory)
	*/
}
