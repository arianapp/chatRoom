package main

import (
    "os"
    "net"
    "bufio"
    "strconv"
    "fmt"
    s "strings"
)
var idAssignmentChan = make(chan string)
var clientTable = make(map[string]net.Conn)
var msgChan = make(chan Message)  // a channel to transit messages

type Message struct {
    sender string
    receiver string
    msgType int
    body string
    connInfo net.Conn
    flag int  // 1 transmit informaiton, 0 setup connection,
    // -1 close this connection
}

func HandleConnection(conn net.Conn) {
    b := bufio.NewReader(conn)

    client_id := <-idAssignmentChan
    msg := Message{client_id, "", 0, "", conn, 0}
    msgChan <- msg

    for {
        msg.msgType = 0
        line, err := b.ReadBytes('\n')
        if err != nil {
            conn.Close()
            msg.flag = -1
            msgChan <- msg  // direct push to the channel
            break
        }
      // creating the msg struct
      if s.Index(string(line), ":") != -1 {
        // get the command
         comm1 := s.SplitN(string(line), ":", 2)
         comm := s.TrimSpace(comm1[0])

         if s.Compare(comm, "all") == 0 {
            // broadcasted messages
            msg.msgType = 1
            msg.body = comm1[1]
          } else if s.Compare(comm, "whoami") == 0 {
            msg.msgType = 2
          } else if _, err := strconv.Atoi(comm); err == nil {
            msg.msgType = 3 // private message
            msg.receiver = string(comm)
            msg.body = comm1[1]
            // check the receiver later
          } else {
            // do nothing invalid command
          }
      //  conn.Write([]byte(client_id + ": " +string(line)))
      } else {
      // boradcast message
          msg.msgType = 1
          msg.body = string(line)
        }
        msg.flag = 1
        msgChan <- msg
    }
}

func IdManager() {
    var i uint64
    for i = 0;  ; i++ {
        idAssignmentChan <- strconv.FormatUint(i, 10)
    }
}

func HandleMsg()  {

  for {
     incomM := <- msgChan
    //  if this a new connection, handle the hashmap
     if incomM.flag == 0 {
          clientTable[incomM.sender] = incomM.connInfo
          // fmt.Println("set up connection successfully!")
     }
     if incomM.flag == -1 {
       // conn is closed , do nothing
       delete(clientTable, incomM.sender)
      //  fmt.Println("close one connection")

     } else {
       msgType := incomM.msgType
       if msgType == 1 {
          // broadcasted msg
          broadcast(incomM.sender, incomM.body)
       } else if msgType == 2 {
          // whoami request
          checkClientID(incomM.sender)
       } else if msgType == 3 {
          if _, ok := clientTable[incomM.receiver]; ok {
           privateMsg(incomM.sender, incomM.receiver, incomM.body)
          }
       } else {
         // do nothing
       }
     }
  }
}
func broadcast(senderID, m string) {
    for _, v := range clientTable {
        v.Write([]byte(senderID + ":" + m))
    }
}

func checkClientID(senderID string)  {
    clientTable[senderID].Write([]byte("chitter:" + senderID + "\n"))
}

func privateMsg(senderID, receiverID, m string)  {
    clientTable[receiverID].Write([]byte(senderID + ":" + m))
}

func main() {
    if len(os.Args) < 2{
        fmt.Fprintf(os.Stderr, "Usage: chitter <port-number>\n")
        os.Exit(1)
        return
    }
    port := os.Args[1]
    server, err := net.Listen("tcp", ":"+ port )
    if err != nil {
        fmt.Fprintln(os.Stderr, "Can't connect to port")
        os.Exit(1)
    }

    go IdManager()
    go HandleMsg()

    fmt.Println("Listening on port", os.Args[1])
    for{
        conn, _ := server.Accept()
        go HandleConnection(conn)
    }
}
