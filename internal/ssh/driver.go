package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	xssh "golang.org/x/crypto/ssh"
)

// Config 定义 SSH 连接配置。
// 优先支持私钥认证；当没有私钥时，可回退到密码认证。
type Config struct {
	Address              string
	Port                 int
	User                 string
	Password             string
	PrivateKey           string
	PrivateKeyPassphrase string
	Timeout              time.Duration
	HostKeyCallback      xssh.HostKeyCallback
}

// Driver 维护一个可复用的 SSH 客户端连接。
type Driver struct {
	cfg    Config
	mu     sync.Mutex
	client *xssh.Client
}

// New 创建一个 SSH Driver。
func New(cfg Config) (*Driver, error) {
	if strings.TrimSpace(cfg.Address) == "" {
		return nil, errors.New("ssh address is required")
	}
	if strings.TrimSpace(cfg.User) == "" {
		return nil, errors.New("ssh user is required")
	}
	if cfg.Port <= 0 {
		cfg.Port = 22
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.HostKeyCallback == nil {
		cfg.HostKeyCallback = xssh.InsecureIgnoreHostKey()
	}
	return &Driver{cfg: cfg}, nil
}

// Connect 建立 SSH 连接；重复调用会复用已有连接。
func (d *Driver) Connect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.client != nil {
		return nil
	}

	authMethods, err := d.authMethods()
	if err != nil {
		return err
	}

	clientCfg := &xssh.ClientConfig{
		User:            d.cfg.User,
		Auth:            authMethods,
		HostKeyCallback: d.cfg.HostKeyCallback,
		Timeout:         d.cfg.Timeout,
	}

	addr := net.JoinHostPort(d.cfg.Address, strconv.Itoa(d.cfg.Port))
	client, err := xssh.Dial("tcp", addr, clientCfg)
	if err != nil {
		return err
	}
	d.client = client
	return nil
}

// Close 关闭底层 SSH 连接。
func (d *Driver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.client == nil {
		return nil
	}
	err := d.client.Close()
	d.client = nil
	return err
}

// Run 执行远程命令并返回标准输出和标准错误。
func (d *Driver) Run(ctx context.Context, command string) (string, string, error) {
	if err := d.Connect(); err != nil {
		return "", "", err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	d.mu.Lock()
	client := d.client
	d.mu.Unlock()
	if client == nil {
		return "", "", errors.New("ssh client is not connected")
	}

	session, err := client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	errCh := make(chan error, 1)
	go func() {
		errCh <- session.Run(command)
	}()

	select {
	case err := <-errCh:
		return stdoutBuf.String(), stderrBuf.String(), err
	case <-ctx.Done():
		_ = session.Close()
		err := <-errCh
		if err == nil {
			err = ctx.Err()
		}
		return stdoutBuf.String(), stderrBuf.String(), err
	}
}

// Ping 用于探测 SSH 连通性。
func (d *Driver) Ping(ctx context.Context) error {
	_, _, err := d.Run(ctx, "true")
	return err
}

func (d *Driver) authMethods() ([]xssh.AuthMethod, error) {
	methods := make([]xssh.AuthMethod, 0, 2)

	if strings.TrimSpace(d.cfg.PrivateKey) != "" {
		signer, err := signerFromKey(d.cfg.PrivateKey, d.cfg.PrivateKeyPassphrase)
		if err != nil {
			return nil, err
		}
		methods = append(methods, xssh.PublicKeys(signer))
	}

	if strings.TrimSpace(d.cfg.Password) != "" {
		methods = append(methods, xssh.Password(d.cfg.Password))
	}

	if len(methods) == 0 {
		return nil, errors.New("no ssh auth method provided")
	}
	return methods, nil
}

func signerFromKey(keyValue string, passphrase string) (xssh.Signer, error) {
	keyBytes, err := readKeyMaterial(keyValue)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(passphrase) != "" {
		return xssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
	}
	return xssh.ParsePrivateKey(keyBytes)
}

func readKeyMaterial(value string) ([]byte, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, errors.New("ssh private key is empty")
	}

	if strings.Contains(trimmed, "BEGIN ") {
		return []byte(value), nil
	}

	if info, err := os.Stat(value); err == nil && !info.IsDir() {
		return os.ReadFile(filepath.Clean(value))
	}

	return []byte(value), nil
}

// DialAndPing 是一个便捷函数，用于快速验证 SSH 配置是否可用。
func DialAndPing(cfg Config, timeout time.Duration) error {
	driver, err := New(cfg)
	if err != nil {
		return err
	}
	defer driver.Close()

	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return driver.Ping(ctx)
}

// RunCommand 是一次性执行远程命令的便捷函数。
func RunCommand(ctx context.Context, cfg Config, command string) (string, string, error) {
	driver, err := New(cfg)
	if err != nil {
		return "", "", err
	}
	defer driver.Close()

	return driver.Run(ctx, command)
}

// ValidateConfig 用于在创建实例前做最小化校验。
func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.Address) == "" {
		return fmt.Errorf("ssh address is required")
	}
	if strings.TrimSpace(cfg.User) == "" {
		return fmt.Errorf("ssh user is required")
	}
	if strings.TrimSpace(cfg.PrivateKey) == "" && strings.TrimSpace(cfg.Password) == "" {
		return fmt.Errorf("ssh auth is required")
	}
	return nil
}
