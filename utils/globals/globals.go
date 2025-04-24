package globals

import (
	"encoding/json"
	"io"
	"log"
	"os"
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

// &-------------------------------------------Structs de PCB para los procesos-------------------------------------------
type PCB struct {
	PID               int            `json:"pid"`
	PC                int            `json:"pc"`
	Estado            Estado         `json:"estado"`
	MetricasDeEstados map[Estado]int `json:"metricas_de_estados"`
	MetricasDeTiempos map[Estado]int `json:"metricas_de_tiempos"` //en milisegundos, con una libreria especifica
	TamanioEnMemoria  int            `json:"tamanio_en_memoria"`  //Por ahora lo tomamos como entero, pero puede variar
}

// &-------------------------------------------Structs de handlers (Request y Response)-------------------------------------------

type PeticionMemoriaRequest struct {
	Modulo     string `json:"modulo"`
	ProcesoPCB PCB    `json:"proceso_pcb"`
}

type PeticionMemoriaResponse struct {
	Modulo    string `json:"modulo"`
	Respuesta bool   `json:"respuesta"`
	Mensaje   string `json:"mensaje"`
}

// &-------------------------------------------Inicio de configuraciones-------------------------------------------

func IniciarConfiguracion(filePath string, config interface{}) {

	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(config) // Se pasa el puntero directamente
	if err != nil {
		log.Fatalf("Error al decodificar la configuración: %s", err.Error())
	}
	if config == nil {
		log.Fatalf("No se pudo cargar la configuración")
	}

}

// &-------------------------------------------Inicio de funciones de logger-------------------------------------------

func ConfigurarLogger(nombreDelModulo string) {

	rutaDelLog := nombreDelModulo + "/" + nombreDelModulo + ".log"

	logFile, err := os.OpenFile(rutaDelLog, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}
