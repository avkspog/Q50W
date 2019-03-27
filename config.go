package main

import "net"

type ServerConfig struct {
	Host        string
	Port        string
	Version     string
	LogFileName string
}

func (c *ServerConfig) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}
