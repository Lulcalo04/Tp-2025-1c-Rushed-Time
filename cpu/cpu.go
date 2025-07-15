package main

import cpu_internal "cpu/internal"

func main() {

	//*Toma los parámetros recibidos por consola
	nombreArchivoConfiguracion := cpu_internal.RecibirParametrosConfiguracion()

	//*Inicia las funcionalidades principales de CPU
	cpu_internal.IniciarCPU(nombreArchivoConfiguracion)

	select {}
}
