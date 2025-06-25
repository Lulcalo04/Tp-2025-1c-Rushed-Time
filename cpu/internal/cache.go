package cpu_internal

type CacheS struct {
	Entradas         []EntradaCache
	Algoritmo        string
	CantidadEntradas int
	Delay            int
}

type EntradaCache struct {
	PID       int
	Pagina    int
	Contenido []byte
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

func AgregarPaginaEnCache(numeroDePagina int, contenido []byte) {
	// Si la Cache esta llena se tiene que elegir una victima a eliminar
	/* if len(Cache.Entradas) == Cache.CantidadEntradas {

		Cache.Entradas = Cache.Entradas[1:]
	} */

	// Agregar la nueva entrada a la Cache
	nuevaEntrada := EntradaCache{
		PID:       ProcesoEjecutando.PID,
		Pagina:    numeroDePagina,
		Contenido: contenido,
	}
	Cache.Entradas = append(Cache.Entradas, nuevaEntrada)
	LogPaginaIngresadaEnCache(ProcesoEjecutando.PID, numeroDePagina)
}

func BuscarPaginaEnCache(numeroDePagina int) []byte {
	// Itero por todas las entradas de la Cache y si encuentro la página, retorno el contenido correspondiente
	for _, entrada := range Cache.Entradas {
		if entrada.Pagina == numeroDePagina && entrada.PID == ProcesoEjecutando.PID {
			LogPaginaEncontradaEnCache(ProcesoEjecutando.PID, numeroDePagina)
			return entrada.Contenido
		}
	}
	LogPaginaFaltanteEnCache(ProcesoEjecutando.PID, numeroDePagina)
	return nil // No se encontró la página en la Cache
}

func EscribirEnPagina(contenido []byte, desplazamiento int, valor string) {
	// Escribir el valor en el byte correspondiente

	arrayFijo := make([]byte, int(EstructuraMemoriaDeCPU.TamanioPagina))
	copy(arrayFijo, valor)

	if len(contenido) > desplazamiento {
		contenido[desplazamiento] = valor[0] // Asumiendo que valor es un string de un solo caracter
	} else {
		// Manejar error si el desplazamiento es mayor al tamaño del contenido
	}

}

func LeerDePagina() {

}
