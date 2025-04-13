package main

import (
	cpu_internal "cpu/internal"
	"utils/client"
	"utils/globals"
)

func main() {

	//Crea el archivo donde se logea cpu
	globals.ConfigurarLogger("cpu")
	globals.IniciarConfiguracion("cpu/config.json", &cpu_internal.Config_CPU)

	//Mandar paquete a Memoria
	client.GenerarYEnviarPaquete(cpu_internal.Config_CPU.IPMemory, cpu_internal.Config_CPU.PortMemory)

	//Mandar paquete a Kernel
	client.GenerarYEnviarPaquete(cpu_internal.Config_CPU.IPKernel, cpu_internal.Config_CPU.PortKernel)

}
