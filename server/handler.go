package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/btoll/deploy-test/git"
	"golang.org/x/net/websocket"
)

func (s *SocketServer) BaseHandler(w http.ResponseWriter, r *http.Request) {
	r.Header = http.Header{
		"Content-Type": {"text/html; charset=utf-8"},
	}

	if err := s.Tpl.Execute(w, s.Location); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *SocketServer) DefaultHandler(socket *websocket.Conn) {
	buf := make([]byte, 1024)
	origin := socket.Config().Origin
	location := socket.Config().Location

	fmt.Println("incoming connection from client", location)

	// Get the deployment dates from the name of the files in the
	// `production-deployments` repository.
	// Abort if this fails because there's no reason to go on.
	dates, err := GetDeploymentDates()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	s.Message(socket, ServerMessage{
		Type: "production-dates",
		Data: dates,
	})

	for {
		n, err := socket.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println(origin)
				// This means the client connection has closed.
				if err != nil {
					fmt.Println("read error:", err)
				} else {
					//					err := s.Message(socket, ServerMessage{
					//						Type: "dates",
					//						Data: "foo",
					//					})
					//					if err != nil {
					//						fmt.Fprintln(os.Stderr, err)
					//						os.Exit(1)
					//					}
				}
				allServices = allServices[:0]
				selectedServices = selectedServices[:0]
				break
			}
			// Don't `return` here, it will break the connection.
			continue
		}

		data := buf[:n]

		var msg ClientMessage
		err = json.Unmarshal(data, &msg)
		if err != nil {
			log.Fatalln(err)
		}

		switch msg.Type {
		case "add-to-selected-services":
			err = s.Message(socket, ServerMessage{
				Type: "selected-services",
				Data: AddToSelectedServices(msg.Data.(string)),
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

		case "create-pr":
			unixTimestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
			_, err := git.Branch(git.GitOps, unixTimestamp)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			err = git.Checkout(git.GitOps, unixTimestamp)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			err = DirtyWorktree(msg.Data)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			err = git.Commit(git.GitOps)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			err = git.Push(git.GitOps)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

		case "deployment-date":
			services, err := GetAllServices(msg.Data.(string))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			err = s.Message(socket, ServerMessage{
				Type: "all-services",
				Data: services,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

		case "remove-from-selected-services":
			RemoveFromSelectedServices(msg.Data.(string))

		case "selected-services":
			err = s.Message(socket, ServerMessage{
				Type: "selected-services",
				Data: GetSelectedServices(msg.Data),
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}
}
