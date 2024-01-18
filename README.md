# eth-faucet

[![Build](https://img.shields.io/github/actions/workflow/status/chainflag/eth-faucet/build.yml?branch=main)](https://github.com/chainflag/eth-faucet/actions/workflows/build.yml)
[![Release](https://img.shields.io/github/v/release/chainflag/eth-faucet)](https://github.com/chainflag/eth-faucet/releases)
[![Report](https://goreportcard.com/badge/github.com/chainflag/eth-faucet)](https://goreportcard.com/report/github.com/chainflag/eth-faucet)
[![Go](https://img.shields.io/github/go-mod/go-version/chainflag/eth-faucet)](https://go.dev/)
[![License](https://img.shields.io/github/license/chainflag/eth-faucet)](https://github.com/chainflag/eth-faucet/blob/main/LICENSE)

The faucet is a web application with the goal of distributing small amounts of Ether in private and test networks.

## Features

* Allow to configure the funding account via private key or keystore
* Asynchronous processing Txs to achieve parallel execution of user requests
* Rate limiting by ETH address and IP address as a precaution against spam
* Prevent X-Forwarded-For spoofing by specifying the count of reverse proxies

## Get started

### Prerequisites

* Go (v1.21)
* [golangci-lint](https://golangci-lint.run/usage/install/)
* Node.js (v18)
* NPM
* [just](https://github.com/casey/just#installation)
* [mprocs](https://github.com/pvolok/mprocs#installation)
* [watchexec](https://github.com/watchexec/watchexec#install)

### Installation

1. Clone the repository and navigate to the app’s directory
```bash
git clone https://github.com/astriaorg/eth-faucet.git
cd eth-faucet
```

2. Install front end app deps
```bash
just web-install-deps
```

## Usage

Run front end dev server and Go binary together with `mprocs`
```bash
just run-watch
```

### Configuration

**Optional Flags**

The following are the available command-line flags(excluding above wallet flags):

| Flag            | Description                                      | Default Value  |
|-----------------|--------------------------------------------------|----------------|
| -httpport       | Listener port to serve HTTP connection           | 8080           |
| -proxycount     | Count of reverse proxies in front of the server  | 0              |
| -queuecap       | Maximum transactions waiting to be sent          | 100            |
| -faucet.amount  | Number of Ethers to transfer per user request    | 1              |
| -faucet.minutes | Number of minutes to wait between funding rounds | 1440           |
| -faucet.name    | Network name to display on the frontend          | testnet        |

### Docker deployment

```bash
# might need to build image first
just docker-build
# run via Docker image
just docker-run
```

## License

Distributed under the MIT License. See LICENSE for more information.
