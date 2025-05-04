package main

import (
	cpu_internal "cpu/internal"
	"utils/client"
	"utils/globals"
)

func main() {

	//Inicializa la config de cpu
	globals.IniciarConfiguracion("cpu/config.json", &cpu_internal.Config_CPU)

	//Crea el archivo donde se logea cpu
	cpu_internal.Logger = globals.ConfigurarLogger("cpu", cpu_internal.Config_CPU.LogLevel)

	//Realiza el handshake con el kernel
	client.HandshakeCon("Kernel", cpu_internal.Config_CPU.IPKernel, cpu_internal.Config_CPU.PortKernel, cpu_internal.Logger)
}
