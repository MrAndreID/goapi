package configs

import (
	"fmt"

	"github.com/common-nighthawk/go-figure"
)

func LoadVersion(cfg *Config, toggle bool) {
	if toggle {
		figure.NewFigure("MrAndreID", "standard", true).Print()

		fmt.Println("====================================================================== " + cfg.AppVersion)

		fmt.Println()
	}
}
