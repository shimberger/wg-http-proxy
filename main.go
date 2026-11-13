package main

import (
	"encoding/base64"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/elazarl/goproxy"
	"github.com/joho/godotenv"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
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
	return `private_key=` + privateKey + `
public_key=` + publicKey + `
endpoint=` + endpoint + `
allowed_ip=0.0.0.0/0`
}

func main() {
	log.Printf("Reading environment variables")
	godotenv.Load()
	var (
		privateKey         = DecodeKey(MustGetEnv("WG_PRIVATE_KEY"))
		publicKey          = DecodeKey(MustGetEnv("WG_PUBLIC_KEY"))
		endpoint           = MustGetEnv("WG_ENDPOINT")
		proxyListenAddress = MustGetEnv("PROXY_LISTEN_ADDRESS")
		proxyVerbose       = os.Getenv("PROXY_VERBOSE") == "true"
	)
	log.Printf("Finished reading configuration")

	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		log.Panicf("Error parsing endpoint '%v': %v", endpoint, err)
	}
	addr := net.ParseIP(host)
	if addr == nil {
		ips, err := net.LookupHost(host)
		if err != nil {
			log.Fatalf("Error looking up hostname %v: %v", host, err)
		}
		addr = net.ParseIP(ips[0])
	}

	endpoint = net.JoinHostPort(addr.String(), port)

	log.Printf("Creating TUN")
	wgTun, err := tun.CreateTUN("wg-http-proxy", 1420)
	if err != nil {
		log.Panicf("Error creating TUN '%v'", err)
	}
	log.Printf("TUN created")

	log.Printf("Setting up wireguard connection")
	dev := device.NewDevice(wgTun, conn.NewStdNetBind(), device.NewLogger(device.LogLevelError, ""))
	dev.IpcSet(GenerateConfig(privateKey, publicKey, endpoint))
	dev.Up()
	log.Printf("Connection created")

	log.Printf("Starting proxy server on port %v", proxyListenAddress)
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = proxyVerbose
	log.Fatal(http.ListenAndServe(proxyListenAddress, proxy))
}
