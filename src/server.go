package main

import "net"
import "fmt"

func handle(conn net.Conn) {
    fmt.Println("New connection")
}

func main() {
    ln, err := net.Listen("tcp", ":4200")
    if err != nil {
        fmt.Println("error")
    }
    for {
        conn, err := ln.Accept()
        if err != nil {
            continue
        }

        go handle(conn)
    }
}
