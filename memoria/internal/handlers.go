package memoria_internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"globals"
)

// -------------------------------------------Funcion para iniciar Server de Memoria-------------------------------------------------------------

func IniciarServerMemoria(puerto int) {

	//Transformo el puerto a string
	stringPuerto := fmt.Sprintf(":%d", puerto)

	//Declaro el server
	mux := http.NewServeMux()

	//Declaro los handlers para el server
	mux.HandleFunc("/handshake", HandshakeHandler)
	mux.HandleFunc("/handshake/cpu", HandshakeConCPU)
	mux.HandleFunc("/ping", PingHandler)
	mux.HandleFunc("/espacio/pedir", PidenEspacioHandler)
	mux.HandleFunc("/espacio/liberar", LiberarEspacioHandler)
	mux.HandleFunc("/cpu/instrucciones", InstruccionesHandler) // 
	mux.HandleFunc("/cpu/frame", CalcularFrameHandler) // pedido de frame desde CPU para la traduccion de direcciones
	mux.HandleFunc("/cpu/write", HacerWriteHandler)
	mux.HandleFunc("/cpu/read", HacerReadHandler)
	mux.HandleFunc("/cpu/goto", HacerGotoHandler)
	mux.HandleFunc("/dump", DumpMemoryHandler)

	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

// -------------------------------------------Endpoints de Memoria-------------------------------------------------------------

// * Endpoint de handshake = /handshake
func HandshakeHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	//Se utiliza el encoder para enviar la respuesta en formato JSON
	json.NewEncoder(w).Encode(
		map[string]string{
			"modulo":  "Memoria",
			"mensaje": "Conexion aceptada desde " + r.RemoteAddr,
		})
}

// * Endpoint de handshake = /handshake/cpu
func HandshakeConCPU(w http.ResponseWriter, r *http.Request) {
	// Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	var cpuHandshakeRequest globals.CPUToMemoriaHandshakeRequest
	if err := json.NewDecoder(r.Body).Decode(&cpuHandshakeRequest); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// Simulamos el handshake con la CPU (checkpoint 2)
	cpuHandshakeResponse := globals.CPUToMemoriaHandshakeResponse{
		TamanioMemoria:   Config_Memoria.MemorySize,
		TamanioPagina:    Config_Memoria.PageSize,
		EntradasPorTabla: Config_Memoria.EntriesPerPage,
		NivelesDeTabla:   Config_Memoria.NumberOfLevels,
	}

	json.NewEncoder(w).Encode(cpuHandshakeResponse)
}

//  *Endpoint de ping = /ping
func PingHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	//Se utiliza el encoder para enviar la respuesta en formato JSON
	var respuestaPing = globals.PingResponse{
		Modulo:  "Memoria",
		Mensaje: "Pong",
	}

	if err := json.NewEncoder(w).Encode(respuestaPing); err != nil {
		http.Error(w, "Error al codificar JSON", http.StatusInternalServerError)
	}
}

// * Endpoint de pedido de espacio = /espacio/pedir
func PidenEspacioHandler(w http.ResponseWriter, r *http.Request) {
	var pedidoRecibido globals.PeticionMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&pedidoRecibido); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// LOGICA PARA VERIFICAR Y RESERVAR ESPACIO EN MEMORIA

	tamanioSolicitado := pedidoRecibido.ProcesoPCB.TamanioEnMemoria

	//Este calculo es para determinar cuantas paginas  se necesitan para el tamanio solicitado, se suma el PageSize y se resta 1 para redondear hacia arriba
	framesNecesarios := (tamanioSolicitado + Config_Memoria.PageSize - 1) / Config_Memoria.PageSize 
	
	//verificar si hay suficiente espacio en memoria para el proceso
	framesLibres := MemoriaGlobal.framesLibres()
	pedidoEnMemoria := false

	if framesLibres >= framesNecesarios {
		Logger.Debug("Solicitud a Memoria aceptada", "PID", pedidoRecibido.ProcesoPCB.PID, "Tamanio", pedidoRecibido.ProcesoPCB.TamanioEnMemoria)
		pedidoEnMemoria = true 

		// Crear tabla raíz si no existe
		if MemoriaGlobal.tablas[pedidoRecibido.ProcesoPCB.PID] == nil {
			MemoriaGlobal.tablas[pedidoRecibido.ProcesoPCB.PID] = NuevaTablaPags()
		}
		tablaRaiz := MemoriaGlobal.tablas[pedidoRecibido.ProcesoPCB.PID]
	
		//Reservar espacio en memoria
		for pagina := 0; pagina < framesNecesarios; pagina++ {
			frameID, err := MemoriaGlobal.obtenerFrameLibre()
			if err != nil {
				http.Error(w, "Error al reservar frames", http.StatusInternalServerError)
				return
			}
			// Insertar en la tabla de páginas
			MemoriaGlobal.insertarEnMultinivel(tablaRaiz, pagina, frameID, 0)
		}
	}else {
		pedidoEnMemoria = false
		Logger.Debug("Solicitud a Memoria rechazada", "PID", pedidoRecibido.ProcesoPCB.PID, "Tamanio", pedidoRecibido.ProcesoPCB.TamanioEnMemoria)
		http.Error(w, "No hay suficiente espacio en memoria", http.StatusInsufficientStorage)
		return
	}
 
	if pedidoEnMemoria {
		// Si el pedido es valido, se hace la concesion de espacio
		Logger.Debug("Solicitud a Memoria aceptada", "PID", pedidoRecibido.ProcesoPCB.PID, "Tamanio", pedidoRecibido.ProcesoPCB.TamanioEnMemoria)

		// Preparar respuesta y codificarla como JSON (se envia automaticamente a traves del encode)
		resp := globals.PeticionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: true,
			Mensaje:   fmt.Sprintf("Espacio concedido para PID %d con %d de espacio", pedidoRecibido.ProcesoPCB.PID, pedidoRecibido.ProcesoPCB.TamanioEnMemoria),
		}
		json.NewEncoder(w).Encode(resp)
	} else {
		// Si el pedido no es valido, se envia un mensaje de error
		Logger.Debug("Solicitud a Memoria rechazada", "PID", pedidoRecibido.ProcesoPCB.PID, "Tamanio", pedidoRecibido.ProcesoPCB.TamanioEnMemoria)

		// Preparar respuesta y codificarla como JSON (se envia automaticamente a traves del encode)
		resp := globals.PeticionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: false,
			Mensaje:   fmt.Sprintf("Espacio no concedido para PID %d con %d de espacio", pedidoRecibido.ProcesoPCB.PID, pedidoRecibido.ProcesoPCB.TamanioEnMemoria),
		}
		json.NewEncoder(w).Encode(resp)
	}
}

