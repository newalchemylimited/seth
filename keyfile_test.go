package seth

import (
	"bytes"
	"encoding/json"
	"testing"
)

// test vector from https://github.com/ethereum/wiki/wiki/Web3-Secret-Storage-Definition
const keyfilejson = `{
    "crypto" : {
        "cipher" : "aes-128-ctr",
        "cipherparams" : {
            "iv" : "6087dab2f9fdbbfaddc31a909735c1e6"
        },
        "ciphertext" : "5318b4d5bcd28de64ee5559e671353e16f075ecae9f99c7a79a38af5f869aa46",
        "kdf" : "pbkdf2",
        "kdfparams" : {
            "c" : 262144,
            "dklen" : 32,
            "prf" : "hmac-sha256",
            "salt" : "ae3cd4e7013836a3df6bd7241b12db061dbe2c6785853cce422d148a624ce0bd"
        },
        "mac" : "517ead924a9d0dc3124507e3393d175ce3ff7c1e96529c6c555ce9e51205e9b2"
    },
    "id" : "3198bc9c-6672-5ab3-d995-4942343ae5b6",
    "version" : 3
}`

func TestKeyfile(t *testing.T) {
	var k Keyfile
	if err := json.Unmarshal([]byte(keyfilejson), &k); err != nil {
		t.Fatal(err)
	}

	priv, err := k.Private([]byte("testpassword"))
	if err != nil {
		t.Fatal(err)
	}

	var want Address
	want.FromString("0x008aeeda4d805471df9b2a5b0f38a0c3bcba786b")
	addr := priv.Address()
	if !bytes.Equal(want[:], addr[:]) {
		t.Fatalf("didn't derive the right address: %x != %x", want[:], addr[:])
	}
}
