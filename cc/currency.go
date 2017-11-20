// package cc is a library for converting
// between cryptocurrencies. It implements
// fixed-point high-precision rate conversion
// and manipulation, and it also contains some
// knowledge about the number of digits on certain
// common cryptocurrencies and ERC20 tokens.
package cc

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/newalchemylimited/seth"
)

// RateDigits is the decimal precision of a Rate,
// where the precision is pow(10, -RateDigits).
const RateDigits = 8

// One is a rate corresponding to one.
var One = NewRate(1)

// Rate represents a fixed-point rate.
type Rate big.Int

// NewRate returns the rate most closely approximating f.
func NewRate(f float64) *Rate {
	var rat big.Rat
	rat.SetFloat64(f)
	num, denom := rat.Num(), rat.Denom()
	if num.Sign() != 0 {
		pow10(num, RateDigits)
		num.Div(num, denom)
	}
	return (*Rate)(num)
}

// ParseRate parses a string into a Rate.
// The string should be a decimal number
// like 123.456, with up to 8 digits after
// the decimal.
func ParseRate(s string) (*Rate, error) {
	r, err := parseDigits(new(big.Int), s, RateDigits)
	if err != nil {
		return nil, err
	}
	return (*Rate)(r), nil
}

// Float returns a floating-point approximation of the rate.
func (r *Rate) Float() float64 {
	var rat big.Rat
	rat.SetFrac((*big.Int)(r), (*big.Int)(One))
	f, _ := rat.Float64()
	return f
}

// Inverse calculates the inverse of the rate.
func (r *Rate) Inverse() *Rate {
	if r.Sign() == 0 {
		return r
	}
	n := big.NewInt(1)
	pow10(n, 2*RateDigits)
	n.Div(n, (*big.Int)(r))
	return (*Rate)(n)
}

// Sign returns the sign of the rate.
func (r *Rate) Sign() int {
	return (*big.Int)(r).Sign()
}

func isdecimal(r rune) bool {
	return isdigit(r) || r == '.' || r == '-' || r == '+'
}

func isdigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// Scan implements fmt.Scanner for the verbs '%s' and '%v'
func (r *Rate) Scan(s fmt.ScanState, verb rune) error {
	switch verb {
	case 's', 'v':
		tok, err := s.Token(true, isdecimal)
		if err != nil {
			return err
		}
		nr, err := ParseRate(string(tok))
		if err != nil {
			return err
		}
		*r = *nr
		return nil
	default:
		return fmt.Errorf("invalid verb %q for Rate", verb)
	}
}

// parseDigits parses a fixed-point decimal number
// with up to 'digits' digits after the '.'
func parseDigits(num *big.Int, s string, digits int) (*big.Int, error) {
	ip, fp := s, ""
	if i := strings.IndexByte(s, '.'); i != -1 {
		ip, fp = s[:i], s[i+1:]
	}
	neg := false
	if ip != "" && ip[0] == '-' {
		neg, ip = true, ip[1:]
	}
	if ip == "" && fp == "" {
		return nil, errors.New("cc: number cannot be empty")
	}
	if ip != "" {
		if !isdigit(rune(ip[0])) {
			return nil, errors.New("cc: invalid character in integer part: " + s)
		}
		if _, ok := num.SetString(ip, 10); !ok {
			return nil, errors.New("cc: error parsing integer part: " + s)
		}
		exp10(num, digits)
	}
	if fp != "" {
		if !isdigit(rune(fp[0])) {
			return nil, errors.New("cc: invalid character in fractional part: " + s)
		}
		var frac big.Int
		if _, ok := frac.SetString(fp, 10); !ok {
			return nil, errors.New("cc: error parsing fractional part: " + s)
		}
		if len(fp) > digits {
			return nil, fmt.Errorf("cc: exceeds %d digits of precision: %s", digits, s)
		}
		exp10(&frac, digits-len(fp))
		num.Add(num, &frac)
	}
	if neg {
		num.Neg(num)
	}
	return num, nil
}

