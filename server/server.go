package server

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"text/template"

	"golang.org/x/net/websocket"
)

//go:embed tpl/*.gohtml
var templateFiles embed.FS

var allServices []string
var selectedServices []string

// A socket server instance is set up to handle
// multiple (concurrent) games.
type SocketServer struct {
	Location URL
	//	Games    map[string]*Game
	Tpl *template.Template
	Mux *http.ServeMux
}

func NewSocketServer(url URL) *SocketServer {
	return &SocketServer{
		Location: url,
		//		Games:    make(map[string]*Game),
		// In `tpl/`, the `_base.html` file **must** be the first file!!
		// The underscore (_) is lexically before any lowercase alpha character,
		// **do not** remove it!!!  Everything will break!!!
		Tpl: template.Must(template.ParseFS(templateFiles, "tpl/*.gohtml")),
		Mux: http.NewServeMux(),
	}
}

type Socket struct {
	Protocol string
	Domain   string
	Port     int
}

func (s Socket) String() string {
	return fmt.Sprintf("%s://%s:%d",
		s.Protocol,
		s.Domain,
		s.Port,
	)
}

type URL struct {
	Sock Socket
	Path string
}

func (u URL) String() string {
	return fmt.Sprintf("%s://%s:%d/%s",
		u.Sock.Protocol,
		u.Sock.Domain,
		u.Sock.Port,
		u.Path,
	)
}

// This is marshaled to the browser client.
// See [SocketServer.Publish].
type ServerMessage struct {
	Type string `json:"type,omitempty"`
	Data any    `json:"data,omitempty"`
}

// The socket server unmarshals the response from the
// browser client into this type.
type ClientMessage struct {
	Type string `json:"type,omitempty"`
	Data any    `json:"data,omitempty"`
}

func (s *SocketServer) Message(socket *websocket.Conn, msg ServerMessage) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = socket.Write(b)
	if err != nil {
		return fmt.Errorf("websocket write error: %v", err)
	}
	return nil
}

func DirtyWorktree(v any) error {
	var out []interface{}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			out = append(out, rv.Index(i).Interface())
		}
	}
	for _, v := range out {
		entry := strings.Split(v.(string), ",")
		cmd := exec.Command(
			"sed",
			"-i",
			fmt.Sprintf("s/\\(.*newTag:\\s\\).*/\\1\"%s\"/", entry[1]),
			fmt.Sprintf("gitops/applications/devops/%s/overlays/production/kustomization.yaml", entry[0]))
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func GetAllServices(filename string) ([]string, error) {
	f, err := os.Open(fmt.Sprintf("production-deployments/%s", filename))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// Since we're (potentially) re-using this slice, zero it out.
	allServices = allServices[:0]
	fileScanner := bufio.NewScanner(f)
	for fileScanner.Scan() {
		allServices = append(allServices, fileScanner.Text())
	}
	return allServices, err
}

func GetDeploymentDates() ([]string, error) {
	dates := []string{}
	err := filepath.WalkDir("production-deployments", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if !d.IsDir() && !strings.Contains(path, ".git") {
			//			fmt.Printf("dir: %v: path: %s name: %s\n", d.IsDir(), path, d.Name())
			dates = append(dates, d.Name())

		}
		return nil
	})
	return dates, err
}

func GetSelectedServices(v any) []string {
	var out []interface{}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			out = append(out, rv.Index(i).Interface())
		}
	}
	for _, v := range out {
		svc := allServices[int(v.(float64))]
		if !slices.Contains(selectedServices, svc) {
			selectedServices = append(selectedServices, svc)
		}
	}
	return selectedServices
}

func remove[T comparable](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}

func AddToSelectedServices(v string) []string {
	if !slices.Contains(selectedServices, v) {
		selectedServices = append(selectedServices, v)
	}
	return selectedServices
}

func RemoveFromSelectedServices(v string) {
	i := slices.Index(selectedServices, v)
	selectedServices = remove(selectedServices, i)
}

// Registers all the handlers with the new mux, adds the middleware
// and starts starts the game server.
func (s *SocketServer) Start() {
	s.Mux.Handle("/ws", websocket.Handler(s.DefaultHandler))
	s.Mux.HandleFunc("/", s.BaseHandler)
	log.Fatal(http.ListenAndServe(":3000", s.Mux))
}
