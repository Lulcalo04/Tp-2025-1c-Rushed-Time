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
	mux.HandleFunc("/syscall/swappeo", SwappingHandler) // ! PREGUNTAR A LOS CHICOS COMO TIENEN ESTOS ENDPOINTS
	mux.HandleFunc("/syscall/restaurar", RestaurarHandler) // ! PREGUNTAR A LOS CHICOS COMO TIENEN ESTOS ENDPOINTS
	mux.HandleFunc("/cpu/instrucciones", InstruccionesHandler)
	mux.HandleFunc("/cpu/frame", CalcularFrameHandler) // pedido de frame desde CPU para la traduccion de direcciones
	mux.HandleFunc("/cpu/write", HacerWriteHandler)
	mux.HandleFunc("/cpu/read", HacerReadHandler)
	mux.HandleFunc("/dump", DumpMemoryHandler)
	mux.HandleFunc("/syscall/actualizarPagina",ActualizarPaginahandler) // ! PREGUNTAR A LOS CHICOS COMO TIENEN ESTOS ENDPOINTS
	

	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

// -------------------------------------------  endpoints generales del modulo Memoria-------------------------------------------------------------

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

// -----------------------------------------------Funciones de pedido/liberacion de Espacio-------------------------------------------------------------

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
			Metricas: &MetricasPorProceso{}, // Inicializa con el valor cero del tipo adecuado
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
		//TODO -- Log Obligatorio 
		LogCreacionDeProceso(pedidoRecibido.ProcesoPCB.PID, pedidoRecibido.ProcesoPCB.TamanioEnMemoria)

		// Preparar respuesta y codificarla como JSON (se envia automaticamente a traves del encode)
		resp := globals.PeticionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: true,
			Mensaje:   fmt.Sprintf("Espacio concedido para PID %d con %d de espacio", pedidoRecibido.ProcesoPCB.PID, pedidoRecibido.ProcesoPCB.TamanioEnMemoria),
		}
		
		//& Actualizacion de metricas
		MemoriaGlobal.infoProc[pedidoRecibido.ProcesoPCB.PID].Metricas.AccesoATablaDePaginas++

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

// *Endopoint de liberacion de espacio = /espacio/liberar
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

	LogDestruccionDeProceso(pedidoRecibido.PID, *MemoriaGlobal.infoProc[pedidoRecibido.PID].Metricas)

	pid := pedidoRecibido.PID
	err := MemoriaGlobal.LiberarProceso(pid)

	w.Header().Set("Content-Type", "application/json") // Setear Content-Type JSON (osea que la respuesta sera del tipo JSON)

	if err == nil {
		Logger.Debug("Liberacion de espacio aceptada", "PID", pid)
		respuesta := globals.LiberacionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: true,
			Mensaje:   fmt.Sprintf("Liberacion de espacio en memoria aceptada para PID %d", pid),
		}
		json.NewEncoder(w).Encode(respuesta)
	} else {
		Logger.Debug("Liberacion de espacio fallida", "PID", pid)
		respuesta := globals.LiberacionMemoriaResponse{
			Modulo:    "Memoria",
			Respuesta: false,
			Mensaje:   fmt.Sprintf("Error liberando memoria para PID %d: %v", pid, err),
		}
		json.NewEncoder(w).Encode(respuesta)
	}

	
}

func (m *Memoria) LiberarProceso(pid int) error {
	pi, ok := m.infoProc[pid]
	if !ok {
		return fmt.Errorf("el proceso %d no existe", pid)
	}

	// Liberar frames en RAM o marcar offset SWAP como libre
	for _, page := range pi.Pages {
		if page.InRAM {
			m.liberarFrame(page.FrameID)
		} else {
			m.freeSwapOffsets = append(m.freeSwapOffsets, page.Offset)
		}
	}

	// Borrar metadata
	delete(m.infoProc, pid)

	return nil
}

// -----------------------------------------------Funciones de Dump Memory-------------------------------------------------------------

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

		// TODO -- Log Obligatorio 
		LogMemoryDump(pedidoRecibido.PID)

		//oppcional: Logger.Debug("Solicitud de Dump Memory aceptada", "PID", pedidoRecibido.PID)

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

// -----------------------------------------------Funciones de Swappeo-------------------------------------------------------------

