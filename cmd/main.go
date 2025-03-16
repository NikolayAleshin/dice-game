package main

import "dice-game/cmd/app"

func main() {
	application := app.NewApplication()
	application.Run()
}
