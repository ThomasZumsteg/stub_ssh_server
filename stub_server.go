package stub_server

import (
    "log"
    "os"
    "net"
    "fmt"
    "io/ioutil"
    "golang.org/x/crypto/ssh"
)

var SSH_LOG_FILE = os.Stdout

func NewSshServer(port int, user, password string) (chan string, chan string) {
    serverLog := log.New(SSH_LOG_FILE, "server: ", 0)

    config := &ssh.ServerConfig{
        PasswordCallback:
            func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
            if c.User() == user && string(pass) == password {
                return nil, nil
            }
            return nil, fmt.Errorf("Could not login %s", c.User())
        },
    }

    privateBytes, _ := ioutil.ReadFile("id_rsa")
    private, _ := ssh.ParsePrivateKey(privateBytes)
    config.AddHostKey(private)

    in, out := make(chan string), make(chan string)
    go func() {
        for {
            listener, _ := net.Listen("tcp", "127.0.0.1:" + string(port))
            serverLog.Printf("Listening for attempt on 127.0.0.1:%d", port)

            nConn, _ := listener.Accept()
            serverLog.Printf("Accepted connection from %s", nConn.RemoteAddr())

            serviceConnection(nConn, config, in, out)
            listener.Close()
        }
    }()

    return in, out
}

func serviceConnection(nConn net.Conn, config *ssh.ServerConfig, in, out chan string) {
    _, chans, reqs, _ := ssh.NewServerConn(nConn, config)
    go ssh.DiscardRequests(reqs)

    for newChannel := range chans {
        if newChannel.ChannelType() != "session" {
            newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
            continue
        }
        channel, requests, _ := newChannel.Accept()
    }
}
