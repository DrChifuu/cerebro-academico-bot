package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const archivoChatsPermitidos = "ALLOWED_ID.md"

func soloChatsPermitidos(ruta string) bot.Middleware {
	return func(siguiente bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message == nil {
				return
			}
			chatID := update.Message.Chat.ID

			permitidos, err := cargarChatsPermitidos(ruta)
			if err != nil {
				log.Printf("no se pudo leer %s, se ignora el mensaje: %v", ruta, err)
				return
			}
			if !permitidos[chatID] {
				log.Printf("chat no autorizado %d, se ignora (añádelo a %s para permitirlo)", chatID, ruta)
				return
			}

			siguiente(ctx, b, update)
		}
	}
}

func cargarChatsPermitidos(ruta string) (map[int64]bool, error) {
	datos, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}

	permitidos := make(map[int64]bool)
	for _, linea := range strings.Split(string(datos), "\n") {
		linea = strings.TrimSpace(linea)
		if linea == "" || strings.HasPrefix(linea, "#") {
			continue
		}
		id, err := strconv.ParseInt(linea, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ID inválido %q: %w", linea, err)
		}
		permitidos[id] = true
	}
	return permitidos, nil
}
