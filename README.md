# Purpose #

The purpose of this project is to develop a simple stub ssh server that can be used for unit testing other programs that communicate with an ssh server. The idea is to create a server end point to simulate a remote server, but with programmatic control over things like sending commands, timeouts, connection errors, etc.

# Example usage #

```
// Start the local ssh server
server, serverIn, serverOut := NewSshServer("2222", "user", "password")

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

serverOut <- "Hello"

buff := make([]byte, 1000)
bytes_read, _ := sshOut.Read(buff)
fmt.Print(buff[:bytes_read])
// Prints "Hello"
```
