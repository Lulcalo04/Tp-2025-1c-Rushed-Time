package memoria_internal

import (
	"encoding/json"
	"fmt"
	"globals"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// * hechos
// ^ en proceso
// ! por hacer
// TODO -- comentarios importantes en funciones
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
	mux.HandleFunc("/syscall/swappeo", SwappingHandler) //PREGUNTAR A LOS CHICOS COMO TIENEN ESTOS ENDPOINTS
	mux.HandleFunc("/syscall/restaurar", RestaurarHandler)
	mux.HandleFunc("/cpu/instrucciones", InstruccionesHandler)
	mux.HandleFunc("/cpu/frame", CalcularFrameHandler) // pedido de frame desde CPU para la traduccion de direcciones
	//mux.HandleFunc("/cpu/write", HacerWriteHandler)
	//mux.HandleFunc("/cpu/read", HacerReadHandler)
	mux.HandleFunc("/dump", DumpMemoryHandler)

	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

// -------------------------------------------Endpoints de Memoria-------------------------------------------------------------

// * Endpoint de handshake = /handshake
func HandshakeHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(
		map[string]string{
			"modulo":  "Memoria",
			"mensaje": "Conexion aceptada desde " + r.RemoteAddr,
		})
}

// * Endpoint de handshake = /handshake/cpu
func HandshakeConCPU(w http.ResponseWriter, r *http.Request) {

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

// *Endpoint de ping = /ping
func PingHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

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
	tamanioSolicitado := pedidoRecibido.ProcesoPCB.TamanioEnMemoria

	//Este calculo es para determinar cuantas paginas  se necesitan para el tamanio solicitado, se suma el PageSize y se resta 1 para redondear hacia arriba
	framesNecesarios := (tamanioSolicitado + Config_Memoria.PageSize - 1) / Config_Memoria.PageSize

	//verificar si hay suficiente espacio en memoria para el proceso
	framesLibres := MemoriaGlobal.framesLibres()
	pedidoEnMemoria := false
	var framesReservados []int

	if framesLibres >= framesNecesarios {
		Logger.Debug("Solicitud a Memoria aceptada", "PID", pedidoRecibido.ProcesoPCB.PID, "Tamanio", pedidoRecibido.ProcesoPCB.TamanioEnMemoria)
		pedidoEnMemoria = true

		// Crear tabla raíz si no existe
		if MemoriaGlobal.tablas[pedidoRecibido.ProcesoPCB.PID] == nil {
			MemoriaGlobal.tablas[pedidoRecibido.ProcesoPCB.PID] = NuevaTablaPags()
		}

		tablaRaiz := MemoriaGlobal.tablas[pedidoRecibido.ProcesoPCB.PID]
		//guardo los valores en el struct del progeso
		MemoriaGlobal.infoProc[pedidoRecibido.ProcesoPCB.PID].Size = pedidoRecibido.ProcesoPCB.TamanioEnMemoria
		MemoriaGlobal.infoProc[pedidoRecibido.ProcesoPCB.PID].TablaRaiz = tablaRaiz

		//Reservar espacio en memoria

		for pagina := 0; pagina < framesNecesarios; pagina++ {
			frameID, err := MemoriaGlobal.obtenerFrameLibre()
			if err != nil {
				http.Error(w, "Error al reservar frames", http.StatusInternalServerError)
				return
			}
			framesReservados = append(framesReservados, frameID)
			// Insertar en la tabla de páginas
			MemoriaGlobal.insertarEnMultinivel(tablaRaiz, pagina, frameID, 0)
		}
	} else {
		pedidoEnMemoria = false
		Logger.Debug("Solicitud a Memoria rechazada", "PID", pedidoRecibido.ProcesoPCB.PID, "Tamanio", pedidoRecibido.ProcesoPCB.PID)
		http.Error(w, "No hay suficiente espacio en memoria", http.StatusInsufficientStorage)
		return
	}

	//El numero de paginas es el mismo que el numero de frames necesarios
	pi := &InfoPorProceso{
		Pages:         make([]PageInfo, framesNecesarios),
		Size:          pedidoRecibido.ProcesoPCB.TamanioEnMemoria,
		TablaRaiz:     MemoriaGlobal.tablas[pedidoRecibido.ProcesoPCB.PID],
		Instrucciones: listaDeInstrucciones(pedidoRecibido.ProcesoPCB.PID),
	}
	for i, frameID := range framesReservados {
		pi.Pages[i] = PageInfo{
			InRAM:   true,
			FrameID: frameID,
			Offset:  0, // No esta en SWAP
		}
	}
	MemoriaGlobal.infoProc[pedidoRecibido.ProcesoPCB.PID] = pi

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

// !Endpoint de liberacion de espacio = /espacio/liberar
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

// *Endpoint de pedido de espacio = /dump

func DumpMemoryHandler(w http.ResponseWriter, r *http.Request) {

	var pedidoRecibido globals.DumpMemoryRequest
	var numPaginas int
	if err := json.NewDecoder(r.Body).Decode(&pedidoRecibido); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// Logica para hacer el dump de memoria:

	// verificamos que el proceso exista en memoria
	tablaRaiz := MemoriaGlobal.tablas[pedidoRecibido.PID]
	if tablaRaiz == nil {
		http.Error(w, "Proceso no encontrado en memoria", http.StatusNotFound)
		return
	}

	//calculo del tamanio de paginas del proceso:

	tamanioPaginas := Config_Memoria.PageSize
	numPaginas = (TamaniosProcesos[pedidoRecibido.PID] + tamanioPaginas - 1) / tamanioPaginas
	//Creo el buffer para el dump

	dump := make([]byte, numPaginas*tamanioPaginas)

	for pagina := 0; pagina < numPaginas; pagina++ {

		entradas := calcularEntradasPorNivel(pagina, Config_Memoria.EntriesPerPage, Config_Memoria.NumberOfLevels)
		frameID, ok := MemoriaGlobal.buscarFramePorEntradas(tablaRaiz, entradas)

		if ok {

			origen := int(frameID) * tamanioPaginas
			destino := pagina * tamanioPaginas
			copy(dump[destino:destino+tamanioPaginas], MemoriaGlobal.datos[origen:origen+tamanioPaginas])
		}
		//TODO -- nota: no ponemos que si no esta en memoria principal, sus valores se pongan en cero, porque ya vienen inicializados en cero.
	}

	//timestamp guarda el dia y hora donde se hace el dump, en formato YYYYMMDD_HHMMSS
	// TODO  nota sobre Unix: creo que devuelve el tiempo  desde el 1 de enero de 1970
	timestamp := time.Now().Unix() //! arreglar esto
	nombreArchivo := fmt.Sprintf("<%d><%d>.dpm", pedidoRecibido.PID, timestamp)
	dumpPath := filepath.Join(Config_Memoria.DumpPath, nombreArchivo)

	if err := os.WriteFile(dumpPath, dump, 0644); err != nil {
		// Si el pedido es valido, se hace la concesion de espacio
		Logger.Debug("Solicitud de Dump Memory aceptada", "PID", pedidoRecibido.PID)

		resp := globals.DumpMemoryResponse{
			Modulo:    "Memoria",
			Respuesta: true,
			Mensaje:   fmt.Sprintf("Dump Memory concedido para PID %d", pedidoRecibido.PID),
		}
		json.NewEncoder(w).Encode(resp)
	} else {
		Logger.Debug("Solicitud de Dump Memory rechazada", "PID", pedidoRecibido.PID)

		resp := globals.DumpMemoryResponse{
			Modulo:    "Memoria",
			Respuesta: false,
			Mensaje:   fmt.Sprintf("Espacio no concedido para PID %d", pedidoRecibido.PID),
		}
		json.NewEncoder(w).Encode(resp)
	}
}

func SwappingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PID int `json:"pid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	pi, ok := MemoriaGlobal.infoProc[req.PID]
	if !ok {
		http.Error(w, "Proceso no existe", http.StatusNotFound)
		return
	}
	// Suspender todas sus páginas
	for i := range pi.Pages {
		if err := MemoriaGlobal.SuspenderPagina(req.PID, i); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"modulo": "Memoria", "status": "suspendido",
	})

	Logger.Debug("Proceso swappeado exitosamente", "PID", req.PID)
}

func RestaurarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PID int `json:"pid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	pi, ok := MemoriaGlobal.infoProc[req.PID]
	if !ok {
		http.Error(w, "Proceso no existe", http.StatusNotFound)
		return
	}

	for i := range pi.Pages {
		if err := MemoriaGlobal.RestaurarPagina(req.PID, i); err != nil {
			http.Error(w, err.Error(), http.StatusInsufficientStorage)
			return
		}
	}
	json.NewEncoder(w).Encode(map[string]string{
		"modulo": "Memoria", "status": "resumido",
	})
}

