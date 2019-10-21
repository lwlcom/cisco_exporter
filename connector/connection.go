package connector

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"sync"

	"time"

	"golang.org/x/crypto/ssh"
)

var (
	cachedConfig *ssh.ClientConfig
	lock         = &sync.Mutex{}
)

func config(user, keyFile string, legacyCiphers bool, timeout int) (*ssh.ClientConfig, error) {
	lock.Lock()
	defer lock.Unlock()

	if cachedConfig != nil {
		return cachedConfig, nil
	}

	pk, err := loadPublicKeyFile(keyFile)
	if err != nil {
		return nil, err
	}

	cachedConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{pk},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	}
	if legacyCiphers {
		cachedConfig.SetDefaults()
		cachedConfig.Ciphers = append(cachedConfig.Ciphers, "aes128-cbc", "3des-cbc")
	}

	return cachedConfig, nil
}

// NewSSSHConnection connects to device
func NewSSSHConnection(host, user, keyFile string, legacyCiphers bool, timeout int, batchSize int) (*SSHConnection, error) {
	if !strings.Contains(host, ":") {
		host = host + ":22"
	}

	c := &SSHConnection{
		Host:          host,
		legacyCiphers: legacyCiphers,
		batchSize:     batchSize,
	}
	err := c.Connect(user, keyFile, timeout)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// SSHConnection encapsulates the connection to the device
type SSHConnection struct {
	client        *ssh.Client
	Host          string
	stdin         io.WriteCloser
	stdout        io.Reader
	session       *ssh.Session
	legacyCiphers bool
	batchSize     int
}

// Connect connects to the device
func (c *SSHConnection) Connect(user, keyFile string, timeout int) error {
	config, err := config(user, keyFile, c.legacyCiphers, timeout)
	if err != nil {
		return err
	}

	c.client, err = ssh.Dial("tcp", c.Host, config)
	if err != nil {
		return err
	}

	session, err := c.client.NewSession()
	if err != nil {
		c.client.Conn.Close()
		return err
	}
	c.stdin, _ = session.StdinPipe()
	c.stdout, _ = session.StdoutPipe()
	modes := ssh.TerminalModes{
		ssh.ECHO:  0,
		ssh.OCRNL: 0,
	}
	session.RequestPty("vt100", 0, 2000, modes)
	session.Shell()
	c.session = session

	c.RunCommand("")
	c.RunCommand("terminal length 0")

	return nil
}

type result struct {
	output string
	err    error
}

// RunCommand runs a command against the device
func (c *SSHConnection) RunCommand(cmd string) (string, error) {
	buf := bufio.NewReader(c.stdout)
	io.WriteString(c.stdin, cmd+"\n")

	outputChan := make(chan result)
	go func() {
		c.readln(outputChan, cmd, buf)
	}()
	select {
	case res := <-outputChan:
		return res.output, res.err
	case <-time.After(cachedConfig.Timeout):
		return "", errors.New("Timeout reached")
	}
}

// Close closes connection
func (c *SSHConnection) Close() {
	if c.client.Conn == nil {
		return
	}
	c.client.Conn.Close()
	if c.session != nil {
		c.session.Close()
	}
}

func loadPublicKeyFile(file string) (ssh.AuthMethod, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

func (c *SSHConnection) readln(ch chan result, cmd string, r io.Reader) {
	re := regexp.MustCompile(`.+#\s?$`)
	buf := make([]byte, c.batchSize)
	loadStr := ""
	for {
		n, err := r.Read(buf)
		if err != nil {
			ch <- result{output: "", err: err}
		}
		loadStr += string(buf[:n])
		if strings.Contains(loadStr, cmd) && re.MatchString(loadStr) {
			break
		}
	}
	loadStr = strings.Replace(loadStr, "\r", "", -1)
	ch <- result{output: loadStr, err: nil}
}
