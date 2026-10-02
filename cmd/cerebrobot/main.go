// Cerebro Académico — bot de Telegram.
//
// Este binario es la cara del sistema: recibe archivos del usuario, los deja
// en una cola (SQLite) y avisa al worker. No procesa nada él mismo.
package main

import "fmt"

const version = "0.1.0"

func main() {
	fmt.Printf("cerebrobot %s listo (todavía no escucha Telegram)\n", version)
}
