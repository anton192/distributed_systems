package main

import (
	"fmt"
	"strconv"
	// "time"
	"sync"
    "net"
    // "os"
	"./graph"
    "encoding/json"
)

type Message struct {
    Id int
    Type string
    Sender int
    Origin int
    Data string
}

var DEBUG_LEVEL = 0
var COUNT_VERTEX = 128
var START_PORT = 10000
var current_message_id = 0

var G graph.Graph
var wg sync.WaitGroup
var wg2 sync.WaitGroup

/* A Simple function to verify error */
func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        // os.Exit(0)
    }
}

func send_message(port_from int, port_to int, msg Message) {

    if (DEBUG_LEVEL >= 2) {
        fmt.Println("Node ", port_from, " send msg to node ", port_to)
    }
    
    LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    CheckError(err)

    ServerAddr,err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(port_to))
    CheckError(err)
 
    Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
    CheckError(err)
    defer Conn.Close()

    text_msg, _ := json.Marshal(msg)
    Conn.Write(text_msg)
}

func node_process_server(node graph.Node) {

    if (DEBUG_LEVEL >= 1) {
	   fmt.Println("node_process_server", node, node.Port(), G[node])
    }

	ServerAddr,err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(node.Port()))
    CheckError(err)

    ServerConn, err := net.ListenUDP("udp", ServerAddr)
    CheckError(err)
    defer ServerConn.Close()

	buf := make([]byte, 1024)
    recieved_messages := make([]bool, COUNT_VERTEX * COUNT_VERTEX)
    var recieved_nodes []int
    wg.Add(-1)
    wg.Wait()
    //  В этом месте все процессы проинициализировались


    if node.String() == "0" {
        wg2.Add(1)
        msg := Message {
            Id: current_message_id,
            Type: "multicast",
            Sender: node.Port(),
            Origin: node.Port(),
            Data: "data"}
        current_message_id += 1

        for vertex_index := range G[node] {
            send_message(node.Port(), G[node][vertex_index].Port(), msg)
        }
    }

    var msg Message
    var already_have bool
    for {
        n,addr,err := ServerConn.ReadFromUDP(buf)
        if (DEBUG_LEVEL >= 2) {
            fmt.Println("[", node.Port(), "] Received ", string(buf[0:n]), " from ", addr)
        }
        if err != nil {
            fmt.Println("Error: ",err)
        }

        var recieved_message Message
        err = json.Unmarshal(buf[0:n], &recieved_message)
        if err != nil {
            fmt.Println("Error:", err)
        }

        
        if (recieved_messages[recieved_message.Id] == false) {
            // Мы впервые получили сообщение
            recieved_messages[recieved_message.Id] = true

            if (recieved_message.Type == "multicast") {
                // Начинаем отсылать сообщение о том, что мы получили сообщение
                msg = Message {
                    Id: current_message_id,
                    Type: "notification",
                    Sender: node.Port(),
                    Origin: node.Port(),
                    Data: ""}
                current_message_id += 1
                for vertex_index := range G[node] {
                    send_message(node.Port(), G[node][vertex_index].Port(), msg)
                }
            }

            // Перешлем его всем
            msg = Message {
                Id: recieved_message.Id,
                Type: recieved_message.Type,
                Sender: node.Port(),
                Origin: recieved_message.Origin,
                Data: recieved_message.Data}
            for vertex_index := range G[node] {
                send_message(node.Port(), G[node][vertex_index].Port(), msg)
            }
        }

        if (node.String() == "0" && recieved_message.Type == "notification") {
            already_have = false
            for _, el := range recieved_nodes {
                if (el == recieved_message.Origin) {
                    already_have = true
                }
            }
            if (!already_have) {
                recieved_nodes = append(recieved_nodes, recieved_message.Origin)
                if (len(recieved_nodes) == len(G)) {
                    wg2.Add(-1)
                }
            }
        }
  
    }    

}

func main() {

	// generating graph
	G = graph.Generate(COUNT_VERTEX, 2, 3, START_PORT);

	// Run server-side
	for vertex := range G {
        wg.Add(1)
		go node_process_server(vertex)
	}

	
    wg.Wait()
    wg2.Wait()
    // time.Sleep(time.Second * 5)
	
}