func widestring(b *big.Int, digits int) string {
	if b.Sign() < 0 {
		var nb big.Int
		return "-" + widestring(nb.Neg(b), digits)
	}
	s := b.String()
	if digits == 0 {
		return s
	}
	if len(s) < digits+1 {
		s = strings.Repeat("0", digits+1-len(s)) + s
	}
	split := len(s) - digits
	return s[:split] + "." + s[split:]
}

// String returns the rate as a string like "0.17439082"
func (r *Rate) String() string {
	return widestring((*big.Int)(r), RateDigits)
}

// MarshalJSON implements json.Marshaler.
func (r *Rate) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}
	return []byte(r.String()), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *Rate) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	if len(b) > 0 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	n, err := ParseRate(string(b))
	if err != nil {
		return err
	}
	*r = *n
	return nil
}

type currency struct {
	name   string
	digits int
	addr   seth.Address
}

// Currency represents a currency along with its significant digits.
type Currency string

func (c Currency) resolve() *currency {
	if c := currencies[string(c)]; c != nil {
		return c
	}
	panic("cc: no such currency: " + string(c))
}

// Name of the currency. This panics if the currency is not registered.
func (c Currency) Name() string {
	return c.resolve().name
}

// Digits in the currency. This panics if the currency is not registered.
func (c Currency) Digits() int {
	return c.resolve().digits
}

// Addr returns the contract address, if this represents an ERC20 token. This
// panics if the currency is not registered.
func (c Currency) Addr() *seth.Address {
	if addr := &c.resolve().addr; !addr.Zero() {
		return addr
	}
	return nil
}

// Token returns whether this is an ERC20 token. This panics if the currency is
// not registered.
func (c Currency) Token() bool { return c.Addr() != nil }

// Exists returns whether this currency is registered.
func (c Currency) Exists() bool { return currencies[string(c)] != nil }

// Nil returns whether c == "".
func (c Currency) Nil() bool { return c == "" }

// String returns a string representation of the currency.
func (c Currency) String() string { return string(c) }

// ParseAmount parses a currency amount to the
// maximum precision of the currency.
func (c Currency) ParseAmount(s string) (*Amount, error) {
	a := new(Amount)
	if err := c.parseAmount(a, s); err != nil {
		return nil, err
	}
	return a, nil
}

