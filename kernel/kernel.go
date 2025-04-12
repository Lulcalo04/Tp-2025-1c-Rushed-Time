package main

import (
	//kernel_utils "kernel/internal" //!ESTO SIRVE PARA CUANDO QUERAMOS USAR LA CONFIG DEL KERNEL NO BORRAR
	"utils/globals"
)

func main() {
	globals.ConfigurarLogger("kernel")
}
