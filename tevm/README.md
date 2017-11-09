# Test EVM

A unit-testing library for solidity code

## Into

The `tevm` package makes it easy to compile and test solidity code using `go test`.

In short, the library glues together the `solc` solidity compiler, the `geth` EVM, and `go test`
in order to make testing easy and high-fidelity. The code includes features like content-based
compilation caching to make the comile-test-debug cycle faster.