func (c Currency) parseAmount(a *Amount, s string) error {
	if _, err := parseDigits(&a.Amount, s, c.Digits()); err != nil {
		return err
	}
	a.Currency = c
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (c Currency) MarshalText() ([]byte, error) {
	return []byte(c), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (c *Currency) UnmarshalText(b []byte) error {
	c2, ok := CurrencyByName(string(b))
	if !ok {
		return errors.New("cc: no such currency: " + string(b))
	}
	*c = c2
	return nil
}

func pow10(b *big.Int, n int) {
	if n > 0 {
		exp10(b, n)
	} else if n < 0 {
		div10(b, -n)
	}
}

// b *= 10^n
func exp10(b *big.Int, n int) {
	if n == 0 {
		return
	}
	var ten big.Int
	var e big.Int
	e.SetInt64(int64(n))
	ten.SetInt64(10)
	ten.Exp(&ten, &e, nil)
	b.Mul(b, &ten)
}

// b /= 10^n
func div10(b *big.Int, n int) {
	if n == 0 {
		return
	}
	var ten, e big.Int
	e.SetInt64(int64(n))
	ten.SetInt64(10)
	ten.Exp(&ten, &e, nil)
	b.Div(b, &ten)
}

// A Pair represents a conversion rate between a pair of currencies.
type Pair struct {
	From, To Currency // From and To are the pair currencies.
	Rate     Rate     // Rate is the exchange rate of the pair.
}

// Inverse returns the inverse of the pair.
func (p *Pair) Inverse() *Pair {
	return &Pair{
		From: p.To,
		To:   p.From,
		Rate: *p.Rate.Inverse(),
	}
}

// Convert the given amount.
func (p *Pair) Convert(a *Amount) *Amount {
	if a.Currency == p.From {
		return Convert(a, p.To, &p.Rate)
	}
	if a.Currency == p.To {
		return Convert(a, p.From, p.Rate.Inverse())
	}
	s := a.Currency.String() + " at " + p.String()
	panic("cc: currency mismatch: " + s)
}

// ParsePair parses a pair from a string.
func ParsePair(s string) (*Pair, error) {
	p := new(Pair)
	if err := p.FromString(s); err != nil {
		return nil, err
	}
	return p, nil
}

// String returns a string representation of the pair.
func (p *Pair) String() string {
	from := p.From.String()
	to := p.To.String()
	rate := p.Rate.String()
	return rate + " " + to + "/" + from
}

// FromString parses a pair from a string.
func (p *Pair) FromString(s string) error {
	return p.UnmarshalText([]byte(s))
}

// MarshalText implements encoding.TextMarshaler.
func (p *Pair) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *Pair) UnmarshalText(b []byte) error {
	i1 := bytes.IndexByte(b, ' ')
	i2 := bytes.IndexByte(b, '/')
	if i1 == -1 || i2 == -1 || i2 < i1 {
		return fmt.Errorf("pair: malformed string: %q", b)
	}
	rate, to, from := b[:i1], b[i1+1:i2], b[i2+1:]
	if r, err := ParseRate(string(rate)); err != nil {
		return err
	} else {
		p.Rate = *r
	}
	if err := p.To.UnmarshalText(to); err != nil {
		return err
	}
	return p.From.UnmarshalText(from)
}

// Amount is a combination of a
// currency and a raw amount.
type Amount struct {
	Currency
	Amount big.Int
}

// Assert that the currencies match.
func (a *Amount) assert(c Currency) {
	if a := a.Currency; a != c {
		panic("cc: currency mismatch: " + a.String() + " != " + c.String())
	}
}

// ParseAmount parses an amount from a string.
func ParseAmount(s string) (*Amount, error) {
	a := new(Amount)
	if _, err := fmt.Sscanln(s, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Copy a into a new Amount.
func (a *Amount) Copy() *Amount {
	return new(Amount).Set(a)
}

// Set a to x and return a.
func (a *Amount) Set(x *Amount) *Amount {
	a.Currency = x.Currency
	a.Amount.Set(&x.Amount)
	return a
}

// Add x to a and return a. Panics if currencies do not match.
func (a *Amount) Add(x *Amount) *Amount {
	if a.zero() {
		return a.Set(x)
	}
	a.assert(x.Currency)
	a.Amount.Add(&a.Amount, &x.Amount)
	return a
}

// Sub x from a and return a. Panics if currencies do not match.
func (a *Amount) Sub(x *Amount) *Amount {
	if a.zero() {
		return a.Set(x).Neg()
	}
	a.assert(x.Currency)
	a.Amount.Sub(&a.Amount, &x.Amount)
	return a
}

// Neg negates and returns a.
func (a *Amount) Neg() *Amount {
	a.Amount.Neg(&a.Amount)
	return a
}

// Cmp a to x (see big.Int.Cmp). Panics if currencies do not match.
func (a *Amount) Cmp(x *Amount) int {
	a.assert(x.Currency)
	return a.Amount.Cmp(&x.Amount)
}

// Sign gets the sign of a (see big.Int.Sign).
func (a *Amount) Sign() int {
	return a.Amount.Sign()
}

// zero returns whether this amount is uninitialized.
func (a *Amount) zero() bool {
	return a.Currency.Nil() && a.Amount.Sign() == 0
}

// Scan implements fmt.Scanner for the verbs '%s' and '%v,'
// and supports text produced by (*Amount).String()
func (a *Amount) Scan(s fmt.ScanState, verb rune) error {
	switch verb {
	case 's', 'v':
		ratetok, err := s.Token(true, isdecimal)
		if err != nil {
			return err
		}
		ratestr := string(ratetok)
		curtok, err := s.Token(true, nil)
		if err != nil {
			return err
		}
		c, ok := CurrencyByName(string(curtok))
		if !ok {
			return fmt.Errorf("no currency %q", curtok)
		}
		return c.parseAmount(a, ratestr)
	default:
		return fmt.Errorf("invalid verb %q for Amount", verb)
	}
}

// String prints the currency amount next to the currency name, e.g.
//  124.09123402 BTC
func (a *Amount) String() string {
	return widestring(&a.Amount, a.Digits()) + " " + string(a.Currency)
}

// MarshalText implements encoding.TextMarshaler.
func (a *Amount) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (a *Amount) UnmarshalText(b []byte) error {
	_, err := fmt.Sscanln(string(b), a)
	return err
}

// Convert converts the given amount to a new currency
// using the given Rate. The conversion is done
// with at least RateDigits decimals of precision.
func Convert(a *Amount, to Currency, r *Rate) *Amount {
	var n big.Int
	n.Set(&a.Amount)
	n.Mul(&n, (*big.Int)(r))                     // now n has RateDigits too many zeros
	pow10(&n, to.Digits()-a.Digits()-RateDigits) // correct the number of zeros
	return &Amount{to, n}
}

var currencies = make(map[string]*currency)
var tokens = make(map[seth.Address]Currency)

// NewCurrency creates and registers a new currency. This is not thread-safe
// and should only be used at initialization time.
func NewCurrency(name string, digits int) Currency {
	return registerCurrency(&currency{
		name:   name,
		digits: digits,
	})
}

// NewToken creates and registers a new ERC20 token. This is not thread-safe
// and should only be used at initialization time.
func NewToken(name string, digits int, addr string) Currency {
	return registerCurrency(&currency{
		name:   name,
		digits: digits,
		addr:   mustaddr(addr),
	})
}

func registerCurrency(c *currency) Currency {
	if c == nil {
		panic("cc: nil currency")
	}
	if c.name == "" {
		panic("cc: no currency name")
	}
	if p, ok := currencies[c.name]; ok {
		if *p != *c {
			panic("cc: conflicting currency: " + c.name)
		}
		return Currency(p.name)
	}
	currencies[c.name] = c
	if !c.addr.Zero() {
		if _, ok := tokens[c.addr]; ok {
			panic("cc: conflicting token address: " + c.addr.String())
		}
		tokens[c.addr] = Currency(c.name)
	}
	return Currency(c.name)
}

// Well-known currencies and tokens.
var (
	EUR = NewCurrency("EUR", 8)
	GBP = NewCurrency("GBP", 8)
	USD = NewCurrency("USD", 8)

	BTC = NewCurrency("BTC", 8)
	ETH = NewCurrency("ETH", 18)
	ZEC = NewCurrency("ZEC", 9)

	AION  = NewToken("AION", 8, "0x4CEdA7906a5Ed2179785Cd3A40A69ee8bc99C466")
	AIR   = NewToken("AIR", 8, "0x27dce1ec4d3f72c3e457cc50354f1f975ddef488")
	ANT   = NewToken("ANT", 18, "0x960b236a07cf122663c4303350609a66a7b288c0")
	BAT   = NewToken("BAT", 18, "0x0d8775f648430679a709e98d2b0cb6250d2887ef")
	BNT   = NewToken("BNT", 18, "0x1f573d6fb3f13d689ff844b4ce37794d79a7ff1c")
	DGD   = NewToken("DGD", 9, "0xe0b7927c4af23765cb51314a0e0521a9645f0e2a")
	DICE  = NewToken("DICE", 16, "0x2e071d2966aa7d8decb1005885ba1977d6038a65")
	EDG   = NewToken("EDG", 0, "0x08711d3b02c8758f2fb3ab4e80228418a7f8e39c")
	EOS   = NewToken("EOS", 18, "0x86fa049857e0209aa7d9e616f7eb3b3b78ecfdb0")
	FUCK  = NewToken("FUCK", 4, "0xc63e7b1dece63a77ed7e4aeef5efb3b05c81438d")
	FUN   = NewToken("FUN", 8, "0xbbb1bd2d741f05e144e6c4517676a15554fd4b8d")
	GNO   = NewToken("GNO", 18, "0x6810e776880c02933d47db1b9fc05908e5386b96")
	GNT   = NewToken("GNT", 18, "0xa74476443119a942de498590fe1f2454d7d4ac0d")
	GOOD  = NewToken("GOOD", 6, "0xae616e72d3d89e847f74e8ace41ca68bbf56af79")
	GUP   = NewToken("GUP", 3, "0xf7b098298f7c69fc14610bf71d5e02c60792894c")
	HODL  = NewToken("HODL", 8, "0xb4b7d0c65b3618bc8706ab7b3719519ead624067")
	ICN   = NewToken("ICN", 18, "0x888666ca69e0f178ded6d75b5726cee99a87d698")
	JTT1  = NewToken("JTT1", 8, "0xb1e7688a1cc678a035342b250d348f2c131bd8fb")
	JTT2  = NewToken("JTT2", 8, "0x4b59841ac0fbe6eaa3ad2978dad8d0e1c76a9237")
	MCAP  = NewToken("MCAP", 8, "0x93e682107d1e9defb0b5ee701c71707a4b2e46bc")
	MKR   = NewToken("MKR", 18, "0xc66ea802717bfb9833400264dd12c2bceaa34a6d")
	MLN   = NewToken("MLN", 18, "0xbeb9ef514a379b997e0798fdcc901ee474b6d9a1")
	OMG   = NewToken("OMG", 18, "0xd26114cd6ee289accf82350c8d8487fedb8a0c07")
	OST   = NewToken("1ST", 18, "0xaf30d2a7e90d7dc361c8c4585e9bb7d2f6f15bc7")
	PLBT  = NewToken("PLBT", 6, "0x0affa06e7fbe5bc9a764c979aa66e8256a631f02")
	PLU   = NewToken("PLU", 18, "0xd8912c10681d8b21fd3742244f44658dba12264e")
	QTUM  = NewToken("QTUM", 18, "0x9a642d6b3368ddc662CA244bAdf32cDA716005BC")
	REP   = NewToken("REP", 18, "0x48c80f1f4d53d5951e5d5438b54cba84f29f32a5")
	RLC   = NewToken("RLC", 9, "0x607f4c5bb672230e8672085532f7e901544a7375")
	SCAM  = NewToken("SCAM", 18, "0x49488350b4b2ed2fd164dd0d50b00e7e3f531651")
	SNGLS = NewToken("SNGLS", 0, "0xaec2e87e0a235266d9c5adc9deb4b2e29b54d009")
	SNT   = NewToken("SNT", 18, "0x744d70fdbe2ba4cf95131626614a1763df805b9e")
	SWT   = NewToken("SWT", 18, "0xb9e7f8568e08d5659f5d29c4997173d84cdf2607")
	TKN   = NewToken("TKN", 8, "0xaaaf91d9b90df800df4f55c205fd6989c977e73a")
	UET   = NewToken("UET", 18, "0x27f706edde3ad952ef647dd67e24e38cd0803dd6")
	UNI   = NewToken("🦄", 0, "0x89205a3a3b2a69de6dbf7f01ed13b2108b2c43e7")
	VEROS = NewToken("VEROS", 5, "0xedbaf3c5100302dcdda53269322f3730b1f0416d")
	VSL   = NewToken("VSL", 18, "0x5c543e7ae0a1104f78406c340e9c64fd9fce5170")
	WINGS = NewToken("WINGS", 18, "0x667088b212ce3d06a1b553a7221e1fd19000d9af")
	XAUR  = NewToken("XAUR", 8, "0x4df812f6064def1e5e029f1ca858777cc98d2d81")
	ZRX   = NewToken("ZRX", 18, "0xe41d2489571d322189246dafa5ebde1f4699f498")
)

func mustaddr(s string) (out seth.Address) {
	if err := out.FromString(s); err != nil {
		panic(err)
	}
	return out
}

// CurrencyByName returns the currency with the given name.
func CurrencyByName(s string) (Currency, bool) {
	if c := currencies[s]; c != nil {
		return Currency(c.name), true
	}
	return "", false
}

// TokenByAddress returns the currency with the given contract address.
func TokenByAddress(addr *seth.Address) (Currency, bool) {
	c, ok := tokens[*addr]
	return c, ok
}
