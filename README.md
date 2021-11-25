[![Build & Publish Docker Image](https://github.com/shimberger/wg-http-proxy/actions/workflows/docker-image.yml/badge.svg)](https://github.com/shimberger/wg-http-proxy/actions/workflows/docker-image.yml)

# wg-http-proxy

This project hacks together the excellent https://github.com/elazarl/goproxy and https://git.zx2c4.com/wireguard-go into an HTTP proxy server which tunnels requests through wireguard.
I needed it for a quick project but it has NOT been audited for privacy leaks. 

DO NOT USE THIS IF YOUR PRIVACY DEPENDS ON IT.

## State of development

This was written in an evening to do exactly what I needed it to do. The vendoring was necessary to fix some compile errors but I suspect it will be obsolete pretty soon.

## Usage

You can use the following environment variables to configure the tunnel either via `.env` file or normal environment variables:

```
WG_PUBLIC_KEY=<base64-encoded-key>
WG_PRIVATE_KEY=<base64-encoded-key>
WG_LOCAL_IPV4_ADDRESS=<ip>
WG_DNS_ADDRESS=<ip>
WG_ENDPOINT=<ip>:<port>
PROXY_LISTEN_ADDRESS=:8080
```

## License

```
MIT License

Copyright (c) 2021 Sebastian Himberger

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
