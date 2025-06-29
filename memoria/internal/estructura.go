package memoria_internal

import (
	"fmt"
	"os"
)

type Memoria struct {
	datos          []byte                  // slice de bytes que simula la RAM de usuario
	pageSize       int                     // tamaño de cada página (frame) en bytes
	totalFrames    int                     // número de frames = len(datos) / pageSize
	listaDeFrames  []bool                  // bitmap: listaDEFrames[i] = true si el frame i está libre
	tablas         map[int]*TablaPags      // mapa PID → raíz de la tabla multinivel de ese proceso
	infoProc       map[int]*InfoPorProceso // asocia el PID con la metadata del proceso
	swapFile       *os.File                // Es un filedescriptor abierto para evitar el overhead de abrir y cerrar el archivo cada vez que se necesite acceder a él
	nextSwapOffset int64                   // próximo offset válido en swapfile
}

//-------------------- INICIALIZACION ------------------

func InicializarMemoria() {

	numFrames := Config_Memoria.MemorySize / Config_Memoria.PageSize

	// 2) Crear un slice que indique si cada marco está libre (true) u ocupado (false)
	marcosLibres := make([]bool, numFrames)
	for i := range marcosLibres {
		marcosLibres[i] = true
	}
}

func InicializarSWAP() {
	f, err := os.OpenFile(Config_Memoria.SwapfilePath,
		os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(fmt.Sprintf("No puedo abrir swapfile: %v", err))
	}
	MemoriaGlobal.swapFile = f
	MemoriaGlobal.nextSwapOffset = 0
	MemoriaGlobal.infoProc = make(map[int]*InfoPorProceso)
}

//---------------- FUNCIONES DE MEMORIA ----------------

func NuevaMemoria(memorySize, pageSize int) *Memoria {
	// 1) Crear el slice de bytes de tamaño memorySize
	datos := make([]byte, memorySize)

	// 2) Calcular cuántos frames caben en esos bytes
	totalFrames := memorySize / pageSize

	// 3) Inicializar el bitmap de frames libres
	listaDeFrames := make([]bool, totalFrames)
	for i := range listaDeFrames {
		listaDeFrames[i] = true
	}

	// 4) Inicializar el mapa de tablas multinivel
	tablas := make(map[int]*TablaPags)

	return &Memoria{
		datos:         datos,
		pageSize:      pageSize,
		totalFrames:   totalFrames,
		listaDeFrames: listaDeFrames,
		tablas:        tablas,
	}
}

// Funcion para obtener un frame libre
func (mp *Memoria) obtenerFrameLibre() (int, error) {
	for i, libre := range mp.listaDeFrames {
		if libre {
			mp.listaDeFrames[i] = false // lo reservamos
			return i, nil               // nil es la ausencia de error, el return devuelve el índice del frame libre.
		}
	}
	return -1, fmt.Errorf("no hay frames libres")
}

// Funcion para liberar un frame
func (mp *Memoria) liberarFrame(frameID int) {

	if frameID < 0 || frameID >= mp.totalFrames {
		fmt.Printf("Error: frameID %d fuera de rango", frameID)
		return
	}
	mp.listaDeFrames[frameID] = true
	// (O opcionalmente: limpiar el contenido de data[frameID*frameSize : (frameID+1)*frameSize])
}

func (mp *Memoria) framesLibres() int {
	// Contamos cuántos frames están libres
	libres := 0
	for _, libre := range mp.listaDeFrames {
		if libre {
			libres++
		}
	}
	return libres
}

func (mp *Memoria) LeerBytes(offset, length int) ([]byte, error) {
	// Verificamos que el offset y la longitud estén dentro de los límites del slice de datos
	if offset < 0 || offset+length > len(mp.datos) {
		return nil, fmt.Errorf("fuera de límites de memoria")
	}
	// Copiamos a un slice nuevo para no modificar los datos originales
	copia := make([]byte, length)
	copy(copia, mp.datos[offset:offset+length])
	return copia, nil
	//Devuelve  un slice de bytes que contiene los datos leídos desde el offset especificado.
}

func (mp *Memoria) EscribirBytes(offset int, contenido []byte) error {

	if offset < 0 || offset+len(contenido) > len(mp.datos) {
		return fmt.Errorf("fuera de límites de memoria")
	}
	// le agrego al slice de datos el contenido en la posición indicada por offset
	copy(mp.datos[offset:offset+len(contenido)], contenido)
	return nil
}

//------------------- TABLAS DE PAGINAS -------------------

func NuevaTablaPags() *TablaPags {
	// 1) Crear un slice de entradas con la longitud de EntriesPerTable
	entradas := make([]EntradaTabla, Config_Memoria.EntriesPerPage)

	// 2) Inicializar todas las entradas a nil (aún no asignadas)
	for i := range entradas {
		// Asignamos nil a cada entrada
		entradas[i] = nil
	}
	return &TablaPags{
		// asignamos el slice de entradas
		Entradas: entradas,
	}
}

func (mp *Memoria) insertarEnMultinivel(tabla *TablaPags, numPagina int, frame int, nivelActual int) {

	entradas := Config_Memoria.EntriesPerPage
	nivelMaximo := Config_Memoria.NumberOfLevels

	// Calcular divisor para obtener el índice correspondiente en este nivel
	divisor := 1
	for i := 0; i < nivelMaximo-nivelActual-1; i++ {
		divisor *= entradas
	}
	indice := (numPagina / divisor) % entradas
	// Último nivel: asigno directamente el frame físico
	if nivelActual == nivelMaximo-1 {
		tabla.Entradas[indice] = FrameID(frame)
		return
	}
	// Si la entrada no existe, creo una  subtabla
	if tabla.Entradas[indice] == nil {
		tabla.Entradas[indice] = NuevaTablaPags()
	}

	// Type assertion
	subTabla, ok := tabla.Entradas[indice].(*TablaPags)
	if !ok {
		Logger.Info(fmt.Sprintf("Error: entrada no es una subtabla en nivel %d", nivelActual))
		return
	}
	// Llamada recursiva
	mp.insertarEnMultinivel(subTabla, numPagina, frame, nivelActual+1)
}

func (mp *Memoria) buscarFramePorEntradas(tabla *TablaPags, entradas []int) (FrameID, bool) {
	actual := tabla
	for nivel, idx := range entradas {
		if nivel == Config_Memoria.NumberOfLevels-1 {
			frame, ok := actual.Entradas[idx].(FrameID)
			return frame, ok
		}
		subTabla, ok := actual.Entradas[idx].(*TablaPags)
		if !ok || subTabla == nil {
			return 0, false
		}
		actual = subTabla
	}
	return 0, false
}
