package main

import (
	b64 "encoding/base64"
	"encoding/hex"
	"os"

	"log"
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/joho/godotenv"
	"golang.zx2c4.com/go118/netip"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

func DecodeKey(s string) string {
	b, _ := b64.StdEncoding.DecodeString(s)
	return hex.EncodeToString(b)
}

func MustGetEnv(s string) string {
	v, ok := os.LookupEnv(s)
	if !ok {
		log.Panicf("Could not read required environment variable '%v'", s)
	}
	return v
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		privateKey       = DecodeKey(MustGetEnv("WG_PRIVATE_KEY"))
		publicKey        = DecodeKey(MustGetEnv("WG_PUBLIC_KEY"))
		endpoint         = MustGetEnv("WG_ENDPOINT")
		localIpV4Address = MustGetEnv("WG_LOCAL_IPV4_ADDRESS")
		dnsAddress       = MustGetEnv("WG_DNS_ADDRESS")
	)

	tun, tnet, err := netstack.CreateNetTUN(
		[]netip.Addr{netip.MustParseAddr(localIpV4Address)},
		[]netip.Addr{netip.MustParseAddr(dnsAddress)},
		1420)
	if err != nil {
		log.Panic(err)
	}
	bind := conn.NewStdNetBind()
	log.Printf("PublicKey '%v'", publicKey)
	dev := device.NewDevice(tun, bind, device.NewLogger(device.LogLevelError, ""))
	dev.IpcSet(`private_key=` + privateKey + `
public_key=` + publicKey + `
endpoint=` + endpoint + `
allowed_ip=0.0.0.0/0
`)
	dev.Up()
	/*
		client := http.Client{
			Transport: &http.Transport{
				DialContext: tnet.DialContext,
			},
		}

			resp, err := client.Get("https://ifconfig.me")
			if err != nil {
				log.Panic(err)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Panic(err)
			}
			log.Println(string(body))
	*/
	proxy := goproxy.NewProxyHttpServer()
	//proxy.Verbose = true
	proxy.ConnectDial = tnet.Dial
	log.Fatal(http.ListenAndServe(":8090", proxy))
}
