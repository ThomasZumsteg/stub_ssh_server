package main

import (
    "fmt"
    "io/ioutil"
    "net"
    "golang.org/x/crypto/ssh"
    "golang.org/x/crypto/ssh/terminal"
)

func main() {
    config := &ssh.ServerConfig{
        PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
            if c.User() == "user" && string(pass) == "password" {
                return nil, nil
            }
            return nil, fmt.Errorf("password not correct for %s", c.User())
        },
    }

    privateBytes, err := ioutil.ReadFile("id_rsa")
    if err != nil {
        panic("Failed to load private key")
    }

    private, err := ssh.ParsePrivateKey(privateBytes)
    if err != nil {
        panic("Failed to parse private key")
    }

    config.AddHostKey(private)

    for {
        listener, err := net.Listen("tcp", "127.0.0.1:2022")
        if err != nil {
            panic(fmt.Sprintf("Failed to accept incoming connection: %s", err))
        }
        nConn, err := listener.Accept()
        if err != nil {
            panic("Failed to accept incoming connection")
        }
        serviceConnection(nConn, config)
        listener.Close()
    }
}

func serviceConnection(nConn net.Conn, config *ssh.ServerConfig) {
    _, chans, reqs, err := ssh.NewServerConn(nConn, config)
    if err != nil {
        panic("Failed handhake")
    }

    go ssh.DiscardRequests(reqs)

    for newChannel := range chans {
        if newChannel.ChannelType() != "session" {
            newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
            continue
        }
        channel, requests, err := newChannel.Accept()
        if err != nil {
            panic("Could not accept channel.")
        }

        go func(in <-chan *ssh.Request) {
            for req := range in {
                ok := false
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

        term := terminal.NewTerminal(channel, "dummy_server> ")

        go func() {
            defer channel.Close()
            for {
                line, err := term.ReadLine()
                if err != nil {
                    break
                }
                if line == "exit" {
                    return
                }
                term.Write([]byte(line))
            }
        }()
    }
}

