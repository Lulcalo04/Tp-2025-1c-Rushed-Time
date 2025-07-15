package main

import io_internal "inputoutput/internal"

func main() {

	//*Toma los parámetros recibidos por consola
	nombreArchivoConfiguracion, ioName := io_internal.RecibirParametrosConfiguracion()

	//*Inicia las funcionalidades principales de IO
	io_internal.IniciarIO(nombreArchivoConfiguracion, ioName)

	//*Escucha señales de desconexión del dispositivo IO
	go io_internal.EscucharSeñalDesconexion(ioName)

	select {}
}
