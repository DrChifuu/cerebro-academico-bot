// Cerebro Académico — worker de procesamiento.
//
// Este binario lee la cola de trabajos y hace el trabajo pesado: extraer texto,
// generar apuntes con IA y archivarlos. Es un programa aparte a propósito:
// si el bot se cae, el trabajo en curso sigue; y viceversa.
package main

import "fmt"

const version = "0.1.0"

func main() {
	fmt.Printf("cerebroworker %s listo (todavía no hay cola)\n", version)
}
