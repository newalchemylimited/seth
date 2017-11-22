package seth

import (
	"encoding/json"
	"math/big"
	"testing"
)

func BenchmarkEncodeCall(b *testing.B) {
	var to Address
	var amount big.Int
	var c CallOpts
	to.FromString("0x78bbe6a0fb1a07fd078bf634dcf2a7d0f444d845")
	amount.SetString("111f904273b", 16)
	a := Int(amount)

	for i := 0; i < b.N; i++ {
		c.EncodeCall("transfer(address,uint256)", &to, &a)
	}
}

func TestArgumentEncoding(t *testing.T) {
	var from, to Address
	var amount big.Int
	var c CallOpts
	from.FromString("0x7b79d72f7eb12b62e1d2e95860b7062dd63f7b7a")
	to.FromString("0x78bbe6a0fb1a07fd078bf634dcf2a7d0f444d845")
	amount.SetString("111f904273b", 16)
	a := Int(amount)
	c.From = &from
	c.EncodeCall("transfer(address,uint256)", &to, &a)

	const want = `"0xa9059cbb00000000000000000000000078bbe6a0fb1a07fd078bf634dcf2a7d0f444d84500000000000000000000000000000000000000000000000000000111f904273b"`
	b, _ := json.Marshal(c.Data)
	if string(b) != want {
		t.Errorf("wanted %q\ngot%q", want, b)
	}

	// now go the opposite direction
	const wantret = `"0x00000000000000000000000078bbe6a0fb1a07fd078bf634dcf2a7d0f444d84500000000000000000000000000000000000000000000000000000111f904273b"`
	to = Address{}
	amount.SetString("0", 10)
	d := NewABIDecoder(&to, &amount)
	err := json.Unmarshal([]byte(wantret), d)
	if err != nil {
		t.Fatal(err)
	}
	if to.String() != "0x78bbe6a0fb1a07fd078bf634dcf2a7d0f444d845" {
		t.Error("bad unmarshal of addr")
	}
}

func TestArgumentEncoding2(t *testing.T) {
	words := []string{
		"0x008eb4da847d0d5cfc908959df5b4e9d52492fd20000000000033e07b4d9d580",
		"0x0035a1bbacd4f771b215715cdb108be436418713000000000000acec45ad61cf",
		"0x00de5f818693eb9a0b38f78866c43d1ff67218a40000000000000b8737d85bda",
		"0x41beed1b9e7d78da2f180364f252eaeb2027c30100000000000114ad3c489c80",
	}
	const want = `"0x9a0e4ebb00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000004008eb4da847d0d5cfc908959df5b4e9d52492fd20000000000033e07b4d9d5800035a1bbacd4f771b215715cdb108be436418713000000000000acec45ad61cf00de5f818693eb9a0b38f78866c43d1ff67218a40000000000000b8737d85bda41beed1b9e7d78da2f180364f252eaeb2027c30100000000000114ad3c489c80"`

	sl := make(IntSlice, len(words))
	for i := range words {
		if err := sl[i].FromString(words[i]); err != nil {
			t.Fatal(err)
		}
	}

	var c CallOpts
	c.EncodeCall("multiMint(uint256[])", &sl)
	b, _ := json.Marshal(c.Data)
	if string(b) != want {
		t.Errorf("wanted %q\ngot%q", want, b)
	}
}

func TestArgumentEncoding3(t *testing.T) {
	a1 := IntSlice{*NewInt(1), *NewInt(2)}
	a2 := IntSlice{*NewInt(3), *NewInt(4)}
	const want = `"0x3f8fc7ea` +
		`0000000000000000000000000000000000000000000000000000000000000040` +
		`00000000000000000000000000000000000000000000000000000000000000a0` +
		`0000000000000000000000000000000000000000000000000000000000000002` +
		`0000000000000000000000000000000000000000000000000000000000000001` +
		`0000000000000000000000000000000000000000000000000000000000000002` +
		`0000000000000000000000000000000000000000000000000000000000000002` +
		`0000000000000000000000000000000000000000000000000000000000000003` +
		`0000000000000000000000000000000000000000000000000000000000000004"`
	var c CallOpts
	c.EncodeCall("foo(uint256[],uint256[])", &a1, &a2)
	b, _ := json.Marshal(c.Data)
	return
	if string(b) != want {
		t.Errorf("wanted %q\ngot%q", want, b)
	}

	// input w/o function selector
	stripped := `"0x` + want[11:]
	var r0, r1 IntSlice
	d := NewABIDecoder(&r0, &r1)
	err := json.Unmarshal([]byte(stripped), d)
	if err != nil {
		t.Fatal(err)
	}
	if len(r0) != 2 {
		t.Errorf("bad length %d r0", len(r0))
	}
	if len(r1) != 2 {
		t.Errorf("bad length %d r1", len(r1))
	}
	if r0[0].Int64() != 1 || r0[1].Int64() != 2 || r1[0].Int64() != 3 || r1[1].Int64() != 4 {
		t.Errorf("bad values: %d %d %d %d", r0[0], r0[1], r1[0], r1[1])
	}
}
