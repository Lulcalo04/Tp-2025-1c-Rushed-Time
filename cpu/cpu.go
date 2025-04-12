package main

import (
	"utils/client"
	"utils/globals"
)

func main() {

	//Crea el archivo donde se logea cpu
	globals.ConfigurarLogger("cpu")
	globals.IniciarConfiguracion("cpu/config.json", &cpu.Config_CPU)

	//Mandar paquete a Memoria
	client.GenerarYEnviarPaquete(cpu.Config_CPU.IPMemory, cpu.Config_CPU.PortMemory)

	//Mandar paquete a Kernel
	client.GenerarYEnviarPaquete(cpu.Config_CPU.IPKernel, cpu.Config_CPU.PortKernel)

}
