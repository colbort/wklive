package main

import (
	"fmt"
	"log"
	"os"

	"wklive/services/itick/internal/priceengine"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: price-replay <evaluation-audit.json>")
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("read audit: %v", err)
	}
	price, err := priceengine.ReplayEvaluationAudit(raw)
	if err != nil {
		log.Fatalf("replay failed: %v", err)
	}
	fmt.Printf("replay verified: price=%s\n", price)
}
