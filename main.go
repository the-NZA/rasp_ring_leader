package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"sort"
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
	flag.StringVar(&ID, "id", "default", "Self ID")
	flag.StringVar(&NextAddr, "n", "localhost", "Next node address")
	flag.Parse()

	log.Println("ID:", ID)
	log.Println("Next node address:", NextAddr)

	if ID == "default" {
		log.Fatal("Next node address is not set")
	}

	if NextAddr == "" || NextAddr == "localhost" {
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

	go handleEnterPress()

	// accept connections
	for {
		conn, err := srv.Accept()
		if err != nil {
			log.Println("error:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleEnterPress() {
	fmt.Println("Press Enter to start choosing Master node")
	_, err := fmt.Scanln()
	if err != nil {
		log.Fatal(err)
	}

	// create initial DTO
	dto := &DTO{
		Command: "who",
		IDs:     []string{ID},
	}

	// send Who is master? request
	if err := sendNext(dto); err != nil {
		log.Fatal(err)
	}
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
		// sort IDs
		sort.Slice(dto.IDs, func(i, j int) bool {
			return dto.IDs[i] < dto.IDs[j]
		})

		// Find self ID
		found := false
		for _, id := range dto.IDs {
			if id == ID {
				found = true
				break
			}
		}

		// send self ID to next node if not found
		if !found {
			dto.IDs = append(dto.IDs, ID)
			if err := sendNext(dto); err != nil {
				return err
			}

			return nil
		}

		// Find max ID
		maxID := dto.IDs[len(dto.IDs)-1]

		// if max ID is self ID, then I am the leader
		if maxID == ID {
			LeaderID = ID
			dto.Command = "leader"

			if err := sendNext(dto); err != nil {
				return err
			}
		}

	case "leader":
		// sort IDs
		sort.Slice(dto.IDs, func(i, j int) bool {
			return dto.IDs[i] < dto.IDs[j]
		})

		// Find max ID
		maxID := dto.IDs[len(dto.IDs)-1]

		// if max ID is self ID, then I am the leader
		if maxID == ID {
			LeaderID = ID
			log.Println("Leader is found:", LeaderID)

			if err := sendNext(dto); err != nil {
				return err
			}

			return nil
		}

		// find self ID index
		var selfIDIndex int
		for i, id := range dto.IDs {
			if id == ID {
				selfIDIndex = i
				break
			}
		}

		// remove self ID from IDs
		dto.IDs = append(dto.IDs[:selfIDIndex], dto.IDs[selfIDIndex+1:]...)

		// send Leader ID to next node
		if err := sendNext(dto); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown command: %s", dto.Command)
	}

	return nil
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

func parseData(data []byte) (*DTO, error) {
	var dto DTO
	err := json.Unmarshal(data, &dto)
	if err != nil {
		return nil, err
	}

	return &dto, nil
}
