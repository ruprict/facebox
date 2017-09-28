package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/net/websocket"

	"github.com/gorilla/mux"
	"github.com/machinebox/sdk-go/facebox"
)

const boundary = "informs"

var (
	fbox   *facebox.Client
	wsChan chan string
)

func main() {
	wsChan = make(chan string)
	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("."))))
	r.HandleFunc("/cam", cam)
	r.HandleFunc("/camFacebox", camFacebox)
	r.HandleFunc("/sound/{name}", soundHandler)
	r.Handle("/socket", websocket.Handler(socket))

	fbox = facebox.New("http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8081", r))
}

func soundHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dir, err := os.Getwd()

	path := filepath.Join(dir, "/sound/", strings.ToLower(vars["name"])+".mp3")
	_, err = os.Stat(path)
	if err != nil {
		fmt.Println("*** file error ", err)
	}
	/*stat, err := os.Stat(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if stat.IsDir() {
		serveDir(w, r, path)
		return
	}*/
	http.ServeFile(w, r, path)
}
func playSound(name string, w http.ResponseWriter, r *http.Request) {
	/*path := filepath.Join(".", r.URL.Path[len("/sound/"):])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if stat.IsDir() {
		serveDir(w, r, path)
		return
	}*/
	path := "./sound/" + name + ".mp3"
	http.ServeFile(w, r, path)
}

func cam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+boundary)
	cmd := exec.CommandContext(r.Context(), "./capture.py")
	cmd.Stdout = w
	err := cmd.Run()
	if err != nil {
		log.Println("[ERROR] capturing webcam", err)
	}
}

func camFacebox(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+boundary)
	cmd := exec.CommandContext(r.Context(), "./capture.py")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("[ERROR] Getting the stdout pipe")
		return
	}
	cmd.Start()

	mr := multipart.NewReader(stdout, boundary)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			log.Println("[DEBUG] EOF")
			break
		}
		if err != nil {
			log.Println("[ERROR] reading next part", err)
			return
		}
		jp, err := ioutil.ReadAll(p)
		if err != nil {
			log.Println("[ERROR] reading from bytes ", err)
			continue
		}
		jpReader := bytes.NewReader(jp)
		faces, err := fbox.Check(jpReader)
		if err != nil {
			log.Println("[ERROR] calling facebox", err)
			continue
		}
		for _, face := range faces {
			if face.Matched {
				fmt.Println("I know you ", face.Name)
				wsChan <- face.Name
			} else {
				wsChan <- "unknown"
				fmt.Println("I DO NOT know you ")
			}
		}

		// just MJPEG
		w.Write([]byte("Content-Type: image/jpeg\r\n"))
		w.Write([]byte("Content-Length: " + string(len(jp)) + "\r\n\r\n"))
		w.Write(jp)
		w.Write([]byte("\r\n"))
		w.Write([]byte("--informs\r\n"))
	}
	cmd.Wait()
}

type message struct {
	// the json tag means this will serialize as a lowercased field
	Message string `json:"message"`
}

func socket(ws *websocket.Conn) {
	// let's do websocket stuff!
	for msg := range wsChan {
		fmt.Println("msg comming ", msg)
		m2 := message{msg}
		if err := websocket.JSON.Send(ws, m2); err != nil {
			log.Println(err)
			break
		}

	}
}
