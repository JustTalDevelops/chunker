package main

import (
	"fmt"
	"github.com/JustTalDevelops/chunker"
	"io/ioutil"
)

func main() {
	w, err := chunker.NewWorld("world.mcworld")
	if err != nil {
		panic(err)
	}
	err = w.Connect(ready)
	if err != nil {
		panic(err)
	}
}

func ready(w *chunker.World) {
	fmt.Println("Sending preview request...")
	err := w.WriteRequest(chunker.NewPreviewRequest())
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			if w.PreviewLoaded() {
				break
			}
		}

		p, err := w.Preview(0, 0)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile("result.png", p, 0777)
		if err != nil {
			panic(err)
		}

		fmt.Println("Saved preview!")
	}()
}