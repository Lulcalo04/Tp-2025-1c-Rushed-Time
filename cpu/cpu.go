package main

import (
	cpu_internal "cpu/internal"
	"utils/client"
	"utils/globals"
)

func main() {

	//Crea el archivo donde se logea cpu
	globals.ConfigurarLogger("cpu")
	//Inicializa la config de cpu
	globals.IniciarConfiguracion("cpu/config.json", &cpu_internal.Config_CPU)

	//Realiza el handshake con el kernel
	client.HandshakeCon("Kernel", cpu_internal.Config_CPU.IPKernel, cpu_internal.Config_CPU.PortKernel)
}