// * Endpoint de swappeo = /syscall/swappeo
func SwappingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		PID int `json:"pid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	pi, ok := MemoriaGlobal.infoProc[request.PID]
	if !ok {
		http.Error(w, "Proceso no existe", http.StatusNotFound)
		return
	}
	
	// Suspender todas sus páginas
	for i := range pi.Pages {
		if err := MemoriaGlobal.SuspenderPagina(request.PID, i); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"modulo": "Memoria", "status": "suspendido",
	})
	
	// & Actualizar las métricas del proceso
	MemoriaGlobal.infoProc[request.PID].Metricas.BajadasASwap++

	Logger.Debug("Proceso swappeado exitosamente", "PID", request.PID)
}

// *Endpoint de restauracion = /syscall/restaurar
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

	//& Metricas por proceso
	MemoriaGlobal.infoProc[req.PID].Metricas.SubidasAMemoriaPrincipal++
}

// SuspenderPagina guarda la página en swap y libera el frame en RAM. (funcion Auxiliar)
func (mp *Memoria) SuspenderPagina(pid, pagina int) error {
	// 1) Validar proceso existe
	pi, ok := mp.infoProc[pid]
	if !ok {
		return fmt.Errorf("proceso %d no existe", pid)
	}

	// 2) Validar índice de página ANTES de hacer cualquier cosa
	if pagina < 0 || pagina >= len(pi.Pages) {
		return fmt.Errorf("página %d fuera de rango para proceso %d", pagina, pid)
	}

	page := &pi.Pages[pagina]

	// 3) Verificar si ya está en swap
	if !page.InRAM {
		return nil // ya está en swap
	}

	// 4) Reservar offset (el tamanio) en swap SOLO si la página está en RAM
	var offset int64
	if len(mp.freeSwapOffsets) > 0 {
		// Si hay offsets libres, usar el último (más eficiente)
		offset = mp.freeSwapOffsets[len(mp.freeSwapOffsets)-1]
		mp.freeSwapOffsets = mp.freeSwapOffsets[:len(mp.freeSwapOffsets)-1]
	} else {
		// Usar nuevo offset
		offset = mp.nextSwapOffset
		mp.nextSwapOffset += int64(Config_Memoria.PageSize)
	}

	// 5) Leer datos de RAM
	start := page.FrameID * Config_Memoria.PageSize
	buf := make([]byte, Config_Memoria.PageSize)
	copy(buf, mp.datos[start:start+Config_Memoria.PageSize])

	// 6) Escribir en swap
	if _, err := mp.swapFile.WriteAt(buf, offset); err != nil {
		// Si falla, devolver el offset libre si era reutilizado
		if offset < mp.nextSwapOffset {
			mp.freeSwapOffsets = append(mp.freeSwapOffsets, offset)
		}
		return fmt.Errorf("error escribiendo swap: %w", err)
	}

	// 7) Liberar frame en bitmap
	mp.liberarFrame(page.FrameID)

	// 8) Actualizo la info del proceso
	page.InRAM = false
	page.Offset = offset
	

	Logger.Debug("Página suspendida exitosamente", "PID", pid, "Pagina", pagina, "SwapOffset", offset)
	return nil
}

// Restaurar pagina desde SWAP (funcion Auxiliar )
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

	//4) Actualizar los valores del proceso
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

// -----------------------------------------------Funcion de instrucciones------------------------------------------------

// * Endpoint de instrucciones = /instrucciones
func InstruccionesHandler(w http.ResponseWriter, r *http.Request) {
	// Verificar que el método sea POST
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Decodificar el request
	var request globals.InstruccionAMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// Verificar que el proceso existe en memoria
	pi, ok := MemoriaGlobal.infoProc[request.PID]
	if !ok {
		http.Error(w, fmt.Sprintf("Proceso con PID: %d no encontrado en memoria", request.PID), http.StatusNotFound)
		return
	}

	// Obtener las instrucciones (cargándolas si es necesario)
	var instrucciones []string
	if len(pi.Instrucciones) == 0 {
		// Cargar instrucciones desde archivo por primera vez
		filename := Config_Memoria.ScriptsPath + request.Path
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
	if request.PC < 0 || request.PC >= len(instrucciones) {
		http.Error(w, "PC fuera de rango", http.StatusBadRequest)
		return
	}

	// Preparar y enviar respuesta
	resp := globals.InstruccionAMemoriaResponse{
		InstruccionAEjecutar: instrucciones[request.PC],
	}

	//& Actualizo metricas: 
	MemoriaGlobal.infoProc[request.PID].Metricas.AccesoATablaDePaginas++

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	// TODO -- log obligatorio 
	LogObtenerInstruccion(request.PID, request.PC, instrucciones[request.PC] )
	// opcional : Logger.Debug("Instrucción solicitada", "PID", request.PID, "PC", request.PC, "Instruccion", instrucciones[request.PC])
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

//-------------------------------------------------Funciones para CPU ------------------------------------------------

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

//--------------Caso CACHE activada -----------------
// *Endpoint de actualizar pagina completa = /syscall/actualizarPagina
func ActualizarPaginahandler(w http.ResponseWriter, r *http.Request) {
	var request globals.CPUActualizarPaginaEnMemoriaRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	processInfo, ok := MemoriaGlobal.infoProc[request.PID]
	if !ok || processInfo.TablaRaiz == nil {
		http.Error(w, fmt.Sprintf("Proceso con PID: %d no encontrado en memoria", request.PID), http.StatusNotFound)
		return
	}
	if request.NumeroDePagina < 0 || request.NumeroDePagina >= len(processInfo.Pages) {
		http.Error(w, fmt.Sprintf("Página del proceso %d fuera de rango", request.PID), http.StatusBadRequest)
		return
	}
	
	page := &processInfo.Pages[request.NumeroDePagina]
	frameID := page.FrameID

	//verifico que el frameID sea valido (creo que es redundante, pero por las dudas)
	if frameID < 0 || frameID*Config_Memoria.PageSize >= len(MemoriaGlobal.datos) {
		http.Error(w, fmt.Sprintf("FrameID del proceso %d inválido", request.PID), http.StatusInternalServerError)
		return
	}
    
	// Actualizo el contenido de la página en memoria

	offset := frameID * Config_Memoria.PageSize
    copy(MemoriaGlobal.datos[offset:offset+len(request.Data)], request.Data)

	// & Metricas 
	MemoriaGlobal.infoProc[request.PID].Metricas.AccesoATablaDePaginas++
	MemoriaGlobal.infoProc[request.PID].Metricas.EscriturasDeMemoria++

	response := globals.CPUActualizarPaginaEnMemoriaResponse{
		Respuesta: true,
		Frame: frameID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

//--------------- Case CACHE desactivada -------------

// * Endpoint de read = /cpu/read
func HacerReadHandler(w http.ResponseWriter, r *http.Request) {
	var request globals.CPUReadAMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	tablaRaiz := MemoriaGlobal.tablas[request.PID]
	if tablaRaiz == nil {
		http.Error(w, fmt.Sprintf("Proceso con PID %d  no encontrado en memoria", request.PID), http.StatusNotFound)
		return
	}

	var direccionFisica = request.DireccionFisica
	tamanio:= request.Tamanio

	if direccionFisica < 0 || tamanio <=0 || direccionFisica >= len(MemoriaGlobal.datos) {
		http.Error(w, "la direccion excede del tamanio direccionable", http.StatusBadRequest)
		return
	}

	respuestaRead := MemoriaGlobal.datos[direccionFisica : direccionFisica + tamanio]

	response := globals.CPUReadAMemoriaResponse{
		Respuesta: true,
		Data:      respuestaRead,
	}

	//& Actualizo las metricas del proceso
	MemoriaGlobal.infoProc[request.PID].Metricas.AccesoATablaDePaginas++
	MemoriaGlobal.infoProc[request.PID].Metricas.LecturasDeMemoria++

	LogOperacionEnEspacioUsuario(request.PID, "read", direccionFisica, tamanio)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// * Endpoint de write = /cpu/write
func HacerWriteHandler(w http.ResponseWriter, r *http.Request) {

	var request globals.CPUWriteAMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	respuestaWrite := false

	tablaRaiz := MemoriaGlobal.tablas[request.PID]
		if tablaRaiz == nil {
		http.Error(w, fmt.Sprintf("Proceso con PID %d  no encontrado en memoria", request.PID), http.StatusNotFound)
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

	// &Actualizo las metricas del proceso
	MemoriaGlobal.infoProc[request.PID].Metricas.AccesoATablaDePaginas++
	MemoriaGlobal.infoProc[request.PID].Metricas.EscriturasDeMemoria++

	LogOperacionEnEspacioUsuario(request.PID, "write", direccionFisica, len(data))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ------------------------------------------------Funciones Auxiliares ---------------------------------------------

// La usamos para el dump memory
func calcularEntradasPorNivel(numPagina int, niveles int, IndicePorNivel int) []int {
	entradas := make([]int, niveles)
	for i := niveles - 1; i >= 0; i-- {
		entradas[i] = numPagina % IndicePorNivel
		numPagina /= IndicePorNivel
	}
	return entradas
}

//Informacion de las metricas por proceso 

