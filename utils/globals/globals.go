package globals

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

func IniciarConfiguracion(filePath string, config interface{}) {
	//var config *globals.Config

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

func ConfigurarLogger(nombreDelModulo string) {

	rutaDelLog := nombreDelModulo + "/" + nombreDelModulo + ".log"

	logFile, err := os.OpenFile(rutaDelLog, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}
