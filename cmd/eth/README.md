# eth command-line tool

The `eth` tool is a utility that allows one to manage an Ethereum account from the command line.

## Mode of operation

The `eth` tool, unlike most tools, _always_ sends transactions using "local" signing. In other words,
transactions are always signed locally and then relayed to whatever client has been configured.
Remote (http) and local (unix socket) clients are supported.

Presently, you can use either local key files ("web3 secret storage") or yubihsm2 HSMs for managing keys.

## Command Reference

All commands respect the `-v` command-line flag, which prints additional diagnostic messages.

Additionally, the following environment variables are used:

 - `SETH_URL`: the URL or file path of an http or ipc endpoint to use as an Ethereum client, respectively (defaults to $HOME/.ethereum/geth.ipc)
 - `KEY_PATH`: the relative path in which to look for key files (ending in .json)
 - `ETHER_ADDR`: the address or account name that determines which private key is used for signing

If you are using the yubihsm2, the following environment variables are also used:

 - `YUBIHSM_PASS`: the password used to derive a session with the HSM
 - `YUBIHSM_HINT`: additional comma-separated hints for how to probe for the device's presence

### Balance

The `eth balance` command shows the balance of an account.

Usage:

```
bash-3.2$ eth balance 0xd551234ae421e3bcba99a0da6d736074f22192ff
9126.098576046598288226 ETH
```

The balance can be printed as a raw integer or hex value by using the
`-d` or `-x` flags, respectively.

### Keys

The `eth keys` command lists all of the keys available for signing.

```
$ eth keys
34c2daa8-3bec-767f-ef97-ea6084fc6a51 0x18250eaf72bbaa0237a662b9b85ebd8fa0cf128f
8b52814a-bf25-7103-0135-33a42d4c503e 0x50b26685bc788e164d940f0a73770f4b9196b052
```

### Code

The `eth code` command outputs hex-encoded contract bytecode.

```
SETH_URL=https://api.infura.io/ eth code 0x419d0d8bdd9af5e606ae2232ed285aff190e711b | head -c10
6060604052
```

### Block

The `eth block` command outputs Ethereum blocks (or a range of blocks) as JSON.

The command takes an arbitrary number of block parameter strings, each of which
may be a number, a relative specifier ("latest" or "pending"), or a block range specifier,
which is a number or relative specifier suffixed with "+/-<number>", e.g. "latest-10," which
corresponds to the most recent 10 blocks.

### Keygen

The `eth keygen` command generates a new private key file. You will be prompted
to enter the unlock passphrase for the new key.

The `-o` flag can be used to direct the output to a file; otherwise the file is
emitted to stdout.

### Post

The `eth post` command posts a signed raw transaction to the network.

The transaction should be hex-encoded, and is read from the specified file, or stdin if "-"
is given as a filename.

The command will return the transaction hash and exit immediately; it is the caller's
responsibility to check the status of the transaction.

### Call

The `eth call` command is used for creating and posting raw transactions.

The command accepts at least two arguments, the first of which is the destination contract address, and the second of which is the canonical method specifier as defined in the solidity ABI. If the method specifier includes arguments, then subsequent command-line arguments are interpreted as the types specified in the method selector.

The default behavior of `eth call` includes a sanity check that examines the bytecode of
the destination contract address to see if the method selector is present in the jump table.
You can use the `-f` flag to bypass this check.

Example usage:

```
$ eth call $TOKEN 'transfer(address,uint256)' $DEST $AMOUNT
```

```
$ eth call $CONTRACT 'changeOwner(address)' $DEST
```

You can use the `-n` flag to specify the transaction nonce (e.g. `-n=8`) and the `-g` flag to specify the gas price in gigawei (default is 4).

### Read

The `eth read` command is used for reading chain state (calling "constant" methods on contracts).

The usage for `eth read` is identical to `eth call`, except that additional arguments
are used to specify the return type(s) of the call.

For example:

```
$ eth call $TOKEN 'balanceOf(address)' $ME uint
23414637255007
```

### Sign

The `eth sign` command signs arbitrary data using an Ethereum private key.

The first argument to the command specifies either the file to sign, or "-", which indicates that stdin should be read in its entirety and then signed.

### Recover

The `eth recover` command returns the address (or public key) used to produce a signature.

The command takes two arguments, both hex-encoded: the signature, and the keccak256 hash of the content that was to be signed.
