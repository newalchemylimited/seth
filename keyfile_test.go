package seth

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"testing"
)

func TestKeyfile(t *testing.T) {
	t.Parallel()
	// test vectors from https://github.com/ethereum/wiki/wiki/Web3-Secret-Storage-Definition
	testcases := []string{
		`{
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
}`,
		`{
    "crypto" : {
        "cipher" : "aes-128-ctr",
        "cipherparams" : {
            "iv" : "83dbcc02d8ccb40e466191a123791e0e"
        },
        "ciphertext" : "d172bf743a674da9cdad04534d56926ef8358534d458fffccd4e6ad2fbde479c",
        "kdf" : "scrypt",
        "kdfparams" : {
            "dklen" : 32,
            "n" : 262144,
            "r" : 1,
            "p" : 8,
            "salt" : "ab0c7876052600dd703518d6fc3fe8984592145b591fc8fb5c6d43190334ba19"
        },
        "mac" : "2103ac29920d71da29f15d75b4a16dbe95cfd7ff8faea1056c33131d846e3097"
    },
    "id" : "3198bc9c-6672-5ab3-d995-4942343ae5b6",
    "version" : 3
}`,
	}

	for _, keyfilejson := range testcases {
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

	// test that we can generate a key file
	// that produces the right output

	var priv PrivateKey
	rand.Read(priv[:])

	kf := priv.ToKeyfile("test", []byte("password"))
	buf, _ := json.Marshal(kf)

	var out Keyfile
	if err := json.Unmarshal(buf, &out); err != nil {
		t.Fatal(err)
	}

	outpriv, err := out.Private([]byte("password"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(outpriv[:], priv[:]) {
		t.Fatal("keys not equal")
	}
}