// ^Endpoint de liberacion de espacio = /espacio/liberar
func LiberarEspacioHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	var pedidoRecibido globals.LiberacionMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&pedidoRecibido); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	// Logica para liberar espacio en memoria



	liberacionDeMemoria := true // Simulacion

	w.Header().Set("Content-Type", "application/json") // Setear Content-Type JSON (osea que la respuesta sera del tipo JSON)

	if liberacionDeMemoria {
		Logger.Debug("Liberacion de espacio aceptada", "PID", pedidoRecibido.PID)

		respuestaMemoria := globals.LiberacionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: true,
			Mensaje:   fmt.Sprintf("Liberacion de espacio en memoria aceptada para PID %d", pedidoRecibido.PID),
		}

		w.WriteHeader(http.StatusOK) //Confirmar que la respuesta es 200 OK
		json.NewEncoder(w).Encode(respuestaMemoria)
	} else {
		Logger.Debug("Liberacion de espacio fallida", "PID", pedidoRecibido.PID)

		respuestaMemoria := globals.LiberacionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: false,
			Mensaje:   fmt.Sprintf("Liberacion de espacio en memoria para PID %d rechazada", pedidoRecibido.PID),
		}

		w.WriteHeader(http.StatusOK) // Siempre responder OK con JSON aunque falle la liberacion
		json.NewEncoder(w).Encode(respuestaMemoria)
	}
}

// ^Endpoint de pedido de espacio = /dump
func DumpMemoryHandler(w http.ResponseWriter, r *http.Request) {
	var pedidoRecibido globals.DumpMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&pedidoRecibido); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// Logica para verificar espacio y reservarlo

	//! HAY QUE DESARROLLAR LA LOGICA DE DUMPEO DEL PROECESO

	dumpMemoria := true //! Simulamos que dumpea (checkpoint 2)

	if dumpMemoria {
		// Si el pedido es valido, se hace la concesion de espacio
		Logger.Debug("Solicitud de Dump Memory aceptada", "PID", pedidoRecibido.PID)

		// Preparar respuesta y codificarla como JSON (se envia automaticamente a traves del encode)
		resp := globals.DumpMemoryResponse{
			Modulo:    "Memoria",
			Respuesta: true, // Simulamos que se concede el espacio
			Mensaje:   fmt.Sprintf("Dump Memory concedido para PID %d", pedidoRecibido.PID),
		}
		json.NewEncoder(w).Encode(resp)
	} else {
		// Si el pedido no es valido, se envia un mensaje de error
		Logger.Debug("Solicitud de Dump Memory rechazada", "PID", pedidoRecibido.PID)

		// Preparar respuesta y codificarla como JSON (se envia automaticamente a traves del encode)
		resp := globals.DumpMemoryResponse{
			Modulo:    "Memoria",
			Respuesta: false,
			Mensaje:   fmt.Sprintf("Espacio no concedido para PID %d", pedidoRecibido.PID),
		}
		json.NewEncoder(w).Encode(resp)
	}

}


