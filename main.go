package main

import (
	"encoding/base64"
	"encoding/hex"
	"log"
	"net/http"
	"os"

	"github.com/elazarl/goproxy"
	"github.com/joho/godotenv"
	"golang.zx2c4.com/go118/netip"
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
	return `private_key=` + privateKey + `
public_key=` + publicKey + `
endpoint=` + endpoint + `
allowed_ip=0.0.0.0/0`
}

func main() {
	log.Printf("Reading environment variables")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var (
		privateKey         = DecodeKey(MustGetEnv("WG_PRIVATE_KEY"))
		publicKey          = DecodeKey(MustGetEnv("WG_PUBLIC_KEY"))
		endpoint           = MustGetEnv("WG_ENDPOINT")
		localIpV4Address   = MustGetEnv("WG_LOCAL_IPV4_ADDRESS")
		dnsAddress         = MustGetEnv("WG_DNS_ADDRESS")
		proxyListenAddress = MustGetEnv("PROXY_LISTEN_ADDRESS")
	)
	log.Printf("Finished reading configuration")

	log.Printf("Creating TUN")
	var (
		localAddresses = []netip.Addr{netip.MustParseAddr(localIpV4Address)}
		dnsAddresses   = []netip.Addr{netip.MustParseAddr(dnsAddress)}
		tunMTU         = 1420
	)
	tun, tnet, err := netstack.CreateNetTUN(localAddresses, dnsAddresses, tunMTU)
	if err != nil {
		log.Panicf("Error creating TUN '%v'", err)
	}
	log.Printf("TUN created")

	log.Printf("Setting up wireguard connection")
	dev := device.NewDevice(tun, conn.NewStdNetBind(), device.NewLogger(device.LogLevelError, ""))
	dev.IpcSet(GenerateConfig(privateKey, publicKey, endpoint))
	dev.Up()
	log.Printf("Connection created")

	log.Printf("Starting proxy server on port %v", proxyListenAddress)
	proxy := goproxy.NewProxyHttpServer()
	proxy.ConnectDial = tnet.Dial
	log.Fatal(http.ListenAndServe(proxyListenAddress, proxy))
}
