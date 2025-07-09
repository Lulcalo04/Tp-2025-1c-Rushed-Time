package main

import (
	"fmt"
	"globals"
	memoria_internal "memoria/internal"
)

func main() {

	//Inicializa la config de memoria
	fmt.Println("Iniciando configuracion de memoria...")
	globals.IniciarConfiguracion("memoria/config.json", &memoria_internal.Config_Memoria)

	//Crea el archivo donde se logea memoria
	fmt.Println("Iniciando logger de memoria...")
	memoria_internal.Logger = globals.ConfigurarLogger("memoria", memoria_internal.Config_Memoria.LogLevel)

	fmt.Println("Levantando memoria...")
	memoria_internal.NuevaMemoria()

	//Prende el server de memoria
	fmt.Println("Iniciando servidor de memoria en el puerto:", memoria_internal.Config_Memoria.PortMemory)
	go memoria_internal.IniciarServerMemoria(memoria_internal.Config_Memoria.PortMemory)

	fmt.Println("Memoria funcionando.")
	select {}
}
