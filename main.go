package main

import (
    "log"
    "fmt"
    "os"
    "io/ioutil"
    "net"
    // "github.com/kr/pty"
    "golang.org/x/crypto/ssh"
    "golang.org/x/crypto/ssh/terminal"
)

var MAIN_LOG_FILE = os.Stdout
var CON_LOG_FILE = os.Stdout

func main() {
    mainLogger := log.New(MAIN_LOG_FILE, "main: ", 0)
    config := &ssh.ServerConfig{
        PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
            if c.User() == "user" && string(pass) == "password" {
                return nil, nil
            }
            return nil, fmt.Errorf("password not correct for %s", c.User())
        },
    }

    privateBytes, err := ioutil.ReadFile("id_rsa")
    if err != nil { mainLogger.Fatal("Failed to load private key") }
    mainLogger.Print("Loaded private key")

    private, err := ssh.ParsePrivateKey(privateBytes)
    if err != nil { mainLogger.Fatal("Failed to parse private key") }
    mainLogger.Print("Parsed private key")

    config.AddHostKey(private)

    for {
        listener, err := net.Listen("tcp", "127.0.0.1:2022")
        if err != nil {
            mainLogger.Print(fmt.Sprintf("Failed to accept incoming connection: %s", err))
            continue
        }
        mainLogger.Print("Listening for attempt on 127.0.0.1:2022")

        nConn, err := listener.Accept()
        if err != nil {
            mainLogger.Print("Failed to accept incoming connection")
            continue
        }
        mainLogger.Printf("Accepted connection from %s", nConn.RemoteAddr())

        serviceConnection(nConn, config)
        listener.Close()
    }
}

func serviceConnection(nConn net.Conn, config *ssh.ServerConfig) {
    connLogger := log.New(CON_LOG_FILE, "conn: ", 0)
    sshConn, chans, reqs, err := ssh.NewServerConn(nConn, config)
    if err != nil {
        connLogger.Printf("Failed handhake: %s", err)
        return
    }
    connLogger.Printf("Opened Connection for %s@%s (%s)", sshConn.User(), sshConn.RemoteAddr(), sshConn.ClientVersion())

    go func() {
        for req := range reqs {
            connLogger.Printf("Discarding request type (%s), with playload:\n%s", req.Type, req.Payload)
            if req.WantReply {
                req.Reply(false, nil)
            }
        }
    }()

    for newChannel := range chans {
        if t := newChannel.ChannelType(); t != "session" {
            newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
            connLogger.Printf("Unknown channel type: %s", t)
            continue
        }

        channel, requests, err := newChannel.Accept()
        if err != nil {
            connLogger.Printf("Could not accept channel: %s", err)
            return
        }

        go func(in <-chan *ssh.Request) {
            for req := range in {
                ok := false
                connLogger.Printf("Request type %s", req.Type)
                switch req.Type {
                case "shell":
                    ok = true
                    if len(req.Payload) > 0 {
                        ok = false
                    }
                }
                req.Reply(ok, nil)
            }
        }(requests)

        connLogger.Print("Starting terminal")

        term := terminal.NewTerminal(channel, "> ")

        go func() {
            defer channel.Close()
            for {
                connLogger.Print("Reading from term")
                line, err := term.ReadLine()
                if err != nil {
                    connLogger.Printf("Got error: %s", err)
                    break
                }

                connLogger.Print(line)
                fmt.Println(line)

                if line == "exit" {
                    return
                }
            }
        }()
    }
}

