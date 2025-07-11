package globals

import (
	"encoding/json"
	"log/slog"
	"os"
	"time"
)

// &-------------------------------------------Tipos de datos para el manejo de los estados de los procesos-------------------------------------------

type Estado string

const (
	New         Estado = "NEW"
	Ready       Estado = "READY"
	Exec        Estado = "EXEC"
	Blocked     Estado = "BLOCKED"
	SuspReady   Estado = "SUSP_READY"
	SuspBlocked Estado = "SUSP_BLOCKED"
	Exit        Estado = "EXIT"
)

type EstructuraRafaga struct {
	TiempoDeRafaga float64
	YaCalculado    bool
}

// &-------------------------------------------Structs de PCB para los procesos-------------------------------------------
type PCB struct {
	PID                  int                      `json:"pid"`                     // Identificador único del proceso
	PC                   int                      `json:"pc"`                      // Contador de programa, indica la posición de la instrucción a ejecutar
	Estado               Estado                   `json:"estado"`                  // Estado actual del proceso
	PathArchivoPseudo    string                   `json:"path"`                    // Ruta del archivo de pseudocódigo del proceso
	InicioEstadoActual   time.Time                `json:"inicio_estado_actual"`    // Marca el tiempo en que el proceso entra al estado actual
	MetricasDeEstados    map[Estado]int           `json:"metricas_de_estados"`     // Contador de veces que el proceso estuvo en cada estado
	MetricasDeTiempos    map[Estado]time.Duration `json:"metricas_de_tiempos"`     // Contador de tiempo que el proceso estuvo en cada estado
	TamanioEnMemoria     int                      `json:"tamanio_en_memoria"`      // Tamaño del proceso en memoria, en bytes
	EstimacionDeRafaga   EstructuraRafaga         `json:"estimacion_de_rafaga"`    // Estimación de la duración de la próxima ráfaga de CPU del proceso
	TiempoDeUltimaRafaga time.Duration            `json:"tiempo_de_ultima_rafaga"` // Marca el tiempo que duró su última ráfaga de CPU
}

// &-------------------------------------------Inicio de configuraciones-------------------------------------------

func IniciarConfiguracion(filePath string, config interface{}) {

	configFile, err := os.Open(filePath)
	if err != nil {
		slog.Debug(err.Error())
		os.Exit(1)
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(config)
	if err != nil {
		slog.Debug("Error al decodificar la configuración", "error", err.Error())
		os.Exit(1)
	}
	if config == nil {
		slog.Debug("No se pudo cargar la configuración")
		os.Exit(1)
	}

}

// &-------------------------------------------Inicio de funciones de logger-------------------------------------------

func ConfigurarLogger(nombreDelModulo string, logLevelModulo string) *slog.Logger {

	// Definimos la ruta del log
	rutaDelLog := nombreDelModulo + "/" + nombreDelModulo + ".log"
	// Definimos el nivel de log
	nivel := PasarStringALogLevel(logLevelModulo)

	logFile, err := os.OpenFile(rutaDelLog, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	handler := slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: nivel})

	// Guardamos la instancia del logger en una variable global o accesible por el módulo
	Logger := slog.New(handler)

	return Logger
}

func PasarStringALogLevel(nivel string) slog.Level {
	var logLevel slog.Level

	switch nivel {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	return logLevel
}
