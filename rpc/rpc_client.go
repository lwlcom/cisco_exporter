package rpc

import (
	"errors"
	"fmt"
	"strings"

	"log"

	"github.com/lwlcom/cisco_exporter/connector"
)

const (
	IOSXE string = "IOSXE"
	NXOS  string = "NXOS"
	IOS   string = "IOS"
)

// Client sends commands to a Cisco device
type Client struct {
	conn   *connector.SSHConnection
	Debug  bool
	OSType string
}

// NewClient creates a new client connection
func NewClient(ssh *connector.SSHConnection, debug bool) *Client {
	rpc := &Client{conn: ssh, Debug: debug}

	return rpc
}

// Identify tries to identify the OS running on a Cisco device
func (c *Client) Identify() error {
	output, err := c.RunCommand("show version")
	if err != nil {
		return err
	}
	switch {
	case strings.Contains(output, "IOS XE"):
		c.OSType = IOSXE
	case strings.Contains(output, "NX-OS"):
		c.OSType = NXOS
	case strings.Contains(output, "IOS Software"):
		c.OSType = IOS
	default:
		return errors.New("Unknown OS")
	}
	if c.Debug {
		log.Printf("Host %s identified as: %s\n", c.conn.Host, c.OSType)
	}
	return nil
}

// RunCommand runs a command on a Cisco device
func (c *Client) RunCommand(cmd string) (string, error) {
	if c.Debug {
		log.Printf("Running command on %s: %s\n", c.conn.Host, cmd)
	}
	output, err := c.conn.RunCommand(fmt.Sprintf("%s", cmd))
	if err != nil {
		println(err.Error())
		return "", err
	}

	return output, nil
}
