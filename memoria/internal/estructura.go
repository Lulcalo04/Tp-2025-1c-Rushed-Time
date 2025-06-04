package memoria_internal

import "fmt"

type Memoria struct {
	datos         []byte             // slice de bytes que simula la RAM de usuario
	pageSize      int                // tamaño de cada página (frame) en bytes
	totalFrames   int                // número de frames = len(datos) / pageSize
	listaDeFrames []bool             // bitmap: listaDEFrames[i] = true si el frame i está libre
	tablas        map[int]*TablaPags // mapa PID → raíz de la tabla multinivel de ese proceso
}

func InicializarMemoria() {

	numFrames := Config_Memoria.MemorySize / Config_Memoria.PageSize

	// 2) Crear un slice que indique si cada marco está libre (true) u ocupado (false)
	marcosLibres := make([]bool, numFrames)
	for i := range marcosLibres {
		marcosLibres[i] = true
	}
}

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

//  prototipo de funcion para leer y escribir en un frame dado.

// dirección física = frameID*frameSize + offset
/*func (mp *Memoria) LeerDesdeFrame(frameID, offset, length int) ([]byte, error) {

	inicio := frameID*mp.pageSize + offset
	fin := inicio + length
	if inicio < 0 || fin > len(mp.datos) {
		return nil, fmt.Errorf("fuera de límites de memoria")
	}
	// Copiamos a un slice nuevo para no exponer el arreglo interno
	copia := make([]byte, length)
	copy(copia, mp.datos[inicio:fin])
	return copia, nil
}

func (mp *Memoria) EscribirEnFrame(frameID, offset int, contenido []byte) error {
	inicio := frameID*mp.pageSize + offset
	fin := inicio + len(contenido)
	if inicio < 0 || fin > len(mp.datos) {
		return fmt.Errorf("fuera de límites de memoria")
	}
	copy(mp.datos[inicio:fin], contenido)
	return nil
}
*/