// -----------------------------------------------Funcion de instrucciones------------------------------------------------

// * Endpoint de instrucciones = /instrucciones
func InstruccionesHandler(w http.ResponseWriter, r *http.Request) {
	// Verificar que el método sea POST
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Decodificar el request
	var req globals.InstruccionAMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// Verificar que el proceso existe en memoria
	pi, ok := MemoriaGlobal.infoProc[req.PID]
	if !ok {
		http.Error(w, "Proceso no encontrado en memoria", http.StatusNotFound)
		return
	}

	// Obtener las instrucciones (cargándolas si es necesario)
	var instrucciones []string
	if len(pi.Instrucciones) == 0 {
		// Cargar instrucciones desde archivo por primera vez
		filename := Config_Memoria.ScriptsPath + fmt.Sprintf("/%d.instr", req.PID)
		content, err := os.ReadFile(filename)
		if err != nil {
			http.Error(w, "No se encontro el archivo del proceso", http.StatusNotFound)
			return
		}
		instrucciones = strings.Split(strings.TrimSpace(string(content)), "\n")

		// Guardar en memoria para futuras consultas
		pi.Instrucciones = instrucciones
	} else {
		// Usar las instrucciones ya cargadas en memoria
		instrucciones = pi.Instrucciones
	}

	// Validar el PC
	if req.PC < 0 || req.PC >= len(instrucciones) {
		http.Error(w, "PC fuera de rango", http.StatusBadRequest)
		return
	}

	// Preparar y enviar respuesta
	resp := globals.InstruccionAMemoriaResponse{
		InstruccionAEjecutar: instrucciones[req.PC],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	// Log opcional para debugging
	Logger.Debug("Instrucción solicitada", "PID", req.PID, "PC", req.PC, "Instruccion", instrucciones[req.PC])
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

//---------------Case CACHE activada ------------------

//--------------- Case CACHE desactivada ------------------

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

	respuestaRead := MemoriaGlobal.datos[direccionFisica]

	response := globals.CPUReadAMemoriaResponse{
		Respuesta: true,
		Data:      []byte{respuestaRead},
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

// ------------------------------Funciones Auxiliares -----------------------------

// La usamos para el dump memory
func calcularEntradasPorNivel(numPagina int, niveles int, IndicePorNivel int) []int {
	entradas := make([]int, niveles)
	for i := niveles - 1; i >= 0; i-- {
		entradas[i] = numPagina % IndicePorNivel
		numPagina /= IndicePorNivel
	}
	return entradas
}

// *SuspenderPagina guarda la página en swap y libera el frame en RAM.
func (mp *Memoria) SuspenderPagina(pid, pagina int) error {
	pi, ok := mp.infoProc[pid]
	if !ok {
		return fmt.Errorf("proceso %d no existe", pid)
	}

	// Validar que el índice de página sea válido
	if pagina < 0 || pagina >= len(pi.Pages) {
		return fmt.Errorf("página %d fuera de rango para proceso %d", pagina, pid)
	}

	info := &pi.Pages[pagina]
	if !info.InRAM {
		return nil // ya está en swap
	}

	// 1) Leer datos de RAM
	start := info.FrameID * Config_Memoria.PageSize
	buf := make([]byte, Config_Memoria.PageSize)
	copy(buf, mp.datos[start:start+Config_Memoria.PageSize])

	// 2) Escribir en swapfile.bin
	off := mp.nextSwapOffset
	if _, err := mp.swapFile.WriteAt(buf, off); err != nil {
		return fmt.Errorf("error escribiendo swap: %w", err)
	}

	// 3) Liberar frame en bitmap
	mp.liberarFrame(info.FrameID)
	Logger.Debug("Página suspendida exitosamente", "PID", pid, "Pagina", pagina, "SwapOffset", off)

	// 4) Actualizar metadata
	info.InRAM = false
	info.Offset = off
	mp.nextSwapOffset += int64(Config_Memoria.PageSize)
	return nil
}

// *Restaurar pagina desde SWAP
func (mp *Memoria) RestaurarPagina(pid, pagina int) error {
	pi, ok := mp.infoProc[pid]
	if !ok {
		return fmt.Errorf("proceso %d no existe", pid)
	}

	info := &pi.Pages[pagina]
	if info.InRAM {
		return nil // ya está en RAM
	}

	//1) Reservar frame libre
	frameID, err := mp.obtenerFrameLibre()
	if err != nil {
		return fmt.Errorf("no hay frames libres para restaurar página: %w", err)
	}

	//2) Leer datos desde swap

	buffer := make([]byte, Config_Memoria.PageSize)
	if _, err := mp.swapFile.ReadAt(buffer, info.Offset); err != nil {
		mp.liberarFrame(frameID)
		return fmt.Errorf("error leyendo swap: %w", err)
	}

	//3) Escribir en RAM
	start := frameID * Config_Memoria.PageSize
	copy(mp.datos[start:start+Config_Memoria.PageSize], buffer)

	//4) Actualizar metadata
	info.InRAM = true
	info.FrameID = frameID
	// No necesitamos actualizar Offset porque ya no está en swap
	Logger.Debug("Página restaurada exitosamente", "PID", pid, "Pagina", pagina, "FrameID", frameID)

	//5) Insertar en la tabla de páginas
	tablaRaiz := mp.tablas[pid]
	if tablaRaiz != nil {
		mp.insertarEnMultinivel(tablaRaiz, pagina, frameID, 0)
	}
	return nil
}

func listaDeInstrucciones(pid int) []string {
	filename := Config_Memoria.ScriptsPath + fmt.Sprintf("/%d.instr", pid)
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}
	instrucciones := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(instrucciones) == 0 {
		return nil
	}
	return instrucciones
}
