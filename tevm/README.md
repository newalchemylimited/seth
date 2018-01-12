# Test EVM

A unit-testing library for solidity code

## Intro

The `tevm` package makes it easy to compile and test solidity code using `go test`.

In short, the library glues together the `solc` solidity compiler, the `geth` EVM, and `go test`
in order to make testing easy and high-fidelity. The code includes features like content-based
compilation caching to make the comile-test-debug cycle faster.


## Setup

Make sure you have the latest version of geth by running 

`go get -t -d -u github.com/ethereum/go-ethereum`


Change into the Test EVM Daemon project directory

`cd seth/tevm/tevmd`


Install the Go package

`go install`


Run the daemon

`tevmd`

You should see output similar to the following:

```
2018/01/12 15:25:06 default account: 0x52fdfc072182654f163f5f0f9a621d729566c74d
2018/01/12 15:25:06 binding to :8043...
```

The default account will be funded with 1 eth and node will be listening on http://localhost:8043.
