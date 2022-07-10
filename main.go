package main

import (
	"chip-8-go/internal/chip8"
	"github.com/faiface/pixel/pixelgl"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	pixelgl.Run(run)
}

func run() {
	var rom string
	var ck int

	app := &cli.App{
		Name:  "chip8",
		Usage: "chip-8 emulator.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "rom",
				Value:       "",
				Usage:       "path to the rom",
				Destination: &rom,
				Required:    true,
			},
			&cli.IntFlag{
				Name:        "ck",
				Value:       60,
				Usage:       "clock speed",
				Destination: &ck,
			},
		},
		Action: func(ctx *cli.Context) error {
			vm, err := chip8.NewVm(rom, ck)
			if err != nil {
				return err
			}

			go vm.Run()

			<-vm.ShutdownCh
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("oops, something went wrong: %v\n", err)
	}
}
