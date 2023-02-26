package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

/*
connect to the next in ring
press Space to start choosing Master node
send Who is master?
with self ID

next sends self ID
*/

var (
	ID   = "rkkozlov"
	Port = "8080"

	LeaderID string
	NextAddr string
)

type DTO struct {
	Command string   `json:"command"`
	IDs     []string `json:"ids"`
}

func main() {
	// parse flags
	flag.StringVar(&ID, "id", "rkkozlov", "Self ID")
	flag.StringVar(&NextAddr, "next", "", "Next node address")
	flag.Parse()

	if NextAddr == "" {
		log.Fatal("Next node address is not set")
	}

	// create TCP server
	srv, err := net.Listen("tcp", fmt.Sprintf(":%s", Port))
	if err != nil {
		log.Fatal(err)
	}
	defer func(srv net.Listener) {
		_ = srv.Close()
	}(srv)

	log.Println("Server started...")

	go handleSpacePress()

	// accept connections
	for {
		conn, err := srv.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleSpacePress() {
	// wait for space press
	// send Who is master?
	// send self ID

	// wait for response
	// if response is self ID - set self as leader
	// if response is not self ID - send Who is master? to next node

	// ---

	// read input from stdin until Space pressed (or Ctrl+C)
	fmt.Println("Press Space to start choosing Master node")

	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		if s.Text() != " " {
			continue
		}

		// create DTO
		dto := &DTO{
			Command: "who",
			IDs:     []string{ID},
		}

		// send Who is master? request
		if err := sendNext(dto); err != nil {
			log.Println(err)
			continue
		}

		return // Exit after first Space press
	}
}

func sendNext(dto *DTO) error {
	// connect to next node
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", NextAddr, Port))
	if err != nil {
		return err
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	// send data to next node
	data, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	// read data from connection until \n
	r := bufio.NewReader(conn)
	data, err := r.ReadBytes('\n')
	if err != nil {
		log.Println(err)
		return
	}

	// parse data
	dto, err := parseData(data)
	if err != nil {
		log.Println(err)
		return
	}

	// handle command
	if err := processDTO(dto); err != nil {
		log.Println(err)
		return
	}
}

func processDTO(dto *DTO) error {
	switch dto.Command {
	case "who":
		// send self ID

	case "leader":

	default:
		return fmt.Errorf("unknown command: %s", dto.Command)
	}

	return nil
}

func parseData(data []byte) (*DTO, error) {
	var dto DTO
	err := json.Unmarshal(data, &dto)
	if err != nil {
		return nil, err
	}

	return &dto, nil
}
