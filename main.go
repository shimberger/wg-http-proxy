package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/joho/godotenv"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

func DecodeKey(s string) string {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Panicf("Error decoding either private or public key from base64: '%v'", err)
	}
	return hex.EncodeToString(b)
}

func MustGetEnv(s string) string {
	v, ok := os.LookupEnv(s)
	if !ok {
		log.Panicf("Could not read required environment variable '%v'", s)
	}
	return v
}

func GenerateConfig(privateKey, publicKey, endpoint string) string {
	return `private_key=` + DecodeKey(privateKey) + `
public_key=` + DecodeKey(publicKey) + `
endpoint=` + endpoint + `
allowed_ip=0.0.0.0/0`
}

func main() {
	// Read the configuration from environment variables
	log.Printf("Reading environment variables")
	godotenv.Load()
	var (
		privateKey         = MustGetEnv("WG_PRIVATE_KEY")
		publicKey          = MustGetEnv("WG_PUBLIC_KEY")
		endpoint           = MustGetEnv("WG_ENDPOINT")
		localIpV4Address   = MustGetEnv("WG_LOCAL_IPV4_ADDRESS")
		dnsAddress         = MustGetEnv("WG_DNS_ADDRESS")
		wgTestendpoint     = os.Getenv("WG_TEST_ENDPOINT")
		proxyTestUrl       = os.Getenv("PROXY_TEST_URL")
		proxyListenAddress = MustGetEnv("PROXY_LISTEN_ADDRESS")
		verboseLogging     = os.Getenv("LOGGING_VERBOSE") == "true"
		config             = GenerateConfig(privateKey, publicKey, endpoint)
		mtu                = 1420
	)
	log.Printf("Finished reading environment variables")

	// Print the read configuration for verification
	log.Printf("Local IP address: %v", localIpV4Address)
	log.Printf("DNS address: %v", dnsAddress)
	log.Printf("Proxy listen address: %v", proxyListenAddress)
	log.Printf("WireGuard test endpoint: %v", wgTestendpoint)
	log.Printf("Proxy test URL: %v", proxyTestUrl)
	log.Printf("WireGuard endpoint: %v", endpoint)
	log.Printf("WireGuard private key: %v", maskString(privateKey))
	log.Printf("WireGuard public key: %v", publicKey)
	log.Printf("WireGuard MTU: %v", mtu)
	log.Printf("Verbose logging: %v", verboseLogging)

	// Create the TUN networking interface
	log.Printf("Creating TUN networking interface")
	tun, tnet, err := netstack.CreateNetTUN(
		[]netip.Addr{netip.MustParseAddr(localIpV4Address)},
		[]netip.Addr{netip.MustParseAddr(dnsAddress)},
		mtu,
	)
	if err != nil {
		log.Panicf("Error creating TUN interface '%v'", err)
	}
	defer tun.Close()
	log.Printf("TUN interface created")

	// Setting up wireguard
	log.Printf("Setting up wireguard connection")
	var logger *device.Logger
	if verboseLogging {
		logger = device.NewLogger(device.LogLevelVerbose, "")
	} else {
		logger = device.NewLogger(device.LogLevelError, "")
	}
	dev := device.NewDevice(tun, conn.NewStdNetBind(), logger)
	defer dev.Close()
	err = dev.IpcSet(config)
	if err != nil {
		log.Panicf("Error setting IPC config '%v'", err)
	}
	err = dev.Up()
	if err != nil {
		log.Panicf("Error bringing up device '%v'", err)
	}
	err = printWireGuardState(dev)
	if err != nil {
		log.Panicf("WireGuard state error: %v", err)
	}
	if wgTestendpoint != "" {
		err = testWireGuard(wgTestendpoint, tnet, 5*time.Second)
		if err != nil {
			log.Panicf("WireGuard test error: %v", err)
		}
	}
	log.Printf("Connection created")

	log.Printf("Starting proxy server on port %v", proxyListenAddress)
	proxy := goproxy.NewProxyHttpServer()
	proxy.ConnectDial = tnet.Dial
	proxy.Tr.Dial = tnet.Dial
	proxy.Verbose = verboseLogging
	if proxyTestUrl != "" {
		err = testProxyConnection(proxy, proxyTestUrl, 5*time.Second)
		if err != nil {
			log.Panicf("Proxy test error: %v", err)
		}
	}
	log.Fatal(http.ListenAndServe(proxyListenAddress, proxy))
}

func testProxyConnection(proxy *goproxy.ProxyHttpServer, url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("Proxy test failed: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Proxy test failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Proxy test returned non-OK status: %v", resp.Status)
	}
	log.Println("Proxy test successful!")
	return nil
}

func printWireGuardState(dev *device.Device) error {
	state, err := dev.IpcGet()
	if err != nil {
		return fmt.Errorf("Error getting WireGuard state: %w", err)
	}
	for _, line := range strings.Split(state, "\n") {
		if strings.HasPrefix(line, "endpoint=") ||
			strings.HasPrefix(line, "last_handshake_time_sec=") ||
			strings.HasPrefix(line, "tx_bytes=") ||
			strings.HasPrefix(line, "rx_bytes=") {
			log.Println(line)
		}
	}
	return nil
}

func testWireGuard(endpoint string, tnet *netstack.Net, timeout time.Duration) error {
	log.Printf("Testing wireguard TCP connection to %v", endpoint)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		timeout,
	)
	defer cancel()
	conn, err := tnet.DialContext(
		ctx,
		"tcp",
		endpoint,
	)
	if err != nil {
		return fmt.Errorf("WireGuard TCP test failed: %w", err)
	}
	defer conn.Close()

	log.Println("WireGuard TCP connection successful!")
	return nil
}

func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}
