package cpu_internal

import (
	"fmt"
	"strconv"
)

type CacheS struct {
	Entradas         []EntradaCache
	Algoritmo        string
	CantidadEntradas int
	Delay            int
}

type EntradaCache struct {
	PID             int
	Pagina          int
	Contenido       []byte
	DireccionFisica int
}

var Cache CacheS
var CacheHabilitada bool = false

func InicializarCache() {
	if Config_CPU.CacheEntries > 0 { // Si la cantidad de entradas es mayor a 0, habilitamos la Cache y la configuramos
		CacheHabilitada = true
		Cache.Algoritmo = Config_CPU.CacheReplacement
		Cache.CantidadEntradas = Config_CPU.CacheEntries
		Cache.Delay = Config_CPU.CacheDelay
	}
}

func AgregarPaginaEnCache(numeroDePagina int, contenido []byte, direccionFisica int) {
	// Si la Cache esta llena se tiene que elegir una victima a eliminar
	/* if len(Cache.Entradas) == Cache.CantidadEntradas {

		Cache.Entradas = Cache.Entradas[1:]
	} */

	// Agregar la nueva entrada a la Cache
	nuevaEntrada := EntradaCache{
		DireccionFisica: direccionFisica,
		PID:             ProcesoEjecutando.PID,
		Pagina:          numeroDePagina,
		Contenido:       contenido,
	}
	Cache.Entradas = append(Cache.Entradas, nuevaEntrada)
	LogPaginaIngresadaEnCache(ProcesoEjecutando.PID, numeroDePagina)
}

func BuscarPaginaEnCache(numeroDePagina int) *EntradaCache {
	for i := range Cache.Entradas {
		entrada := &Cache.Entradas[i]
		if entrada.Pagina == numeroDePagina && entrada.PID == ProcesoEjecutando.PID {
			LogPaginaEncontradaEnCache(ProcesoEjecutando.PID, numeroDePagina)
			return entrada
		}
	}
	LogPaginaFaltanteEnCache(ProcesoEjecutando.PID, numeroDePagina)
	return nil
}

func EscribirEnPagina(paginaCache *EntradaCache, desplazamiento int, valor string) {

	// Ver si el desplazamiento mas el valor a escribir no supera el tamaño de la pagina
	if EstructuraMemoriaDeCPU.TamanioPagina >= desplazamiento+len(valor) {
		// Copio el valor en el array de bytes a partir del desplazamiento indicado
		copy(paginaCache.Contenido[desplazamiento:], []byte(valor))
	} else {
		Logger.Debug("Error al escribir en la pagina: el desplazamiento mas el valor a escribir supera el tamaño de la pagina")
		// Manejar error si el desplazamiento es mayor al tamaño del contenido
	}

	direccionFisica := paginaCache.DireccionFisica + desplazamiento
	LogLecturaEscrituraMemoria(ProcesoEjecutando.PID, "ESCRIBIR", direccionFisica, valor)

}

func LeerDePagina(paginaCache *EntradaCache, desplazamiento int, tamanio string) {

	tamanioInt, err := strconv.Atoi(tamanio)
	if err != nil {
		Logger.Error("Error al convertir tamanio a int", "error", err)
		return
	}
	var TextoExtraido []byte
	// Ver si el desplazamiento mas el valor a leer no supera el tamaño de la pagina
	if EstructuraMemoriaDeCPU.TamanioPagina >= desplazamiento+tamanioInt {
		// Copio el valor del array de bytes a partir del desplazamiento indicado

		TextoExtraido = paginaCache.Contenido[desplazamiento : desplazamiento+tamanioInt]
		//texto3 := arr[inicio : inicio+hasta] /
	} else {
		Logger.Debug("Error al leer de la pagina: el desplazamiento mas el valor a leer supera el tamaño de la pagina")
		// Manejar error si el desplazamiento es mayor al tamaño del contenido
	}

	fmt.Println(string(TextoExtraido)) // Imprimir el contenido extraído de la página

	direccionFisica := paginaCache.DireccionFisica + desplazamiento
	LogLecturaEscrituraMemoria(ProcesoEjecutando.PID, "LEER", direccionFisica, string(TextoExtraido))

}
