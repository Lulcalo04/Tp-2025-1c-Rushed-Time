package main

import (
	kernel_internal "kernel/internal"
)

func main() {

	//*Inicia las funcionalidades principales de kernel
	kernel_internal.IniciarKernel()

	//*Inicializa el proceso cero
	kernel_internal.InicializarProcesoCero()

	select {}
}
