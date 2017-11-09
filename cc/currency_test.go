package cc

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParseRate(t *testing.T) {
	ins := []string{
		"0.50000000",
		"0.05000000",
		"0.00000005",
		"0.00000001",
		"1.00000000",
		"123.00000000",
		"123.00000001",
		"-0.50000000",
		"-0.05000000",
		"-5.00000000",
	}

	for _, c := range ins {
		r, err := ParseRate(c)
		if err != nil {
			t.Errorf("%q: %s", c, err)
			continue
		}
		if r.String() != c {
			t.Errorf("%q != %q", c, r.String())
		}

		var nr Rate
		_, err = fmt.Sscanf(c, "%v", &nr)
		if err != nil {
			t.Fatal(err)
		}
		if nr.String() != c {
			t.Errorf("%q != %q", c, nr.String())
		}
	}

	goods := []string{
		"1",
		"1.1",
		"-1",
		"-.1",
		"1.",
		"-0",
	}

	for _, c := range goods {
		_, err := ParseRate(c)
		if err != nil {
			t.Errorf("%q: got error: %v", c, err)
		}
	}

	bads := []string{
		"",
		"-",
		".",
		"-.",
		"..",
		"1..1",
		"+1",
		"0.-1",
		"0.+1",
		"0.0.",
		"0." + strings.Repeat("9", RateDigits+1),
	}

	for _, c := range bads {
		_, err := ParseRate(c)
		if err == nil {
			t.Errorf("%q: expected error", c)
		}
	}
}

// Check that we don't get improper truncation
// for tokens that have very few digits. (SNLGS has zero.)
// We should never end up dividing a SNLGS amount by anything
// until it has been converted to a unit with many more digits.
func TestConvert2(t *testing.T) {
	str := "9 SNGLS 0.10953000"
	var a Amount
	var r Rate
	_, err := fmt.Sscanf(str, "%v %v", &a, &r)
	if err != nil {
		t.Fatal(err)
	}
	usd := Convert(&a, USD, &r)
	want := "0.98577000 USD"
	if usd.String() != want {
		t.Fatalf("%q != %q", want, usd.String())
	}
}

func TestConvert(t *testing.T) {
	testcase := []struct {
		in         string
		out        string
		indigits   int
		outdigits  int
		ratestring string
	}{
		// NOTE: add 'outdigits' number of zeros
		// to the output amount string
		{"1", "0.50000000 TESTOUT", 8, 8, "0.5"},
		{"0.25", "1.000000 TESTOUT", 8, 6, "4.0"},
		{"1", "0.500000 TESTOUT", 10, 6, "0.5"},
		{"1", "0.5000000000 TESTOUT", 6, 10, "0.5"},
		{"2", "1 TESTOUT", 6, 0, "0.5"},
		{"2", "-1 TESTOUT", 6, 0, "-0.5"},
		{"-2", "1 TESTOUT", 6, 0, "-0.5"},
	}

	for _, c := range testcase {
		cur := NewCurrency("TESTIN", c.indigits)
		cur2 := NewCurrency("TESTOUT", c.outdigits)
		r, err := ParseRate(c.ratestring)
		if err != nil {
			t.Fatal(err)
		}
		amt, err := cur.ParseAmount(c.in)
		if err != nil {
			t.Fatal(err)
		}
		out := Convert(amt, cur2, r)
		if out.String() != c.out {
			t.Errorf("%q != %q", out.String(), c.out)
		}
		out = Convert(out, cur, r.Inverse())
		if out.String() != amt.String() {
			t.Errorf("%q != %q", out.String(), amt.String())
		}
		pair := &Pair{
			From: cur,
			To:   cur2,
			Rate: *r,
		}
		out = pair.Convert(amt)
		if out.String() != c.out {
			t.Errorf("%q != %q", out.String(), c.out)
		}
		out = pair.Convert(out)
		if out.String() != amt.String() {
			t.Errorf("%q != %q", out.String(), amt.String())
		}
		delete(currencies, cur.Name())
		delete(currencies, cur2.Name())
	}
}

func TestCurrencyByName(t *testing.T) {
	cases := []struct {
		name string
		want Currency
	}{
		{"USD", USD},
		{"ETH", ETH},
		{"BTC", BTC},
		{"ZEC", ZEC},
		{"REP", REP},
		{"XAUR", XAUR},
	}
	for _, c := range cases {
		cc, ok := CurrencyByName(c.name)
		if !ok {
			t.Errorf("no currency for %q", c.name)
			continue
		}
		if cc.Name() != c.name {
			t.Errorf("CurrencyByName(%q)=%q ???", c.name, cc.Name())
		}
		if cc != c.want {
			t.Errorf("wanted %+v but got %+v", c.want, cc)
		}
	}
}

