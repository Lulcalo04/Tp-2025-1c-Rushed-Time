package main

import memoria_internal "memoria/internal"

func main() {

	//*Toma los parámetros recibidos por consola
	nombreArchivoConfiguracion := memoria_internal.RecibirParametrosConfiguracion()

	//*Inicia las funcionalidades principales de MEMORIA
	memoria_internal.IniciarMemoria(nombreArchivoConfiguracion)

	select {}
}
