package main

import kernel_internal "kernel/internal"

func main() {

	//*Toma los parámetros recibidos por consola
	nombreArchivoConfiguracion, nombreArchivoPseudocodigo, tamanioProceso := kernel_internal.RecibirParametrosConfiguracion()

	//*Inicia las funcionalidades principales de KERNEL
	kernel_internal.IniciarKernel(nombreArchivoConfiguracion)

	//*Inicializa el proceso cero
	kernel_internal.InicializarProcesoCero(tamanioProceso, nombreArchivoPseudocodigo)

	select {}
}
