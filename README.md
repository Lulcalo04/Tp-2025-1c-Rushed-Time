# TP Sistemas Operativos 1C2025 – Rushed Time
Trabajo práctico de la cátedra **Sistemas Operativos (UTN FRBA)** desarrollado en **GoLang**🦦. 

## ©️ Autores
- [@TomasJulianoUTN](https://github.com/TomasJulianoUTN) *y creador del readme
- [@LucaLorenzonUTN](https://github.com/LucaLorenzonUTN)
- [@JuanchiNogueira](https://github.com/JuanchiNogueira)
- [@Roccomcs](https://github.com/Roccomcs)
- [@EliCVargasJ](https://github.com/EliCVargasJ)

---

## 📄 Enunciados

- [Enunciado General](https://docs.google.com/document/d/1zoFRoBn9QAfYSr0tITsL3PD6DtPzO2sq9AtvE8NGrkc/edit?tab=t.0#heading=h.xgbc8rcw891t)  
- [Enunciado de Pruebas](https://docs.google.com/document/d/13XPliZvUBtYjaRfuVUGHWbYX8LBs8s3TDdaDa9MFr_I/edit?tab=t.0)  
- [Repositorio de Pruebas](https://github.com/sisoputnfrba/revenge-of-the-cth-pruebas)

---

## Cómo Ejecutarlo

### 1️⃣ Clonar el Repositorio

```bash
git clone https://github.com/Lulcalo04/Tp-2025-1c-Rushed-Time
User: <USUARIO>
Token: <TOKEN>
```

### 2️⃣ Configurar las IPs

```bash
cd tp-2025-1c-Rushed-Time

./configure.sh ip_memory <IP MEMORIA>
./configure.sh ip_kernel <IP KERNEL>
./configure.sh ip_cpu <IP CPU>
./configure.sh ip_io <IP IO>
```

### 3️⃣ Ejecutar las Pruebas

---

#### 🔹 Pruebas de Corto Plazo

**FIFO**
```bash
go run memoria/memoria.go memoria_cortoplazo
go run kernel/kernel.go kernel_fifo_cortoplazo PLANI_CORTO_PLAZO 0
go run cpu/cpu.go cpu1_cortoplazo 1
go run cpu/cpu.go cpu2_cortoplazo 2
go run io/io.go config_io1 DISCO
go run io/io.go config_io2 DISCO
```

**SJF**
```bash
go run memoria/memoria.go memoria_cortoplazo
go run kernel/kernel.go kernel_sjf_cortoplazo PLANI_CORTO_PLAZO 0
go run cpu/cpu.go cpu1_cortoplazo 1
go run io/io.go config_io1 DISCO
```

**SRT**
```bash
go run memoria/memoria.go memoria_cortoplazo
go run kernel/kernel.go kernel_srt_cortoplazo PLANI_CORTO_PLAZO 0
go run cpu/cpu.go cpu1_cortoplazo 1
go run io/io.go config_io1 DISCO
```

---

#### 🔹 Pruebas de Largo y Mediano Plazo

**FIFO**
```bash
go run memoria/memoria.go memoria_lym
go run kernel/kernel.go kernel_fifo_lym PLANI_LYM_PLAZO 0
go run cpu/cpu.go cpu_lym 1
go run io/io.go config_io1 DISCO
```

**PMCP**
```bash
go run memoria/memoria.go memoria_lym
go run kernel/kernel.go kernel_pmcp_lym PLANI_LYM_PLAZO 0
go run cpu/cpu.go cpu_lym 1
go run io/io.go config_io1 DISCO
```

---

#### 🔹 Prueba de SWAP

```bash
go run memoria/memoria.go memoria_swap
go run kernel/kernel.go kernel_swap MEMORIA_IO 90
go run cpu/cpu.go cpu_swap 1
go run io/io.go config_io1 DISCO
```

> ⚠️ **Recordatorio:** ¡Borrar `swapfile.bin` antes de cada ejecución!

---

#### 🔹 Prueba de Caché

**CLOCK**
```bash
go run memoria/memoria.go memoria_cache
go run kernel/kernel.go kernel_cache MEMORIA_BASE 256
go run cpu/cpu.go cpu_clock_cache 1
go run io/io.go config_io1 DISCO
```

**CLOCK-M**
```bash
go run memoria/memoria.go memoria_cache
go run kernel/kernel.go kernel_cache MEMORIA_BASE 256
go run cpu/cpu.go cpu_clockm_cache 1
go run io/io.go config_io1 DISCO
```

---

#### 🔹 Prueba de TLB

**FIFO**
```bash
go run memoria/memoria.go memoria_tlb
go run kernel/kernel.go kernel_tlb MEMORIA_BASE_TLB 256
go run cpu/cpu.go cpu_fifo_tlb 1
go run io/io.go config_io1 DISCO
```

**LRU**
```bash
go run memoria/memoria.go memoria_tlb
go run kernel/kernel.go kernel_tlb MEMORIA_BASE_TLB 256
go run cpu/cpu.go cpu_lru_tlb 1
go run io/io.go config_io1 DISCO
```

> ⚠️ **Recordatorio:** ¡Borrar `swapfile.bin` y los dumps antes de cada ejecución!

---

#### 🔹 Prueba de Estabilidad General

```bash
go run memoria/memoria.go memoria_estabilidad
go run kernel/kernel.go kernel_estabilidad ESTABILIDAD_GENERAL 0
go run cpu/cpu.go cpu1_estabilidad 1
go run cpu/cpu.go cpu2_estabilidad 2
go run cpu/cpu.go cpu3_estabilidad 3
go run cpu/cpu.go cpu4_estabilidad 4
go run io/io.go config_io1 DISCO
go run io/io.go config_io2 DISCO
go run io/io.go config_io3 DISCO
go run io/io.go config_io4 DISCO
```

---

## 🛠️ Comandos Útiles

```bash
# Eliminar un archivo
rm nombre_archivo

# Eliminar todo el repositorio clonado
rm -rf tp-2025-1c-Rushed-Time

# Visualizar contenido de swapfile.bin
hexdump -C swapfile.bin
```
