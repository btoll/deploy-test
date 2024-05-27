package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/btoll/deploy-test/git"
	"github.com/btoll/deploy-test/server"
)

var (
	wsURL   = flag.String("ws", "ws://127.0.0.1:3000", "URL of game websocket server")
	hostURL = flag.String("host", "https://127.0.0.1:3000", "URL of game host server")
)

func parseURL(s string) server.Socket {
	parsedUrl, err := url.Parse(s)
	if err != nil {
		log.Fatalln("server url could not be parsed")
	}
	port, err := strconv.Atoi(parsedUrl.Port())
	if err != nil {
		log.Fatalln("port could not be parsed")
	}
	return server.Socket{
		Protocol: parsedUrl.Scheme,
		Domain:   parsedUrl.Hostname(),
		Port:     port,
	}
}

func bound(n int) string {
	return strings.Repeat("-", n)
}

func main() {
	flag.Parse()

	var err error
	git.ProductionDeployments, err = git.Clone(&git.Cloner{
		URL:        "git@bitbucket.org:pecteam",
		Repository: "production-deployments",
		Branch:     "master",
		CloneDir:   "production-deployments",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	git.GitOps, err = git.Clone(&git.Cloner{
		URL:        "git@bitbucket.org:pecteam",
		Repository: "owls-nest-farm",
		Branch:     "master",
		CloneDir:   "gitops",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	wsSock := parseURL(*wsURL)
	//	hostSock := parseURL(*hostURL)
	socketServer := server.URL{
		Sock: wsSock,
		Path: "ws",
	}

	sockserv := server.NewSocketServer(socketServer)
	fmt.Printf("%s\ncreated new websocket server `%s`\n",
		bound(75),
		socketServer)
	sockserv.Start()
}
