package main

import (
	kernel_internal "kernel/internal"
	"time"
)

func main() {

	//*Inicializa el proceso cero
	go kernel_internal.InicializarProcesoCero()


	//*Inicia las funcionalidades principales de kernel
	kernel_internal.IniciarKernel()


	time.Sleep(30 * time.Second) //?

}


