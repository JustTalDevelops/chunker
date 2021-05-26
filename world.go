package chunker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// Handler handles packet data that wasn't automatically handled here.
type Handler interface {
	// Handle ...
	Handle(data []byte) bool
}

// Type is a type of Minecraft world.
type Type struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Version contains version specific info for Chunker worlds.
type Version struct {
	Input   Type   `json:"input"`
	Writers []Type `json:"writers"`
}

// World is a Chunker world upload.
type World struct {
	Success    bool    `json:"success"`
	Session    string  `json:"session"`
	Version    Version `json:"version"`
	settings   Settings
	connection net.Conn
	loggedIn   bool
	preview    bool
}

// Settings is a list of settings that can configure some parts of this library.
type Settings struct {
	// Handler ...
	Handler Handler
	// Log ...
	Log *logrus.Logger
}

// PreviewLoaded returns a bool if the preview is loaded.
func (w *World) PreviewLoaded() bool {
	return w.preview
}

// Preview returns a preview of a chunk as a PNG in bytes.
func (w *World) Preview(x, z int) (b []byte, err error) {
	if w.preview && w.connection != nil {
		resp, err := http.Get("https://chunker.app/api/preview/" + w.Session + "/OVERWORLD/" + strconv.Itoa(x) + "/" + strconv.Itoa(z))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return b, nil
	}

	return nil, errors.New("the connection is no longer active or the preview is not ready")
}

// WriteRequest writes a request to Chunker.
func (w *World) WriteRequest(v interface{}) error {
	if w.connection == nil {
		return errors.New("connection is not open but request was attempted")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	err = wsutil.WriteClientMessage(w.connection, ws.OpText, b)
	if err != nil {
		return err
	}

	return nil
}

// LoggedIn ...
func (w *World) LoggedIn() bool {
	return w.loggedIn
}

// Disconnect ...
func (w *World) Disconnect() {
	w.loggedIn = false
	w.connection = nil
}

// Connect connects to Chunker's websocket.
func (w *World) Connect(readyFunc func()) (err error) {
	w.connection, _, _, err = ws.DefaultDialer.Dial(context.Background(), "wss://chunker.app/")
	if err != nil {
		return
	}

	err = w.WriteRequest(NewLoginRequest(w.Session))
	if err != nil {
		return
	}

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		var err error
		var pinging bool
		for {
			if w.connection == nil {
				break
			}
			if pinging {
				continue
			}
			pinging = true
			go func() {
				time.Sleep(30 * time.Second)
				err = w.WriteRequest(NewPingRequest())
				if err != nil {
					panic(err)
				}
				pinging = false
			}()
		}
		wg.Done()
	}()
	go func() {
		var err error
		var msg []byte
		var reading bool
		for {
			if w.connection == nil {
				break
			}
			if reading {
				continue
			}
			reading = true
			go func() {
				msg, _, err = wsutil.ReadServerData(w.connection)
				if err != nil {
					panic(err)
				}
				reading = false
				switch gjson.GetBytes(msg, "type").String() {
				case "pong":
					// We can ignore this.
				case "login_success":
					w.settings.Log.Println("Logged in to Chunker!")
					w.loggedIn = true
					readyFunc()
				case "preview":
					w.settings.Log.Debugln("Generated a preview!")
					w.preview = true
				default:
					if w.settings.Handler != nil {
						if w.settings.Handler.Handle(msg) {
							return
						}
					}

					w.settings.Log.Debugln("Unhandled packet data: " + string(msg))
				}
			}()
		}
		wg.Done()
	}()

	wg.Wait()
	w.settings.Log.Println("Disconnected from Chunker.")

	return nil
}

// NewWorld creates a new Chunker world from a file.
func NewWorld(file string, settings ...Settings) (*World, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	w, err := NewWorldFromReader(f, settings...)
	if err != nil {
		return nil, err
	}
	return w, nil
}

// NewWorldFromReader creates a new Chunker world from a reader.
func NewWorldFromReader(reader io.Reader, settings ...Settings) (resultWorld *World, err error) {
	if len(settings) == 0 {
		settings = append(settings, Settings{})
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	var fw io.Writer

	if fw, err = w.CreateFormFile("file", "world.mcworld"); err != nil {
		return nil, err
	}
	if _, err = io.Copy(fw, reader); err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://chunker.app/api/input/uploadWorld", &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", res.Status)
	}

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var world World
	err = json.Unmarshal(result, &world)
	if err != nil {
		return nil, err
	}
	world.settings = settings[0]

	return &world, nil
}
