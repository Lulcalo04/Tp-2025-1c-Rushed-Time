package main

import (
	"fmt"
	"globals"
	memoria_internal "memoria/internal"
)

func main() {

	//*Toma los parámetros recibidos por consola
	nombreArchivoConfiguracion := memoria_internal.RecibirParametrosConfiguracion()

	//*Inicializa la config de memoria
	fmt.Println("Iniciando configuracion " + nombreArchivoConfiguracion + " de memoria...")
	globals.IniciarConfiguracion("utils/configs/"+nombreArchivoConfiguracion+".json", &memoria_internal.Config_Memoria)

	//*Crea el archivo donde se logea memoria
	fmt.Println("Iniciando logger de memoria...")
	memoria_internal.Logger = globals.ConfigurarLogger("memoria", memoria_internal.Config_Memoria.LogLevel)

	//*Levanta la memoria con su configuración
	fmt.Println("Levantando memoria...")
	memoria_internal.NuevaMemoria()

	//*Prende el server de memoria
	fmt.Println("Iniciando servidor de MEMORIA, en el puerto:", memoria_internal.Config_Memoria.PortMemory)
	go memoria_internal.IniciarServerMemoria(memoria_internal.Config_Memoria.PortMemory)

	fmt.Println("Memoria funcionando.")
	select {}
}
