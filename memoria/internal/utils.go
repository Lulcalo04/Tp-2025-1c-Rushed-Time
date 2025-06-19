package memoria_internal

import "log/slog"

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
}

var MemoriaGlobal *Memoria

var Config_Memoria *ConfigMemoria

var Logger *slog.Logger



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

