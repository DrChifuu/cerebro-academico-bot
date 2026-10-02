# 🎓 Cerebro Académico Bot

Reconstrucción desde cero del bot **Cerebro Académico**: un sistema que recibe
audios de clases, PDFs, fotos y videos por Telegram, los transcribe, genera
apuntes con IA y los archiva organizados por ramo en tu nube personal.

> Este repo es a la vez **proyecto funcional** y **cuaderno de aprendizaje de Go**.
> Cada etapa del historial de commits es una lección: se lee de atrás hacia
> adelante como un curso.

---

## Cómo funciona

```
   tú (Telegram)
        │  audio / PDF / foto
        ▼
┌───────────────────┐        SQLite         ┌──────────────────────┐
│    cerebrobot     │ ──── tabla jobs ────▶ │   cerebroworker      │
│  recibe y cola    │      (la cola)        │  extrae, IA, archiva │
└───────────────────┘                       └──────────────────────┘
                                                        │
                              ┌─────────────────────────┼──────────────────┐
                              ▼                         ▼                  ▼
                        transcripción              apuntes .md        indexa RAG
                        (Whisper/OCR)           (resumen+preguntas)  ( ChromaDB )
```

Dos binarios independientes que comparten una biblioteca `internal/`:

| Binario | Responsabilidad |
|---|---|
| `cerebrobot` | Long-polling de Telegram, descarga archivos, pide el ramo con botones inline, consulta el historial |
| `cerebroworker` | Procesa la cola: extrae texto → clasifica ramo → genera apuntes con LLM → sube a la nube → indexa para RAG |

**Por qué dos procesos:** son dos ritmos distintos. El bot debe responder
milisegundos; el worker tarda minutos transcribiendo un audio. Separarlos
significa que un cuello de botella en IA nunca bloquea el chat, y si uno cae,
el otro sigue (la cola SQLite sobrevive a los reinicios).

---

## Estructura

```
cerebro-academico-bot/
├── go.mod                 módulo: github.com/DrChifuu/cerebro-academico-bot
├── cmd/
│   ├── cerebrobot/        binario 1: interfaz de Telegram
│   └── cerebroworker/     binario 2: motor de procesamiento
└── internal/              biblioteca privada (el compilador impide importarla fuera)
    ├── config/            config.yaml + variables de entorno
    ├── store/             SQLite: la cola de trabajos
    ├── bot/               handlers de Telegram
    ├── extract/           texto desde audio/PDF/imagen/office
    ├── llm/               cliente de LLM compatible con OpenAI
    ├── notes/             generación de apuntes
    ├── cloud/             archivado en la nube personal
    └── rag/               chunks + búsqueda semántica
```

`internal/` en lugar de `shared/` (el nombre que usa el original): en Go,
`internal/` es **una garantía del compilador**, no solo un nombre.

---

## Requisitos

- Go ≥ 1.26 (`go version`)
- SQLite (vía `go-sqlite3`, requiere CGO y un compilador C)
- Para el worker: `ffmpeg`, `pdftotext`, `tesseract`, `pdftoppm`
- Un token de bot de [@BotFather](https://t.me/BotFather)

## Compilar y correr

```bash
go vet ./...                       # análisis estático
gofmt -l .                         # debe salir vacío

go build -o bin/cerebrobot    ./cmd/cerebrobot
go build -o bin/cerebroworker ./cmd/cerebroworker

cp .env.example .env              # pon tu token
set -a; . ./.env; set +a          # cargar el entorno
./bin/cerebrobot
```

---

## Etapas del aprendizaje

| # | Etapa | Conceptos Go |
|---|---|---|
| 0 | Estructura del proyecto | módulos, `cmd/`, `internal/`, `go vet` |
| 1 | Configuración | structs, punteros, tags YAML, precedencia env > archivo |
| 2 | La cola (SQLite) | `database/sql`, transacciones, errores `%w` |
| 3 | Bot de Telegram | canales, `for range`, métodos, `defer` |
| 4 | Worker y señales | goroutines, `atomic`, SIGTERM, `recover` |
| 5 | Extracción de texto | `os/exec`, `io`, `defer` |
| 6 | LLM | HTTP client, JSON, `context` y timeouts |
| 7 | Archivado en la nube | `os`, rutas, manejo de errores |
| 8 | RAG | slices, maps, interfaces |
| 9 | Tests y documentación | `testing`, benchmarks, README final |

---

## Seguridad

- **Nunca** se commitea `.env`: contiene el token de Telegram y API keys.
- `ALLOWED_CHAT_IDS` restringe quién puede usar el bot (vacío = abierto).
- Ver `.gitignore` para la lista completa de lo que queda fuera del repo.

## Licencia

MIT
