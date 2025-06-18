package kernel_internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"utils/client"
	"utils/globals"
)

// &-------------------------------------------Funcion para iniciar Server de Kernel-------------------------------------------------------------

func IniciarServerKernel(puerto int) {
	//Transformo el puerto a string
	stringPuerto := fmt.Sprintf(":%d", puerto)

	//Declaro el server
	mux := http.NewServeMux()

	//^ Handshakes
	mux.HandleFunc("/handshake", HandshakeHandler)
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

	//Escucha el puerto y espera conexiones
	err := http.ListenAndServe(stringPuerto, mux)
	if err != nil {
		panic(err)
	}
}

// &-------------------------------------------Endpoints de Kernel-------------------------------------------------------------

// * Endpoint de handshake = /handshake
func HandshakeHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	//Se utiliza el encoder para enviar la respuesta en formato JSON
	json.NewEncoder(w).Encode(
		map[string]string{
			"modulo":  "Kernel",
			"mensaje": "Conexión aceptada desde " + r.RemoteAddr,
		})
}

// * Endpoint de handshake IO = /handshake/io
func IoHandshakeHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	var dispositivoIOBody globals.IoHandshakeRequest
	if err := json.NewDecoder(r.Body).Decode(&dispositivoIOBody); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	RegistrarInstanciaIO(dispositivoIOBody.IPio, dispositivoIOBody.PortIO, dispositivoIOBody.Nombre)
	Logger.Debug("Handshake con IO realizado", "ip_io", dispositivoIOBody.IPio, "port_io", dispositivoIOBody.PortIO, "nombre", dispositivoIOBody.Nombre)
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
}

// * Endpoint de ping = /ping
func PingHandler(w http.ResponseWriter, r *http.Request) {
	//Establecemos el header de la respuesta (Se indica que la respuesta es de tipo JSON)
	w.Header().Set("Content-Type", "application/json")

	//Se utiliza el encoder para enviar la respuesta en formato JSON
	var respuestaPing = client.PingResponse{
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

	Logger.Debug("Desconexion de IO recibida",
		"nombre_dispositivo", pedidoIO.NombreDispositivo,
		"ip_instancia", pedidoIO.IpInstancia,
		"puerto_instancia", pedidoIO.PuertoInstancia)

	// Procesar la desconexión de IO
	DesconectarInstanciaIO(pedidoIO.NombreDispositivo, pedidoIO.IpInstancia, pedidoIO.PuertoInstancia)
}
