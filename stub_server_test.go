package stub_server

import (
    "testing"
    "io"
    "golang.org/x/crypto/ssh"
)

func dialHost() (io.Reader, io.WriteCloser) {
    // Connect to the ssh server
    config := &ssh.ClientConfig{
        User: "user",
        Auth: []ssh.AuthMethod{
            ssh.Password("password"),
        },
    }

    client, _ := ssh.Dial("tcp", "127.0.0.1:2222", config)
    session, _ := client.NewSession()
    sshOut, _ := session.StdoutPipe()
    sshIn, _ := session.StdinPipe()

    modes := ssh.TerminalModes{
        ssh.ECHO:          1,      // disable echoing
        ssh.TTY_OP_ISPEED: 144000, // input speed = 14.4kbaud
        ssh.TTY_OP_OSPEED: 144000, // output speed = 14.4kbaud
    }

    session.RequestPty("xterm", 80, 0, modes)
    session.Shell()

    return sshOut, sshIn
}


func TestCreateNew(t *testing.T) {
    sshOut, sshIn := NewSshServer()
    command := "Hello, world!"
    sshIn <- command
    if outCommand := <-sshOut; outCommand != command {
        t.Errorf("Expected \"%s\", got \"%s\"", command, outCommand)
    }
}

