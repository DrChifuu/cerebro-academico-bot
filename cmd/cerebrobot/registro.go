package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const archivoRegistro = "registro.md"

const encabezadoRegistro = "| Fecha | Chat | Tipo | Detalle |\n| --- | --- | --- | --- |\n"

const maxHistorial = 10

var mutexRegistro sync.Mutex

func registrarMensajes(ruta string) bot.Middleware {
	return func(siguiente bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message != nil {
				if err := agregarRegistro(ruta, update.Message); err != nil {
					log.Printf("no se pudo escribir en %s: %v", ruta, err)
				}
			}
			siguiente(ctx, b, update)
		}
	}
}

func agregarRegistro(ruta string, msg *models.Message) error {
	mutexRegistro.Lock()
	defer mutexRegistro.Unlock()

	f, err := os.OpenFile(ruta, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		if _, err := f.WriteString(encabezadoRegistro); err != nil {
			return err
		}
	}

	fecha := time.Unix(int64(msg.Date), 0).Format("2006-01-02 15:04:05")
	tipo, detalle := describirMensaje(msg)
	fila := fmt.Sprintf("| %s | %d | %s | %s |\n", fecha, msg.Chat.ID, tipo, escaparCelda(detalle))
	_, err = f.WriteString(fila)
	return err
}

func describirMensaje(msg *models.Message) (string, string) {
	switch {
	case msg.Document != nil:
		return "documento", unirDetalle(msg.Document.FileName, msg.Caption)
	case len(msg.Photo) > 0:
		return "foto", msg.Caption
	case msg.Audio != nil:
		return "audio", msg.Caption
	case msg.Video != nil:
		return "video", msg.Caption
	case msg.Voice != nil:
		return "voz", msg.Caption
	case strings.HasPrefix(msg.Text, "/"):
		return "comando", msg.Text
	case msg.Text != "":
		return "texto", msg.Text
	default:
		return "otro", ""
	}
}

func esArchivo(update *models.Update) bool {
	m := update.Message
	return m != nil && (m.Document != nil || len(m.Photo) > 0 || m.Audio != nil || m.Video != nil || m.Voice != nil)
}

func acuseHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	tipo, detalle := describirMensaje(update.Message)
	texto := "Recibido: " + tipo
	if detalle != "" {
		texto += " - " + detalle
	}
	texto += ". Quedó registrado; el archivo todavía no se guarda."

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   texto,
	})
	if err != nil {
		log.Printf("error enviando acuse: %v", err)
	}
}

func unirDetalle(partes ...string) string {
	var con []string
	for _, p := range partes {
		if p != "" {
			con = append(con, p)
		}
	}
	return strings.Join(con, " - ")
}

func escaparCelda(texto string) string {
	texto = strings.NewReplacer("\r", " ", "\n", " ").Replace(texto)
	return strings.ReplaceAll(texto, "|", `\|`)
}

func historialHandler(ruta string) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		texto, err := textoHistorial(ruta, maxHistorial)
		if err != nil {
			log.Printf("no se pudo leer %s: %v", ruta, err)
			texto = "No se pudo leer el historial."
		}

		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   texto,
		})
		if err != nil {
			log.Printf("error enviando /historial: %v", err)
		}
	}
}

func textoHistorial(ruta string, n int) (string, error) {
	filas, err := leerUltimasFilas(ruta, n)
	if err != nil {
		return "", err
	}
	if len(filas) == 0 {
		return "Todavía no hay registros.", nil
	}

	var sb strings.Builder
	sb.WriteString("Registros más recientes:\n")
	for i := len(filas) - 1; i >= 0; i-- {
		fila := filas[i]
		detalle := fila[3]
		if detalle == "" {
			detalle = "(sin texto)"
		}
		fmt.Fprintf(&sb, "%d. %s · %s · %s\n", len(filas)-i, fila[0], fila[2], detalle)
	}
	return sb.String(), nil
}

func leerUltimasFilas(ruta string, n int) ([][]string, error) {
	mutexRegistro.Lock()
	defer mutexRegistro.Unlock()

	datos, err := os.ReadFile(ruta)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var filas [][]string
	for _, linea := range strings.Split(string(datos), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(linea), "|") {
			continue
		}
		celdas := separarCeldas(linea)
		if len(celdas) != 4 || celdas[0] == "Fecha" || strings.HasPrefix(celdas[0], "---") {
			continue
		}
		filas = append(filas, celdas)
	}
	if len(filas) > n {
		filas = filas[len(filas)-n:]
	}
	return filas, nil
}

func separarCeldas(linea string) []string {
	linea = strings.TrimSpace(linea)
	linea = strings.TrimPrefix(linea, "|")
	linea = strings.TrimSuffix(linea, "|")

	var celdas []string
	var actual strings.Builder
	for i := 0; i < len(linea); i++ {
		c := linea[i]
		if c == '\\' && i+1 < len(linea) && linea[i+1] == '|' {
			actual.WriteByte('|')
			i++
			continue
		}
		if c == '|' {
			celdas = append(celdas, strings.TrimSpace(actual.String()))
			actual.Reset()
			continue
		}
		actual.WriteByte(c)
	}
	return append(celdas, strings.TrimSpace(actual.String()))
}