func TestTokenByAddress(t *testing.T) {
	cases := []struct {
		addr string
		want Currency
	}{
		{"0xa74476443119A942dE498590Fe1f2454d7D4aC0d", GNT},
		{"0x888666CA69E0f178DED6D75b5726Cee99A87D698", ICN},
		{"0x48c80F1f4D53D5951e5D5438B54Cba84f29F32a5", REP},
	}

	for _, c := range cases {
		addr := mustaddr(c.addr)
		tok, ok := TokenByAddress(&addr)
		if !ok {
			t.Errorf("no token for address %s", addr)
			continue
		}
		if tok != c.want {
			t.Errorf("wanted %+v but got %+v", c.want, tok)
		}
	}
}

func TestScanAmount(t *testing.T) {
	// test that these strings produce
	// Amount structs that yield identical
	// output with .String()
	cases := []string{
		"38.088093360000000000 REP",
		"180.266918874367300824 BAT",
		"499999.900000000000000000 SNT",
		"10 SNGLS",
		"-99.0000 FUCK",
	}

	var a Amount
	for _, s := range cases {
		_, err := fmt.Sscanf(s, "%v", &a)
		if err != nil {
			t.Errorf("%q %s", s, err)
			continue
		}
		out := a.String()
		if out != s {
			t.Errorf("%q != %q", s, out)
		}
	}
}

func TestJSON(t *testing.T) {
	var expect, test1, test2 struct {
		A *Amount
		R *Rate
		C Currency
		P *Pair
	}

	known := []byte(`{
		"A":"1.230000000000000000 ETH",
		"R":1.23000000,
		"C":"ETH",
		"P":"300 USD/ETH"
	}`)

	expect.A, _ = ParseAmount("1.23 ETH")
	expect.R, _ = ParseRate("1.23")
	expect.C = ETH
	expect.P, _ = ParsePair("300 USD/ETH")

	b, err := json.Marshal(&expect)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(known, &test1); err != nil {
		t.Fatal(err)
	} else if test1.C != ETH {
		t.Fatal("unexpected currency:", test1.C)
	} else if !reflect.DeepEqual(&test1, &expect) {
		t.Fatal("unexpected value:", &test1)
	}

	if err := json.Unmarshal(b, &test2); err != nil {
		t.Fatal(err)
	} else if test1.C != ETH {
		t.Fatal("unexpected currency:", test2.C)
	} else if !reflect.DeepEqual(&test2, &expect) {
		t.Fatal("unexpected value:", &test2)
	}
}

func TestInverse(t *testing.T) {
	for _, c := range [][2]float64{
		{0, 0}, {1, 1}, {2, 0.5},
	} {
		r1, r2 := NewRate(c[0]), NewRate(c[1])
		if !reflect.DeepEqual(r1, r2.Inverse()) {
			t.Error("mismatch:", r1, "!=", r2.Inverse())
		}
		if !reflect.DeepEqual(r2, r1.Inverse()) {
			t.Error("mismatch:", r2, "!=", r1.Inverse())
		}
		if r1.Float() != c[0] {
			t.Error("mismatch:", r1.Float(), "!=", c[0])
		}
		if r2.Float() != c[1] {
			t.Error("mismatch:", r2.Float(), "!=", c[1])
		}
	}
}

func TestPair(t *testing.T) {
	var (
		from, _ = CurrencyByName("ETH")
		to, _   = CurrencyByName("USD")
		rate, _ = ParseRate("300")
	)
	expect := &Pair{
		From: from,
		To:   to,
		Rate: *rate,
	}
	pair, err := ParsePair("300 USD/ETH")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(pair, expect) {
		t.Error("mismatch:", pair, "!=", expect)
	}
	bads := []string{
		"", " ", "/", " /", "/ ", "100", "100 ", "100 ETH", "100 ETH/", "100 ETH/",
		"100 /ETH", "100ETH/USD", "100  ETH/USD", "100 ETH-USD", "100 ETH//USD",
		"100 ETH/USD ", "100 ETH/USD/", "ETH", "ETH/USD", " ETH/USD", "- ETH/USD",
	}
	for _, b := range bads {
		_, err := ParsePair(b)
		if err == nil {
			t.Errorf("expected error: %q", b)
		}
	}
}
