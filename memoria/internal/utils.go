package memoria_internal

import (
	"fmt"
	"log/slog"
	"os"
)

type ConfigMemoria struct {
	PortMemory     int    `json:"port_memory"`
	MemorySize     int    `json:"memory_size"`
	PageSize       int    `json:"page_size"`
	EntriesPerPage int    `json:"entries_per_page"`
	NumberOfLevels int    `json:"number_of_levels"`
	MemoryDelay    int    `json:"memory_delay"`
	SwapfilePath   string `json:"swapfile_path"`
	SwapDelay      int    `json:"swap_delay"`
	LogLevel       string `json:"log_level"`
	DumpPath       string `json:"dump_path"`
	ScriptsPath    string `json:"scripts_path"`
}

var MemoriaGlobal *Memoria

var Config_Memoria *ConfigMemoria

var Logger *slog.Logger

var TamaniosProcesos = make(map[int]int)

type InstruccionesRequest struct {
	PID int `json:"pid"`
}

type InstruccionesResponse struct {
	PID           int      `json:"pid"`
	Instrucciones []string `json:"instrucciones"`
}

// FrameID es el índice de un marco físico en memoriaPrincipal (0 .. numFrames-1).
type FrameID int

// EntradaTabla:
//
//	– nil  ->  no asignado
//	– *TablaPags  -> uun puntero a una subtabla de nivel inferior
//	– FrameID   -> un frame en memoriaPrincipal
type EntradaTabla interface{}

// TablaPags representa una tabla de páginas en un nivel concreto.
type TablaPags struct {
	// Entradas tiene siempre longitud = Config_Memoria.EntriesPerTable.
	Entradas []EntradaTabla
}

//-------------------------------------------------------------------------

// type TamanioProcesos map[int]int  // asocia el PID con el tamaño del proceso
type InfoPorProceso struct {
	Pages         []PageInfo          //  slice de información de páginas del proceso
	Instrucciones []string            //  lista de instrucciones de ese proceso
	Size          int                 // tamaño del proceso en bytes
	TablaRaiz     *TablaPags          // tabla raíz del proceso
	Metricas      *MetricasPorProceso // lista de métricas del proceso
	PathArchivo   string              // path al archivo de pseudocodigo del proceso
}

type PageInfo struct {
	InRAM   bool  // si está en RAM o ya pasó a swap
	FrameID int   // el frame físico (si InRAM)
	Offset  int64 // desplazamiento en swapfile.bin (si !InRAM)
}

type MetricasPorProceso struct {
	AccesoATablaDePaginas    int // cantidad de veces que se accedió a la tabla de páginas	(listo)
	InstruccionesSolicitadas int // cantidad de instrucciones solicitadas					(listo)
	BajadasASwap             int // cantidad de veces que se bajó a swap								(listo)
	SubidasAMemoriaPrincipal int // cantidad de veces que se subió a memoria principal		(listo)
	LecturasDeMemoria        int // cantidad de lecturas de memoria								(Listo)
	EscriturasDeMemoria      int // cantidad de escrituras de memoria							(listo)
}

func RecibirParametrosConfiguracion() string {
	if len(os.Args) < 2 {
		fmt.Println("Error, mal escrito usa: .memoria/memoria.go [archivo_configuracion]")
		os.Exit(1)
	}

	// Leer y cargar la configuración del archivo de configuración
	nombreArchivoConfiguracion := os.Args[1]

	return nombreArchivoConfiguracion
}

func NuevaMemoria() {
	// 1) Crear el slice de bytes de tamaño memorySize
	datos := make([]byte, Config_Memoria.MemorySize)

	// 2) Calcular cuántos frames caben en esos bytes
	totalFrames := Config_Memoria.MemorySize / Config_Memoria.PageSize

	// 3) Inicializar el bitmap de frames libres
	listaDeFrames := make([]bool, totalFrames)
	for i := range listaDeFrames {
		listaDeFrames[i] = true
	}

	// 4) Inicializar el mapa de tablas multinivel
	tablas := make(map[int]*TablaPags)

	// 5) Inicializar el archivo de swap
	f, err := os.OpenFile(Config_Memoria.SwapfilePath,
		os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(fmt.Sprintf("No puedo abrir swapfile: %v", err))
	}

	// 6) Inicializar la memoria global
	MemoriaGlobal = &Memoria{
		datos:           datos,
		pageSize:        Config_Memoria.PageSize,
		totalFrames:     totalFrames,
		listaDeFrames:   listaDeFrames,
		tablas:          tablas,
		infoProc:        make(map[int]*InfoPorProceso), // Inicializamos el mapa de info de procesos
		swapFile:        f,
		freeSwapOffsets: make([]int64, 0), // Inicializamos con un slice
		nextSwapOffset:  0,                // Inicializamos el offset de swap
	}
}