// -----------------------------------------------Funcion de instrucciones------------------------------------------------

// * Endpoint de instrucciones = /instrucciones
func InstruccionesHandler(w http.ResponseWriter, r *http.Request) {

	//verifica que el metodo sea POST ya que es el unico valido que nos puede llegar
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	//Declaro la variable request de tipo InstruccionesRequest para almacenar los datos enviadoos
	var req globals.InstruccionesRequest

	//Verifica que le hallamos mandado un JSON valido y decodifica el contenido en la variable request, si no puedo decodificarlo, devuelvo un error
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// Armamos el nombre del archivo: ejemplo "proceso_1.txt"
	filename := fmt.Sprintf("memoria/proceso_%d.txt", req.PID)

	//Verificamos si el archivo existe o se puede leer , si se puede content contendra los bytes del archivo
	content, err := os.ReadFile(filename)
	if err != nil {
		http.Error(w, "No se encontro el archivo del proceso", http.StatusNotFound)
		return
	}

	// Suponemos que el proceso1.txt contiene las instrucciones una por linea

	// string(content) convierte el proceso (que es un slice de bytes) a string
	// strings.TrimSpace elimina los espacios en blanco al principio y al final de la cadena para evitar saltos de linea
	// strings.Split divide la cadena en un slice de strings
	// lineas me devolvera, por ejemplo : []string{"Instruccion 1", "Instruccion 2", "Instruccion 3"}

	lineas := strings.Split(strings.TrimSpace(string(content)), "\n")

	// Preparamos la respuesta
	resp := globals.InstruccionesResponse{
		PID:           req.PID,
		Instrucciones: lineas,
	}

	//Establece el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	//Codifica el objeto resp en formato JSON y lo envia al cliente
	json.NewEncoder(w).Encode(resp)
}

//-------------------------------------------------Funcion para darle el frame a CPU ------------------------------------------------

// * Endpoint de frame = /cpu/frame (traduccion de direcciones)
func CalcularFrameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

	var request globals.SolicitudFrameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	//Busca la tabla de páginas raíz del proceso
    tablaRaiz := MemoriaGlobal.tablas[request.PID]
    if tablaRaiz == nil {
        http.Error(w, "Proceso no encontrado en memoria", http.StatusNotFound)
        return
    }
	
	// Busca el frame físico recorriendo la tabla multinivel
	frame, ok := MemoriaGlobal.buscarFramePorEntradas(tablaRaiz, request.EntradasPorNivel)
	if !ok {
		http.Error(w, "Página no asignada en memoria", http.StatusNotFound)
		return
	}

	response := globals.SolicitudFrameResponse{
		Frame: int(frame),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ^ Endpoint de write = /cpu/write 
func HacerWriteHandler(w http.ResponseWriter, r *http.Request) {

	var request globals.CPUWriteAMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	respuestaWrite := false

	tablaRaiz := MemoriaGlobal.tablas[request.PID]
    if tablaRaiz == nil {
        http.Error(w, "Proceso no encontrado en memoria", http.StatusNotFound)
        return
    }
	
    direccionFisica := request.DireccionFisica
	data := []byte(request.Data)

    // validacion
    if direccionFisica < 0 || direccionFisica+len(data) > len(MemoriaGlobal.datos) {
        http.Error(w, "No hay suficiente espacio en memoria para escribir el string", http.StatusBadRequest)
        return
    }

    // Escribo el string convertido en un slice de bytes en la memoria
    copy(MemoriaGlobal.datos[direccionFisica:], data)
    respuestaWrite = true

	response := globals.CPUWriteAMemoriaResponse{
		Respuesta: respuestaWrite,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ^ Endpoint de read = /cpu/read 
func HacerReadHandler(w http.ResponseWriter, r *http.Request) {
	var request globals.CPUReadAMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

    tablaRaiz := MemoriaGlobal.tablas[request.PID]
    if tablaRaiz == nil {
        http.Error(w, "Proceso no encontrado en memoria", http.StatusNotFound)
        return
    }

	var direccionFisica = request.DireccionFisica

    if direccionFisica < 0 || direccionFisica >= len(MemoriaGlobal.datos) {
        http.Error(w, "JSON invalido", http.StatusBadRequest)
        return
    }

    var respuestaRead int = int(MemoriaGlobal.datos[direccionFisica])

    response := globals.CPUReadAMemoriaResponse{
        Respuesta: true,
		Data:      respuestaRead,
    }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ^Endpoint de GoTo = /cpu/goto 
func HacerGotoHandler(w http.ResponseWriter, r *http.Request) {
	var request globals.CPUGotoAMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	
	// Simulamos la logica de salto en memoria (checkpoint 2)
	respuestaGoto := true //! Simulacion

	response := globals.CPUGotoAMemoriaResponse{
		Respuesta: respuestaGoto,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
