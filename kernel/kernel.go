package main

import (
kernel_internal "kernel/internal"
"utils/globals"
// "utils/client"
// "utils/server"
)

func main() {

//Crea el archivo donde se logea kernel
globals.ConfigurarLogger("kernel")

//Inicializa la config de kernel
globals.IniciarConfiguracion("kernel/config.json", &kernel_internal.Config_Kernel)

//Prende el server de kernel
kernel_internal.IniciarServerKernel(kernel_internal.Config_Kernel.PortKernel)

}
