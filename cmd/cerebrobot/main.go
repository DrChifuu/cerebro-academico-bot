// Cerebro Académico — bot de Telegram.
//
// Este binario es la cara del sistema: recibe archivos del usuario, los deja
// en una cola (SQLite) y avisa al worker. No procesa nada él mismo.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const version = "0.1.0"

const startText = "¡Hola! Soy Cerebro-academico_bot. Los comandos son: bla bla bla"

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "falta la variable de entorno TELEGRAM_BOT_TOKEN")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	b, err := bot.New(token, bot.WithMiddlewares(
		soloChatsPermitidos(archivoChatsPermitidos),
		registrarMensajes(archivoRegistro),
	))
	if err != nil {
		log.Fatalf("no se pudo crear el bot: %v", err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, startHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/historial", bot.MatchTypeExact, historialHandler(archivoRegistro))
	b.RegisterHandlerMatchFunc(esArchivo, acuseHandler)

	log.Printf("cerebrobot %s escuchando Telegram", version)
	b.Start(ctx)
}

var comandos = []models.BotCommand{
	{Command: "start", Description: "Muestra el mensaje de bienvenida"},
	{Command: "historial", Description: "Muestra los últimos 10 registros"},
}

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   startText,
	})
	if err != nil {
		log.Printf("error enviando /start: %v", err)
	}

	_, err = b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: comandos,
		Scope:    &models.BotCommandScopeChat{ChatID: update.Message.Chat.ID},
	})
	if err != nil {
		log.Printf("error registrando comandos: %v", err)
	}
}
