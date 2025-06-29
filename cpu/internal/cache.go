package cpu_internal

import (
	"fmt"
	"strconv"
	"time"
)

type CacheStruct struct {
	Entradas                []EntradaCache
	Algoritmo               string
	CantidadEntradas        int
	Delay                   int
	PunteroEntradaReemplazo int // Puntero que se modifica al ejecutar los algoritmos de reemplazo
}

type EntradaCache struct {
	PID             int
	Pagina          int
	Contenido       []byte
	DireccionFisica int
	Usado           bool // Indica si la pagina fue usada
	Modificado      bool // Indica si la pagina fue modificada
}

var Cache CacheStruct
var CacheHabilitada bool = false

func InicializarCache() {
	if Config_CPU.CacheEntries > 0 { // Si la cantidad de entradas es mayor a 0, habilitamos la Cache y la configuramos
		CacheHabilitada = true
		Cache.Algoritmo = Config_CPU.CacheReplacement
		Cache.CantidadEntradas = Config_CPU.CacheEntries
		Cache.Delay = Config_CPU.CacheDelay
		Cache.PunteroEntradaReemplazo = 0
	}
}

func AgregarPaginaEnCache(numeroDePagina int, contenido []byte, direccionFisica int) *EntradaCache {
	// Simulamos el delay de la Cache
	time.Sleep(time.Duration(Cache.Delay) * time.Millisecond)

	// Declaro la nueva entrada de Cache
	nuevaEntrada := EntradaCache{
		DireccionFisica: direccionFisica,
		PID:             ProcesoEjecutando.PID,
		Pagina:          numeroDePagina,
		Contenido:       contenido,
		Usado:           true,
		Modificado:      false,
	}

	LogPaginaIngresadaEnCache(ProcesoEjecutando.PID, numeroDePagina)

	// Si la Cache esta llena se tiene que elegir una victima a eliminar
	if len(Cache.Entradas) == Cache.CantidadEntradas {
		paginaVictima, posicionEnCache := ElegirPaginaVictima()

		if paginaVictima.Modificado {
			// Reemplazamos la pagina victima con la nueva entrada
			ReemplazarPaginaVictima(paginaVictima, &nuevaEntrada, posicionEnCache)
		}

		// Retorna la entrada reemplazada en la Cache
		return &Cache.Entradas[posicionEnCache]
	} else {
		// Si la Cache no esta llena, simplemente agregamos la nueva entrada al final de la lista
		Cache.Entradas = append(Cache.Entradas, nuevaEntrada)

		// Retorna la nueva entrada, que es la ultima de la lista
		return &Cache.Entradas[len(Cache.Entradas)-1]
	}
}

func BuscarPaginaEnCache(numeroDePagina int) *EntradaCache {
	// Simulamos el delay de la Cache
	time.Sleep(time.Duration(Cache.Delay) * time.Millisecond)

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

func EscribirEnPaginaCache(paginaCache *EntradaCache, desplazamiento int, valor string) {
	// Simulamos el delay de la Cache
	time.Sleep(time.Duration(Cache.Delay) * time.Millisecond)

	// Ver si el desplazamiento mas el valor a escribir no supera el tamaño de la pagina
	if EstructuraMemoriaDeCPU.TamanioPagina >= desplazamiento+len(valor) {

		// Copio el valor en el array de bytes a partir del desplazamiento indicado
		copy(paginaCache.Contenido[desplazamiento:], []byte(valor))

		// Actualizo el flag de modificado a true
		paginaCache.Modificado = true

		// Actualizo el flag de usado a true
		paginaCache.Usado = true

	} else {
		Logger.Debug("Error al escribir en la pagina: el desplazamiento mas el valor a escribir supera el tamaño de la pagina")
		// Manejar error si el desplazamiento es mayor al tamaño del contenido
	}

	direccionFisica := paginaCache.DireccionFisica + desplazamiento
	LogLecturaEscrituraMemoria(ProcesoEjecutando.PID, "ESCRIBIR", direccionFisica, valor)

}

func LeerDePaginaCache(paginaCache *EntradaCache, desplazamiento int, tamanio string) {
	// Simulamos el delay de la Cache
	time.Sleep(time.Duration(Cache.Delay) * time.Millisecond)

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

		// Actualizo el flag de usado a true
		paginaCache.Usado = true

	} else {
		Logger.Debug("Error al leer de la pagina: el desplazamiento mas el valor a leer supera el tamaño de la pagina")
		// Manejar error si el desplazamiento es mayor al tamaño del contenido
	}

	fmt.Println(string(TextoExtraido)) // Imprimir el contenido extraído de la página

	direccionFisica := paginaCache.DireccionFisica + desplazamiento
	LogLecturaEscrituraMemoria(ProcesoEjecutando.PID, "LEER", direccionFisica, string(TextoExtraido))

}

func ElegirPaginaVictima() (*EntradaCache, int) {
	paginaVictima := &EntradaCache{}
	posicionEnCache := -1
	cantidad := len(Cache.Entradas) // Me guardo la cantidad de entradas en la Cache

	switch Cache.Algoritmo {
	case "CLOCK":
		for {
			// Me guardo en "entrada" la direccion en memoria de la entrada a analizar
			entrada := &Cache.Entradas[Cache.PunteroEntradaReemplazo]
			if !entrada.Usado {
				// Si la entrada no fue usada, la elijo como victima
				// Me guardo la pagina victima y su posicion en la Cache
				posicion := Cache.PunteroEntradaReemplazo
				Cache.PunteroEntradaReemplazo = (Cache.PunteroEntradaReemplazo + 1) % cantidad

				// Retorno la pagina victima y su posicion en la Cache
				return entrada, posicion
			} else {
				// Si la entrada fue usada, la marco como no usada y avanzo el puntero
				entrada.Usado = false
				Cache.PunteroEntradaReemplazo = (Cache.PunteroEntradaReemplazo + 1) % cantidad
			}
		}
	case "CLOCK-M":
		// Primera pasada: buscar U=false y M=false
		for i := 0; i < cantidad; i++ {
			idx := Cache.PunteroEntradaReemplazo
			entrada := &Cache.Entradas[idx]
			if !entrada.Usado && !entrada.Modificado {
				posicion := idx
				Cache.PunteroEntradaReemplazo = (idx + 1) % cantidad
				return entrada, posicion
			}
			// Si no cumple, pongo Usado en false
			entrada.Usado = false
			Cache.PunteroEntradaReemplazo = (idx + 1) % cantidad
		}
		// Segunda pasada: buscar U=false y M=true
		for i := 0; i < cantidad; i++ {
			idx := Cache.PunteroEntradaReemplazo
			entrada := &Cache.Entradas[idx]
			if !entrada.Usado && entrada.Modificado {
				posicion := idx
				Cache.PunteroEntradaReemplazo = (idx + 1) % cantidad
				return entrada, posicion
			}
			Cache.PunteroEntradaReemplazo = (idx + 1) % cantidad
		}
	}
	return paginaVictima, posicionEnCache
}

func ReemplazarPaginaVictima(paginaVictima *EntradaCache, nuevaEntrada *EntradaCache, posicionEnCache int) {
	// Simulamos el delay de la Cache
	time.Sleep(time.Duration(Cache.Delay) * time.Millisecond)

	// Le enviamos a memoria la pagina victima para que sea actualizada
	ActualizarPaginaEnMemoria(paginaVictima.PID, paginaVictima.Pagina, paginaVictima.Contenido)

	// Reemplazamos la pagina victima con la nueva entrada
	Cache.Entradas[posicionEnCache] = *nuevaEntrada
}
