package kernel_internal

import (
	"encoding/json"
	"fmt"
	"globals"
	"net/http"
)

// &-------------------------------------------Funcion para iniciar Server de Kernel-------------------------------------------------------------

func IniciarServerKernel(puerto int) {
	//Transformo el puerto a string
	stringPuerto := fmt.Sprintf(":%d", puerto)

	//Declaro el server
	mux := http.NewServeMux()

	//^ Handshakes
	mux.HandleFunc("/handshake/io", IoHandshakeHandler)
	mux.HandleFunc("/handshake/cpu", CPUHandshakeHandler)

	//^ Ping
	//* Este endpoint es utilizado por el cliente para verificar que el kernel está activo
	mux.HandleFunc("/ping", PingHandler)

	//^ Syscalls
	mux.HandleFunc("/syscall/init_proc", InitProcHandler)
	mux.HandleFunc("/syscall/exit", ExitHandler)
	mux.HandleFunc("/syscall/dump_memory", DumpMemoryHandler)
	mux.HandleFunc("/syscall/io", IoHandler)

	//^ Handlers para IO
	mux.HandleFunc("/io/fin", FinIOHandler)
	mux.HandleFunc("/io/desconexion", DesconexionIOHandler)

	//^ Handlers para CPU
	mux.HandleFunc("/cpu/desalojo", DesalojoHandler)

	//Escucha el puerto y espera conexiones
	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

// &-------------------------------------------Endpoints de Kernel-------------------------------------------------------------

// * Endpoint de handshake IO = /handshake/io
func IoHandshakeHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	var dispositivoIOBody globals.IoHandshakeRequest
	if err := json.NewDecoder(r.Body).Decode(&dispositivoIOBody); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}
	// Respondo a IO que ser recibió el mensaje correctamente
	var respuesta = globals.IoHandshakeResponse{
		Respuesta: true,
		Mensaje:   "Handshake realizado, dispositivo registrado en KERNEL correctamente.",
	}

	json.NewEncoder(w).Encode(respuesta)

	go RegistrarInstanciaIO(dispositivoIOBody.Nombre, dispositivoIOBody.PortIO, dispositivoIOBody.IPio)

	mensajeHandshakeIO := fmt.Sprintf("Handshake con IO realizado: IP %s, Puerto %d, Nombre %s", dispositivoIOBody.IPio, dispositivoIOBody.PortIO, dispositivoIOBody.Nombre)
	fmt.Println(mensajeHandshakeIO)
	Logger.Debug(mensajeHandshakeIO)
}

// * Endpoint de handshake CPU = /handshake/CPU
func CPUHandshakeHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	var dispositivoCPUBody globals.CPUToKernelHandshakeRequest
	if err := json.NewDecoder(r.Body).Decode(&dispositivoCPUBody); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	bodyRespuesta := RegistrarIdentificadorCPU(dispositivoCPUBody.CPUID, dispositivoCPUBody.Puerto, dispositivoCPUBody.Ip)

	json.NewEncoder(w).Encode(bodyRespuesta)

	mensajeHandshakeCPU := fmt.Sprintf("Handshake con CPU realizado: IP %s, Puerto %d, ID %s", dispositivoCPUBody.Ip, dispositivoCPUBody.Puerto, dispositivoCPUBody.CPUID)
	fmt.Println(mensajeHandshakeCPU)
	Logger.Debug(mensajeHandshakeCPU)
}

// * Endpoint de ping = /ping
func PingHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	//Se utiliza el encoder para enviar la respuesta en formato JSON
	var respuestaPing = globals.PingResponse{
		Modulo:  "Kernel",
		Mensaje: "Pong",
	}

	json.NewEncoder(w).Encode(respuestaPing)
}

// &-------------------------------------------Endpoints de Syscalls/Kernel-------------------------------------------------------------

// * Endpoint de InitProc = /syscall/init_proc
func InitProcHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var PeticionSyscall globals.InitProcSyscallRequest
	if err := json.NewDecoder(r.Body).Decode(&PeticionSyscall); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}
	var respuestaInitProc = globals.InitProcSyscallResponse{
		Respuesta: true,
	}
	json.NewEncoder(w).Encode(respuestaInitProc)

	SyscallInitProc(PeticionSyscall.PID, PeticionSyscall.NombreArchivo, PeticionSyscall.Tamanio)
}

