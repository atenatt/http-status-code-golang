package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Server struct {
	serverName    string
	serverURL     string
	tempoExecucao float64
	status        int
	dataFalha     string
}

func criarListaServidores(serverList *os.File) []Server {
	csvReader := csv.NewReader(serverList)
	data, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var servidores []Server
	for i, line := range data {
		if i > 0 {
			servidor := Server{
				serverName: line[0],
				serverURL:  line[1],
			}
			servidores = append(servidores, servidor)
		}
	}
	return servidores
}

func checkServer(servidores []Server) []Server {
	var downServer []Server
	for _, servidor := range servidores {
		agora := time.Now()
		get, err := http.Get(servidor.serverURL)
		if err != nil {
			fmt.Printf("Server %s is down [%s]\n", servidor.serverName, err.Error())
			servidor.status = 0
			servidor.dataFalha = agora.Format("02/01/2006 15:04:05")
			downServer = append(downServer, servidor)
			continue
		}
		servidor.status = get.StatusCode
		if servidor.status != 200 {
			servidor.dataFalha = agora.Format("02/01/2006 15:04:05")
			downServer = append(downServer, servidor)
		}
		servidor.tempoExecucao = time.Since(agora).Seconds()
		fmt.Printf("Status: [%d] Tempo de carga: [%f] URL: [%s]\n", servidor.status, servidor.tempoExecucao, servidor.serverURL)
	}
	return downServer
}

func openFiles(serverListFile string, downTimeFile string) (*os.File, *os.File) {
	serverList, err := os.OpenFile(serverListFile, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo de servidores")
	}
	downTimeList, err := os.OpenFile(downTimeFile, os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return serverList, downTimeList
}

func generateDownTime(downTimeList *os.File, downServers []Server) {
	csvWriter := csv.NewWriter(downTimeList)
	for _, servidor := range downServers {
		line := []string{servidor.serverName, servidor.serverURL, servidor.dataFalha, fmt.Sprintf("%f", servidor.tempoExecucao, fmt.Sprintf("%d", servidor.status))}
		csvWriter.Write(line)
	}
	csvWriter.Flush()
}

func main() {
	serverList, downTimeList := openFiles(os.Args[1], os.Args[2])
	defer serverList.Close()
	defer downTimeList.Close()
	servidores := criarListaServidores(serverList)
	downServer := checkServer(servidores)
	generateDownTime(downTimeList, downServer)
}
