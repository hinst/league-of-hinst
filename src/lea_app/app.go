package main

import (
	"fmt"
	"lea"
)

func main() {
	fmt.Println("STARTING...")
	var app = (&lea.TApp{}).Create()
	app.Run()
}
