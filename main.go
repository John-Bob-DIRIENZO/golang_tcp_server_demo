package main

import (
    "errors"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "os/signal"
    "sync"
    "syscall"
)

type Message struct {
    Message []byte
    Sender  string
}

type Server struct {
    ListenAddr string
    ln         net.Listener
    quit       chan bool
    msg        chan Message
    peers      map[string]net.Conn
    mutex      sync.Mutex
}

func NewServer(listenAddr string) *Server {
    return &Server{
        ListenAddr: listenAddr,
        quit:       make(chan bool),
        msg:        make(chan Message),
        peers:      make(map[string]net.Conn),
        mutex:      sync.Mutex{},
    }
}

func (s *Server) Start() error {
    ln, err := net.Listen("tcp", s.ListenAddr)
    if err != nil {
        return err
    }

    s.ln = ln

    go s.acceptConnection()

    <-s.quit
    err = ln.Close()
    if err != nil {
        return err
    }
    close(s.msg)
    return nil
}

// acceptConnection est une boucle qui va simplement prendre toutes les connexions,
// les ajouter à la peerList et les passer à la boucle de gestion "metier" de la connexion
func (s *Server) acceptConnection() {
    for {
        conn, err := s.ln.Accept()
        if err != nil {
            if errors.Is(err, net.ErrClosed) {
                log.Println("Seems like the connection died...")
                break
            }

            log.Println("Accept error: ", err)
            continue
        }

        fmt.Println("New connection from ", conn.RemoteAddr())
        s.addPeer(conn)
        s.Broadcast(Message{
            Message: []byte(fmt.Sprintf("%s vient d'arriver \n", conn.RemoteAddr().String())),
            Sender:  conn.RemoteAddr().String(),
        })
        go s.handleConnection(conn)
    }
}

// handleConnection est une boucle qui prend tous les messages passés via la connexion pour en faire quelque chose.
// Ici, on ne fait que les broadcast aux autres utilisateurs
func (s *Server) handleConnection(conn net.Conn) {
    defer conn.Close()
    buffer := make([]byte, 2048)

    for {
        n, err := conn.Read(buffer)

        if err != nil {
            if err != io.EOF {
                log.Println("Read error: ", err)
                continue
            }

            s.Broadcast(Message{
                Message: []byte(fmt.Sprintf("%s s'est déconnecté \n", conn.RemoteAddr().String())),
                Sender:  conn.RemoteAddr().String(),
            })
            s.delPeer(conn)
            fmt.Println("Connection closed by ", conn.RemoteAddr())
            break
        }

        // On fait une copie du slice pour éviter de passer une référence via la chanel
        data := make([]byte, n)
        copy(data, buffer[:n])

        s.msg <- Message{data, conn.RemoteAddr().String()}
        _, _ = conn.Write([]byte("Merci mec, c'est cool\n"))
    }
}

// addPeer utilise un mutex pour éviter une condition de course sur l'accès à la liste des peers
func (s *Server) addPeer(conn net.Conn) {
    s.mutex.Lock()
    s.peers[conn.RemoteAddr().String()] = conn
    s.mutex.Unlock()
}

func (s *Server) delPeer(conn net.Conn) {
    s.mutex.Lock()
    delete(s.peers, conn.RemoteAddr().String())
    s.mutex.Unlock()
}

// Broadcast va simplement passer le message à tous les autres utilisateurs
func (s *Server) Broadcast(message Message) {
    for _, peer := range s.peers {
        if peer.RemoteAddr().String() != message.Sender {
            _, _ = peer.Write(message.Message)
        }
    }
}

func main() {
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    s := NewServer(":1287")

    go func() {
        <-sigs
        s.quit <- true
    }()

    go func() {
        for message := range s.msg {
            s.Broadcast(message)
        }
    }()

    err := s.Start()
    if err != nil {
        log.Fatal(err)
    }
}
