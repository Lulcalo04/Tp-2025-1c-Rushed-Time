package main

import (
	"fmt"
	kernel_internal "kernel/internal"
)

func main() {

	//*Inicia las funcionalidades principales de kernel
	kernel_internal.IniciarKernel()

	//*Inicializa el proceso cero
	//nombreArchivoPseudocodigo, tamanioProceso := kernel_internal.InicializarProcesoCero()
	
	//*Funcion de prueba
	ImprimirDispositivosIO()
}

// Función de prueba para imprimir los dispositivos de IO
func ImprimirDispositivosIO() {
	fmt.Println("Dispositivos de IO registrados:")
	for _, dispositivo := range kernel_internal.ListaDispositivosIO {
		fmt.Printf("Nombre: %s, Instancias: %d\n", dispositivo.NombreIO, dispositivo.InstanciasIO)
	}
}
