package main

import (
	"os"

	"github.com/MrAndreID/goapi/v2/internal/application"
)

func main() {
	if application.Start(true) == nil {
		os.Exit(1)
	}
}
