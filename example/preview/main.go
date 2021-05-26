package main

import (
	"github.com/JustTalDevelops/chunker"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func main() {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true}
	log.Level = logrus.DebugLevel

	w, err := chunker.NewWorld("world.mcworld", chunker.Settings{
		Log: log,
	})
	if err != nil {
		panic(err)
	}
	err = w.Connect(func() {
		log.Println("Sending preview request...")
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

			log.Println("Saved preview!")

			w.Disconnect()
		}()
	})
	if err != nil {
		panic(err)
	}
}
