package main

import (
	"flag"
	"fmt"
	"sync"
    "time"
	"net"
	"strconv"
    "encoding/json"
)

type SysMsg struct {
    Type string
    Dst int
    Data string
}
type Msg struct {
    Type string
    Src int
    Dst int
    Data string
}

var N = 4
var T = 10
var PORT_MESSAGE = 30000
var PORT_COMMAND = 40000
var queue [][]Msg
var drop []int
var curHop = 0
var lastUpdate []int

var wgInit sync.WaitGroup
var wgEnd  sync.WaitGroup

func parse_args() {
	var param_N = flag.Int("n", N, "Count of process")
	var param_T = flag.Int("t", T, "Delay time")
	flag.Parse()
    N = *param_N
    T = *param_T
    // println(N, T)
}

func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        // os.Exit(0)
    }
}

func send_message(port_to int, msg []byte) {
    
    LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    CheckError(err)

    ServerAddr,err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(port_to))
    CheckError(err)
 
    Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
    CheckError(err)
    defer Conn.Close()

    Conn.Write(msg)
}




func service_process(number int) {
	ServerAddr,err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(PORT_COMMAND + number))
    CheckError(err)

    ServerConn, err := net.ListenUDP("udp", ServerAddr)
    CheckError(err)
    defer ServerConn.Close()

    buf := make([]byte, 1024)
    wgInit.Add(-1)
    wgInit.Wait()

    var msg Msg

	for {
        n, addr, err := ServerConn.ReadFromUDP(buf)
        fmt.Println("node", number, ": recieved service message:", string(buf[0:n]))
        if err != nil {
            fmt.Println("Error: ", err, addr)
        }

        var recieved_message SysMsg
        err = json.Unmarshal(buf[0:n], &recieved_message)
        if err != nil {
            fmt.Println("Error:", err)
        }

        if (recieved_message.Type == "send") {
            msg = Msg {
                Type: "send",
                Src: number,
                Dst: recieved_message.Dst,
                Data: recieved_message.Data}
            queue[number] = append(queue[number], msg)
        } else if (recieved_message.Type == "drop") {
            drop[number]++
        }
    }

}

func checkMessageDelivery(beginHop int, number int, msg Msg) {
    time.Sleep(time.Duration((N + 1) * T) * time.Millisecond)
    if (lastUpdate[number] == beginHop) {
        fmt.Println("Node", number, ": error with message", msg)
        newQueue := []Msg{msg}
        for elementIndex := range queue[number] {
            newQueue = append(newQueue, queue[number][elementIndex])
        }
        queue[number] = newQueue
        send_new_message(number)
    }
}

func checkEmpty(beginHop int, number int, msg []byte) {
    time.Sleep(time.Duration((N + 1) * T) * time.Millisecond)
    if (lastUpdate[number] == beginHop) {
        fmt.Println("Node", number, ": error with empty message")
        send_message(PORT_MESSAGE + number, msg)
    }
}

func send_new_message(number int) {
    msg := queue[number][0]
    text, _ := json.Marshal(msg)
    queue[number] = queue[number][1:]
    currentHop := lastUpdate[number]
    send_message(PORT_MESSAGE + (number + 1) % N, text)
    go checkMessageDelivery(currentHop, number, msg)
}

func local_process(number int) {
	ServerAddr,err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(PORT_MESSAGE + number))
    CheckError(err)

    ServerConn, err := net.ListenUDP("udp", ServerAddr)
    CheckError(err)
    defer ServerConn.Close()

	buf := make([]byte, 1024)
	wgInit.Add(-1)
    wgInit.Wait()

    empty_token := Msg {
        Type: "empty",
        Src: 0,
        Dst: 0,
        Data: ""}
    empty_token_text, _ := json.Marshal(empty_token)

    for {

        n, addr, err := ServerConn.ReadFromUDP(buf)
        // fmt.Println("node", number, ": recieved", string(buf[0:n]))

        if err != nil {
            fmt.Println("Error: ", err, addr)
        }

        var recieved_message Msg
        err = json.Unmarshal(buf[0:n], &recieved_message)
        if err != nil {
            fmt.Println("Error:", err)
        }
        
        time.Sleep(time.Duration(T) * time.Millisecond)
        if (drop[number] != 0) {
            drop[number]--
            continue
        }
        lastUpdate[number] = curHop
        if (number == 0 && recieved_message.Type == "empty") {
            go checkEmpty(curHop, 0, empty_token_text)
        }
        curHop++


        next := (number + 1) % N

        if (recieved_message.Type == "send") {
            if (recieved_message.Dst == number) {
                fmt.Println("node", number, ": recieved token from node", addr, 
                    "with data from node", recieved_message.Src, "(",
                    recieved_message.Data, ")")
                recieved_message.Type = "confirmation"
                recieved_message.Dst = recieved_message.Src
                recieved_message.Src = number
                conf_msg_text, _ := json.Marshal(recieved_message)
                send_message(PORT_MESSAGE + next, conf_msg_text)
            } else {
                fmt.Println("node", number, ": recieved token from node", addr, 
                    "with data from node", recieved_message.Src, "(",
                    recieved_message.Data, "), sending token to node", next)
                send_message(PORT_MESSAGE + next, buf[0:n])
            }
        } else if (recieved_message.Type == "empty") {
            fmt.Println("node", number, ": recieved token from node", addr, 
                ", sending token to node", next)
            if (len(queue[number]) == 0) {
                send_message(PORT_MESSAGE + next, empty_token_text)
            } else {
                send_new_message(number)
            }
        } else if (recieved_message.Type == "confirmation") {
            if (recieved_message.Dst == number) {
                fmt.Println("node", number, ": recieved token from node", addr, 
                    "with delivery confirmation from node", recieved_message.Src)
                if (len(queue[number]) == 0) {
                    send_message(PORT_MESSAGE + next, empty_token_text)
                } else {
                    send_new_message(number)
                }
            } else {
                fmt.Println("node", number, ": recieved token from node", addr, 
                    "with delivery confirmation from node", recieved_message.Src,
                    ", sending token to node", next)
                send_message(PORT_MESSAGE + next, buf[0:n])
            }
        }
    }

    wgEnd.Done()
}

func main() {

    parse_args()
    queue = make([][]Msg, N)
    drop = make([]int, N)
    lastUpdate = make([]int, N)

    empty_token := Msg {
        Type: "empty",
        Src: 0,
        Dst: 0,
        Data: ""}
    empty_token_text, _ := json.Marshal(empty_token)

    wgInit.Add(2 * N)
    wgEnd.Add(N)
    for i := 0; i < N; i++ {
    	go local_process(i)
    	go service_process(i)
    }
    wgInit.Wait()
    send_message(PORT_MESSAGE, empty_token_text)



    msg := SysMsg {
        Type: "send",
        Dst: 7,
        Data: "data"}
    text_msg, _ := json.Marshal(msg)
    send_message(PORT_COMMAND + 3, text_msg)

    msg = SysMsg {
        Type: "drop",
        Dst: 0,
        Data: ""}
    text_msg, _ = json.Marshal(msg)
    send_message(PORT_COMMAND + 2, text_msg)


    wgEnd.Wait()
    // time.Sleep(time.Second * 5)

}       