// * Endpoint de Exit = /syscall/exit
func ExitHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var PeticionSyscall globals.ExitSyscallRequest
	if err := json.NewDecoder(r.Body).Decode(&PeticionSyscall); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	// Respondo a CPU que ser recibió el mensaje correctamente
	var respuesta = globals.ExitSyscallResponse{
		Respuesta: true,
	}
	json.NewEncoder(w).Encode(respuesta)

	SyscallExit(PeticionSyscall.PID)
}

// * Endpoint de DumpMemory = /syscall/dump_memory
func DumpMemoryHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var PeticionSyscall globals.DumpMemorySyscallRequest
	if err := json.NewDecoder(r.Body).Decode(&PeticionSyscall); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	// Respondo a CPU que ser recibió el mensaje correctamente
	var respuesta = globals.DumpMemorySyscallResponse{
		Respuesta: true,
	}
	json.NewEncoder(w).Encode(respuesta)

	SyscallDumpMemory(PeticionSyscall.PID)
}

// * Endpoint de IO = /syscall/io
func IoHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var PeticionSyscall globals.IoSyscallRequest
	if err := json.NewDecoder(r.Body).Decode(&PeticionSyscall); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	// Respondo a CPU que ser recibió el mensaje correctamente
	var respuesta = globals.IoSyscallResponse{
		Respuesta: true,
	}
	json.NewEncoder(w).Encode(respuesta)

	SyscallEntradaSalida(PeticionSyscall.PID, PeticionSyscall.NombreDispositivo, PeticionSyscall.Tiempo)
}

// & -------------------------------------------Handlers para IO-------------------------------------------------------------
// * Endpoint fin de IO = /io/fin
func FinIOHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Decodificar el body
	var respuestaIO globals.IOResponse
	if err := json.NewDecoder(r.Body).Decode(&respuestaIO); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	mensajePedidoFinIO := fmt.Sprintf("Fin de IO recibido: Dispositivo %s, PID %d, Respuesta %t", respuestaIO.NombreDispositivo, respuestaIO.PID, respuestaIO.Respuesta)
	fmt.Println(mensajePedidoFinIO)
	Logger.Debug(mensajePedidoFinIO)

	Logger.Debug("Finalización de IO recibida",
		"nombre_dispositivo", respuestaIO.NombreDispositivo,
		"pid", respuestaIO.PID,
		"respuesta", respuestaIO.Respuesta,
	)

	ProcesarFinIO(respuestaIO.PID, respuestaIO.NombreDispositivo)

	Logger.Debug("Fin de IO procesado",
		"nombre_dispositivo", respuestaIO.NombreDispositivo,
		"pid", respuestaIO.PID,
		"respuesta", respuestaIO.Respuesta)
}

// * Endpoint de desconexión de IO = /io/desconexion
func DesconexionIOHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Decodificar el body
	var pedidoIO globals.IOtoKernelDesconexionRequest
	if err := json.NewDecoder(r.Body).Decode(&pedidoIO); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	mensajeDesconexion := fmt.Sprintf("Desconexion de IO recibida: Dispositivo %s, IP %s, Puerto %d", pedidoIO.NombreDispositivo, pedidoIO.IpInstancia, pedidoIO.PuertoInstancia)
	fmt.Println(mensajeDesconexion)
	Logger.Debug(mensajeDesconexion)

	// Procesar la desconexión de IO
	go DesconectarInstanciaIO(pedidoIO.NombreDispositivo, pedidoIO.IpInstancia, pedidoIO.PuertoInstancia)
}

// & -------------------------------------------Handlers para CPU-------------------------------------------------------------

func DesalojoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Decodificar el body
	var respuestaDesalojo globals.CPUtoKernelDesalojoRequest
	if err := json.NewDecoder(r.Body).Decode(&respuestaDesalojo); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	mensajeDesalojo := fmt.Sprintf("Desalojo recibido de CPU: ID %s, PID %d, PC %d, Motivo %s", respuestaDesalojo.CPUID, respuestaDesalojo.PID, respuestaDesalojo.PC, respuestaDesalojo.Motivo)
	fmt.Println(mensajeDesalojo)
	Logger.Debug(mensajeDesalojo)

	// Respondo a CPU que ser recibió el mensaje correctamente
	var respuesta = globals.CPUtoKernelDesalojoResponse{
		Respuesta: true,
	}
	json.NewEncoder(w).Encode(respuesta)

	// Verifica si se desaloja por: Planificador (SJF CD), IO, o por fin de proceso
	// Dependiendo el motivo, se enviará el proceso a la cola correspondiente
	AnalizarDesalojo(respuestaDesalojo.CPUID, respuestaDesalojo.PID, respuestaDesalojo.PC, respuestaDesalojo.Motivo)
}
