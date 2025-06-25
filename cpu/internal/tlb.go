package cpu_internal

type TLBs struct {
	Entrada          []EntradasTLB
	Algoritmo        string
	CantidadEntradas int
}

type EntradasTLB struct {
	PID    int
	Pagina int
	Marco  int
}

var TLB TLBs
var TLBHabilitada bool = false

// Funcion que inicializa la configuración de la TLB
func InicializarTLB() {
	if Config_CPU.TLBEntries > 0 { // Si la cantidad de entradas mayor a 0, habilitamos la TLB y la configuramos
		TLBHabilitada = true
		TLB.Algoritmo = Config_CPU.TLBReplacement
		TLB.CantidadEntradas = Config_CPU.TLBEntries
	}
}

func BuscarFrameEnTLB(numeroDePagina int) int {
	//Itero por todas las entradas de la TLB y si encuentro la página, retorno el marco correspondiente
	for i, entrada := range TLB.Entrada {
		if entrada.Pagina == numeroDePagina && entrada.PID == ProcesoEjecutando.PID {

			//Si el algoritmo es LRU, reordeno la entrada encontrada al final del slice
			if TLB.Algoritmo == "LRU" {
				// Mover la entrada encontrada al final del slice
				TLB.Entrada = append(append(TLB.Entrada[:i], TLB.Entrada[i+1:]...), entrada)
			}

			LogTLBHit(ProcesoEjecutando.PID, numeroDePagina)
			return entrada.Marco
		}
	}
	LogTLBMiss(ProcesoEjecutando.PID, numeroDePagina)
	return -1 // No se encontró la página en la TLB
}

func AgregarEntradaTLB(numeroDePagina int, numeroDeMarco int) {
	// Si la TLB esta llena se tiene que elegir una victima a eliminar
	if len(TLB.Entrada) == TLB.CantidadEntradas {
		//Tanto FIFO como LRU eliminan la primera entrada, ya que LRU reordena cada vez que se usa la TLB
		TLB.Entrada = TLB.Entrada[1:]
	}
	// Agregar la nueva entrada a la TLB
	nuevaEntrada := EntradasTLB{
		PID:    ProcesoEjecutando.PID,
		Pagina: numeroDePagina,
		Marco:  numeroDeMarco,
	}
	TLB.Entrada = append(TLB.Entrada, nuevaEntrada)
}
