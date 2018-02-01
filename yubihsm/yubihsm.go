// Package yubihsm implements a cgo wrapper around yubihsm.h provided by the
// YubiHSM2 SDK.
package yubihsm

// #cgo LDFLAGS: -lyubihsm
// #include <stdlib.h>
// #include <stdio.h>
// #define static
// #include <yubihsm.h>
//
// uint16_t *yh_log_entry_length(yh_log_entry *e)      { return &e->length; }
// uint16_t *yh_log_entry_session_key(yh_log_entry *e) { return &e->session_key; }
// uint16_t *yh_log_entry_target_key(yh_log_entry *e)  { return &e->target_key; }
// uint16_t *yh_log_entry_second_key(yh_log_entry *e)  { return &e->second_key; }
//
// yh_object_type *yh_obj_desc_type(yh_object_descriptor *o)      { return &o->type; }
// yh_algorithm   *yh_obj_desc_algorithm(yh_object_descriptor *o) { return &o->algorithm; }
import "C"

import (
	"encoding/asn1"
	"errors"
	"math/big"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func init() {
	if err := rcerr(C.yh_init()); err != nil {
		panic(err)
	}
}

// A Connector represents a connector for communicating with a device.
type Connector C.yh_connector

// Connect instantiates a connector with the given URL and tries to connect.
func Connect(url string) (*Connector, error) {
	c, err := NewConnector(url)
	if err != nil {
		return nil, err
	}
	if err := c.Connect(); err != nil {
		return nil, err
	}
	return c, nil
}

// NewConnector instantiates a new connector with the given URL.
func NewConnector(url string) (*Connector, error) {
	var c *C.yh_connector
	if err := rcerr(C.yh_init_connector(cstr(url), &c)); err != nil {
		return nil, err
	}
	return (*Connector)(c), nil
}

// Connect the connector to the configured URL.
func (c *Connector) Connect() error {
	cc := (*C.yh_connector)(c)
	return rcerr(C.yh_connect_best(&cc, 1, nil))
}

// Disconnect a connected connector.
func (c *Connector) Disconnect() error {
	return rcerr(C.yh_disconnect((*C.yh_connector)(c)))
}

// SetHTTPSCA sets the path to a file with a CA certificate to validate the
// connector with.
func (c *Connector) SetHTTPSCA(path string) error {
	return rcerr(C.yh_set_connector_option((*C.yh_connector)(c),
		C.YH_CONNECTOR_HTTPS_CA, unsafe.Pointer(cstr(path))))
}

// SetProxyServer sets the proxy server to use for connecting to the connector.
func (c *Connector) SetProxyServer(url string) error {
	return rcerr(C.yh_set_connector_option((*C.yh_connector)(c),
		C.YH_CONNECTOR_PROXY_SERVER, unsafe.Pointer(cstr(url))))
}

// Send a plain message, receiving the response into res. Up to cap(res.Data)
// will be used to receive data, though after the call len(res.Data) will
// reflect the length of the data actually received.
func (c *Connector) Send(msg, res *Message) error {
	rsize := C.size_t(cap(res.Data))
	rc := C.yh_send_plain_msg((*C.yh_connector)(c),
		C.yh_cmd(msg.Command), (*C.uint8_t)(&msg.Data[0]), C.size_t(len(msg.Data)),
		(*C.yh_cmd)(&res.Command), (*C.uint8_t)(&res.Data[0]), &rsize)
	if err := rcerr(rc); err != nil {
		return err
	}
	res.Data = res.Data[:int(rsize)]
	return nil
}

// NewDerivedSession creates a new session encrypted with a key derived from
// the password.
func (c *Connector) NewDerivedSession(id int, password []byte, recreate bool, ctx *Context) (*Session, error) {
	s := new(Session)
	rc := C.yh_create_session_derived((*C.yh_connector)(c), C.uint16_t(id),
		(*C.uint8_t)(&password[0]), C.size_t(len(password)),
		C._Bool(recreate), (*C.uint8_t)(&ctx[0]), C.size_t(len(ctx)), &s.session)
	if err := rcerr(rc); err != nil {
		return nil, err
	}
	runtime.SetFinalizer(s, (*Session).destroy)
	return s, nil
}

// NewSession creates a new session encrypted with the given key.
func (c *Connector) NewSession(id int, key, mac []byte, recreate bool, ctx *Context) (*Session, error) {
	s := new(Session)
	rc := C.yh_create_session((*C.yh_connector)(c), C.uint16_t(id),
		(*C.uint8_t)(&key[0]), C.size_t(len(key)),
		(*C.uint8_t)(&mac[0]), C.size_t(len(mac)),
		C._Bool(recreate), (*C.uint8_t)(&ctx[0]), C.size_t(len(ctx)), &s.session)
	if err := rcerr(rc); err != nil {
		return nil, err
	}
	runtime.SetFinalizer(s, (*Session).destroy)
	return s, nil
}

// DeviceInfo gets device info from the connector.
func (c *Connector) DeviceInfo() (*DeviceInfo, error) {
	di := new(DeviceInfo)
	algs := make([]Algorithm, len(C.yh_algorithms))
	nalgs := C.size_t(len(algs))
	rc := C.yh_util_get_device_info((*C.yh_connector)(c),
		(*C.uint8_t)(&di.Major), (*C.uint8_t)(&di.Minor), (*C.uint8_t)(&di.Patch),
		(*C.uint32_t)(&di.Serial), (*C.uint8_t)(&di.LogTotal), (*C.uint8_t)(&di.LogUsed),
		(*C.yh_algorithm)(&algs[0]), &nalgs)
	if err := rcerr(rc); err != nil {
		return nil, err
	}
	di.Algorithms = algs[:int(nalgs)]
	return di, nil
}

// Context for authentication.
type Context [C.YH_CONTEXT_LEN]byte

// A Message represents a message sent to or received from a connector.
type Message struct {
	Command Command // Command in message.
	Data    []byte  // Data in message.
}

// A Session is a session with a device.
type Session struct {
	session *C.yh_session
}

// Authenticate a session.
func (s *Session) Authenticate(ctx *Context) error {
	rc := C.yh_authenticate_session(s.session,
		(*C.uint8_t)(&ctx[0]), C.size_t(len(ctx)))
	return rcerr(rc)
}

// ListObjects lists objects on the device, taking an optional filter.
func (s *Session) ListObjects(f *Filter) ([]*Object, error) {
	var id C.uint16_t
	var typ C.yh_object_type
	var doms C.uint16_t
	var caps C.yh_capabilities
	var alg C.yh_algorithm
	var label *C.char

	if f != nil {
		id = C.uint16_t(f.ID)
		typ = C.yh_object_type(f.Type)
		doms = C.uint16_t(f.Domains)
		caps = C.yh_capabilities(f.Capabilities)
		alg = C.yh_algorithm(f.Algorithm)
		if f.Label != "" {
			label = cstr(f.Label)
		}
	}

	ods := make([]C.yh_object_descriptor, C.YH_MAX_ITEMS_COUNT)
	nods := C.size_t(len(ods))

	rc := C.yh_util_list_objects(s.session,
		id, typ, doms, &caps, alg, label,
		&ods[0], &nods)

	if err := rcerr(rc); err != nil {
		return nil, err
	}

	if len(ods) > int(nods) {
		ods = ods[:int(nods)]
	}

	objects := make([]*Object, len(ods))

	for i := range ods {
		o := new(Object)
		o.unpack(&ods[i])
		objects[i] = o
	}

	return objects, nil
}

// GetObject gets info about an object.
func (s *Session) GetObject(id int, typ ObjectType) (*Object, error) {
	var od C.yh_object_descriptor
	rc := C.yh_util_get_object_info(s.session, C.uint16_t(id), C.yh_object_type(typ), &od)
	if err := rcerr(rc); err != nil {
		return nil, err
	}
	o := new(Object)
	o.unpack(&od)
	return o, nil
}

// GetPublicKey gets a public key from a key object.
func (s *Session) GetPublicKey(id int) ([]byte, error) {
	var algo C.yh_algorithm
	b := make([]byte, 4096)
	sz := C.size_t(len(b))
	rc := C.yh_util_get_pubkey(s.session, C.uint16_t(id), (*C.uint8_t)(&b[0]), &sz, &algo)
	if err := rcerr(rc); err != nil {
		return nil, err
	}
	if int(sz) > len(b) {
		return nil, errors.New("yubihsm: pubkey too big")
	}
	b = b[:int(sz)]
	if Algorithm(algo) != AlgoECK256 {
		return nil, errors.New("yubihsm: only secp256k1 supported")
	}
	return b, nil
}

// GenerateECKey generates a new EC key on the device, returning the object ID.
func (s *Session) GenerateECKey(label string, domains int, caps *Capabilities, algo Algorithm) (id int, err error) {
	if algo != AlgoECK256 {
		return 0, errors.New("yubihsm: only secp256k1 supported")
	}

	var cid C.uint16_t
	var clabel *C.char

	if label != "" {
		clabel = cstr(label)
	}

	rc := C.yh_util_generate_key_ec(s.session, &cid, clabel, C.uint16_t(domains),
		(*C.yh_capabilities)(caps), C.yh_algorithm(algo))

	return int(cid), rcerr(rc)
}

// SignECDSA signs data using ECDSA.
func (s *Session) SignECDSA(id int, data []byte) (R, S *big.Int, err error) {
	out := make([]byte, 2048)
	outlen := C.size_t(len(out))

	rc := C.yh_util_sign_ecdsa(s.session, C.uint16_t(id),
		(*C.uint8_t)(&data[0]), C.size_t(len(data)),
		(*C.uint8_t)(&out[0]), &outlen)

	if err := rcerr(rc); err != nil {
		return nil, nil, err
	}

	if int(outlen) > len(out) {
		return nil, nil, errors.New("yubihsm: result too large")
	}

	out = out[:int(outlen)]

	var rs struct{ R, S *big.Int }

	if rest, err := asn1.Unmarshal(out, &rs); err != nil {
		return nil, nil, err
	} else if len(rest) > 0 {
		return nil, nil, errors.New("yubihsm: unexpected bytes after signature")
	}

	return rs.R, rs.S, nil
}

// Destroy a session, freeing data associated with the session. This will be
// called automatically by a finalizer, but it's safe to call multiple times.
func (s *Session) Destroy() error {
	return rcerr(C.yh_destroy_session(&s.session))
}

// destroy a session, but don't return an error.
func (s *Session) destroy() { s.Destroy() }

// A Filter for filtering lists of objects.
type Filter struct {
	ID           int          // ID to filter by.
	Type         ObjectType   // Type to filter by.
	Domains      int          // Domains to filter by.
	Capabilities Capabilities // Capabilities to filter by.
	Algorithm    Algorithm    // Algorithm to filter by.
	Label        string       // Label to filter by.
}

// Domains encodes a set of domains as an int.
func Domains(domains ...int) int {
	var n int
	for _, d := range domains {
		if d < 1 || d > C.YH_MAX_DOMAINS {
			panic("yubihsm: invalid domain: " + strconv.Itoa(d))
		}
		n |= 1 << uint(d-1)
	}
	return n
}

// Capabilities represent a set of capability supported by an object.
type Capabilities C.yh_capabilities

// Parse a list of strings into a capability set.
func (c *Capabilities) Parse(names ...string) error {
	cs := cstr(strings.Join(names, ","))
	rc := C.yh_capabilities_to_num(cs, (*C.yh_capabilities)(c))
	return rcerr(rc)
}

// ParseCapabilities parses a list of strings into a capability set.
func CapabilitiesByName(names ...string) (*Capabilities, error) {
	c := new(Capabilities)
	if err := c.Parse(names...); err != nil {
		return nil, err
	}
	return c, nil
}

// String returns a string representation of the capabilities.
func (c Capabilities) String() string {
	strs := make([]*C.char, 0xff)
	sz := C.size_t(len(strs))
	rc := C.yh_num_to_capabilities((*C.yh_capabilities)(&c), &strs[0], &sz)
	if err := rcerr(rc); err != nil {
		return "<err: " + err.Error() + ">"
	}
	str := ""
	for _, s := range strs[:int(sz)] {
		if str == "" {
			str = C.GoString(s)
		} else {
			str += "," + C.GoString(s)
		}
	}
	return str
}

// A Capability supported by an object.
type Capability uint8

// CapabilityByName returns the capability with the given name, or 0 if there
// is no capability with that name.
func CapabilityByName(name string) Capability {
	return capabilityByName[name]
}

// String returns the name of a capability.
func (c Capability) String() string {
	if s, ok := capabilityToName[c]; ok {
		return s
	}
	return "<nil>"
}

var capabilityByName = make(map[string]Capability)
var capabilityToName = make(map[Capability]string)

func init() {
	for _, v := range C.yh_capability {
		name := C.GoString(v.name)
		capabilityByName[name] = Capability(v.bit)
		capabilityToName[Capability(v.bit)] = name
	}
}

// A ReturnCode returned by an operation to indicate its completion status.
type ReturnCode C.yh_rc

// Return codes.
const (
	CodeSuccess           = ReturnCode(C.YHR_SUCCESS)             // Success
	ErrMemory             = ReturnCode(C.YHR_MEMORY)              // Memory error
	ErrInitError          = ReturnCode(C.YHR_INIT_ERROR)          // Init error
	ErrNetError           = ReturnCode(C.YHR_NET_ERROR)           // Network error
	ErrConnectorNotFound  = ReturnCode(C.YHR_CONNECTOR_NOT_FOUND) // Connector not found
	ErrInvalidParams      = ReturnCode(C.YHR_INVALID_PARAMS)      // Invalid parameters
	ErrWrongLength        = ReturnCode(C.YHR_WRONG_LENGTH)        // Wrong length
	ErrBufferTooSmall     = ReturnCode(C.YHR_BUFFER_TOO_SMALL)    // Buffer too small
	ErrCryptogramMismatch = ReturnCode(C.YHR_CRYPTOGRAM_MISMATCH) // Cryptogram error
	ErrAuthSessionError   = ReturnCode(C.YHR_AUTH_SESSION_ERROR)  // Authenticate session error
	ErrMACMismatch        = ReturnCode(C.YHR_MAC_MISMATCH)        // MAC not matching

	CodeDeviceOK          = ReturnCode(C.YHR_DEVICE_OK)             // Device success
	ErrInvalidCommand     = ReturnCode(C.YHR_DEVICE_INV_COMMAND)    // Invalid command
	ErrInvalidData        = ReturnCode(C.YHR_DEVICE_INV_DATA)       // Malformed command/data
	ErrInvalidSession     = ReturnCode(C.YHR_DEVICE_INV_SESSION)    // Invalid session
	ErrAuthFail           = ReturnCode(C.YHR_DEVICE_AUTH_FAIL)      // Encryption/verification failed
	ErrSessionsFull       = ReturnCode(C.YHR_DEVICE_SESSIONS_FULL)  // All sessions are allocated
	ErrSessionFailed      = ReturnCode(C.YHR_DEVICE_SESSION_FAILED) // Session creation failed
	ErrStorageFailed      = ReturnCode(C.YHR_DEVICE_STORAGE_FAILED) // Storage failure
	ErrDeviceWrongLength  = ReturnCode(C.YHR_DEVICE_WRONG_LENGTH)   // Wrong length
	ErrInvalidPermissions = ReturnCode(C.YHR_DEVICE_INV_PERMISSION) // Wrong permissions
	ErrLogFull            = ReturnCode(C.YHR_DEVICE_LOG_FULL)       // Log buffer is full
	ErrObjectNotFound     = ReturnCode(C.YHR_DEVICE_OBJ_NOT_FOUND)  // Object not found
	ErrIDIllegal          = ReturnCode(C.YHR_DEVICE_ID_ILLEGAL)     // ID use is illegal
	ErrInvalidOTP         = ReturnCode(C.YHR_DEVICE_INVALID_OTP)    // OTP submitted is invalid
	ErrDemoMode           = ReturnCode(C.YHR_DEVICE_DEMO_MODE)      // Device is in demo mode
	ErrUnexecuted         = ReturnCode(C.YHR_DEVICE_CMD_UNEXECUTED) // Command has not terminated

	ErrGeneric            = ReturnCode(C.YHR_GENERIC_ERROR)        // Unknown error
	ErrDeviceObjectExists = ReturnCode(C.YHR_DEVICE_OBJECT_EXISTS) // Object with that ID already exists
	ErrConnector          = ReturnCode(C.YHR_CONNECTOR_ERROR)      // Connector operation failed
)

// rcerr returns nil if rc is CodeSuccess or CodeDeviceOK, or else returns rc
// as a ReturnCode.
func rcerr(rc C.yh_rc) error {
	if ReturnCode(rc) == CodeSuccess || ReturnCode(rc) == CodeDeviceOK {
		return nil
	}
	return ReturnCode(rc)
}

// Error implements error.
func (c ReturnCode) Error() string {
	return C.GoString(C.yh_strerror(C.yh_rc(c)))
}

// A Command which can be executed on a device.
type Command C.yh_cmd

// Commands identifiers.
const (
	CmdEcho               = Command(C.YHC_ECHO)                  // Echo
	CmdCreateSession      = Command(C.YHC_CREATE_SES)            // Create session
	CmdAuthSession        = Command(C.YHC_AUTH_SES)              // Authenticate session
	CmdSessionMessage     = Command(C.YHC_SES_MSG)               // Session message
	CmdGetDeviceInfo      = Command(C.YHC_GET_DEVICE_INFO)       // Get device info
	CmdBSL                = Command(C.YHC_BSL)                   // BSL
	CmdReset              = Command(C.YHC_RESET)                 // Reset
	CmdCloseSession       = Command(C.YHC_CLOSE_SES)             // Close session
	CmdStats              = Command(C.YHC_STATS)                 // Storage statistics
	CmdPutOpaque          = Command(C.YHC_PUT_OPAQUE)            // Put opaque
	CmdGetOpaque          = Command(C.YHC_GET_OPAQUE)            // Get opaque
	CmdPutAuthkey         = Command(C.YHC_PUT_AUTHKEY)           // Put authentication key
	CmdPutAsymmetricKey   = Command(C.YHC_PUT_ASYMMETRIC_KEY)    // Put asymmetric key
	CmdGenAsymmetricKey   = Command(C.YHC_GEN_ASYMMETRIC_KEY)    // Generate asymmetric key
	CmdSignDataPKCS1      = Command(C.YHC_SIGN_DATA_PKCS1)       // Sign data with PKCS1
	CmdList               = Command(C.YHC_LIST)                  // List objects
	CmdDecryptPKCS1       = Command(C.YHC_DECRYPT_PKCS1)         // Decrypt data with PKCS1
	CmdExportWrapped      = Command(C.YHC_EXPORT_WRAPPED)        // Export an object wrapped
	CmdImportWrapped      = Command(C.YHC_IMPORT_WRAPPED)        // Import a wrapped object
	CmdPutWrapKey         = Command(C.YHC_PUT_WRAP_KEY)          // Put wrap key
	CmdGetLogs            = Command(C.YHC_GET_LOGS)              // Get audit logs
	CmdGetObjectInfo      = Command(C.YHC_GET_OBJECT_INFO)       // Get object information
	CmdPutOption          = Command(C.YHC_PUT_OPTION)            // Put a global option
	CmdGetOption          = Command(C.YHC_GET_OPTION)            // Get a global option
	CmdGetPseudoRandom    = Command(C.YHC_GET_PSEUDO_RANDOM)     // Get pseudo random data
	CmdPutHMACKey         = Command(C.YHC_PUT_HMAC_KEY)          // Put HMAC key
	CmdHMACData           = Command(C.YHC_HMAC_DATA)             // HMAC data
	CmdGetPubkey          = Command(C.YHC_GET_PUBKEY)            // Get a public key
	CmdSignDataPSS        = Command(C.YHC_SIGN_DATA_PSS)         // Sign data with PSS
	CmdSignDataECDSA      = Command(C.YHC_SIGN_DATA_ECDSA)       // Sign data with ECDSA
	CmdDecryptECDH        = Command(C.YHC_DECRYPT_ECDH)          // Perform a ECDH exchange
	CmdDeleteObject       = Command(C.YHC_DELETE_OBJECT)         // Delete an object
	CmdDecryptOAEP        = Command(C.YHC_DECRYPT_OAEP)          // Decrypt data with OAEP
	CmdGenerateHMACKey    = Command(C.YHC_GENERATE_HMAC_KEY)     // Generate HMAC key
	CmdGenerateWrapKey    = Command(C.YHC_GENERATE_WRAP_KEY)     // Generate wrap key
	CmdVerifyHMAC         = Command(C.YHC_VERIFY_HMAC)           // Verify HMAC data
	CmdSSHCertify         = Command(C.YHC_SSH_CERTIFY)           // SSH Certify
	CmdPutTemplate        = Command(C.YHC_PUT_TEMPLATE)          // Put template
	CmdGetTemplate        = Command(C.YHC_GET_TEMPLATE)          // Get template
	CmdOTPDecrypt         = Command(C.YHC_OTP_DECRYPT)           // Decrypt OTP
	CmdOTPAEADCreate      = Command(C.YHC_OTP_AEAD_CREATE)       // Create OTP AEAD
	CmdOTPAEADRandom      = Command(C.YHC_OTP_AEAD_RANDOM)       // Create OTP AEAD from random
	CmdOTPAEADRewrap      = Command(C.YHC_OTP_AEAD_REWRAP)       // Rewrap OTP AEAD
	CmdAttestAsymmetric   = Command(C.YHC_ATTEST_ASYMMETRIC)     // Attest an asymmetric key
	CmdPutOTPAEADKey      = Command(C.YHC_PUT_OTP_AEAD_KEY)      // Put OTP AEAD key
	CmdGenerateOTPAEADKey = Command(C.YHC_GENERATE_OTP_AEAD_KEY) // Generate OTP AEAD key
	CmdSetLogIndex        = Command(C.YHC_SET_LOG_INDEX)         // Set log index
	CmdWrapData           = Command(C.YHC_WRAP_DATA)             // Wrap data
	CmdUnwrapData         = Command(C.YHC_UNWRAP_DATA)           // Unwrap data
	CmdSignDataEDDSA      = Command(C.YHC_SIGN_DATA_EDDSA)       // Sign data with EDDSA
	CmdBlink              = Command(C.YHC_BLINK)                 // Blink the device
	CmdError              = Command(C.YHC_ERROR)                 // Error
)

// An ObjectType represents the type of an object on a device.
type ObjectType C.yh_object_type

// Object types
const (
	TypeOpaque     = ObjectType(C.YH_OPAQUE)       // Opaque object
	TypeAuthKey    = ObjectType(C.YH_AUTHKEY)      // Authentication key
	TypeAsymmetric = ObjectType(C.YH_ASYMMETRIC)   // Asymmetric key
	TypeWrapKey    = ObjectType(C.YH_WRAPKEY)      // Wrap key
	TypeHMACKey    = ObjectType(C.YH_HMACKEY)      // HMAC key
	TypeTemplate   = ObjectType(C.YH_TEMPLATE)     // Template
	TypeOTPAEADKey = ObjectType(C.YH_OTP_AEAD_KEY) // OTP AEAD key
	TypePublic     = ObjectType(C.YH_PUBLIC)       // Public key (virtual)
)

// TypeByName returns the object type with the given name, or 0 if there is no
// object type with that name.
func TypeByName(name string) ObjectType {
	for _, v := range C.yh_types {
		if name == C.GoString(v.name) {
			return ObjectType(v._type)
		}
	}
	return 0
}

// String returns a string representation of the object type.
func (t ObjectType) String() string {
	for _, v := range C.yh_types {
		if t == ObjectType(v._type) {
			return C.GoString(v.name)
		}
	}
	return "<nil>"
}

// An Algorithm represents an algorithm understood by the device.
type Algorithm C.yh_algorithm

// A Digest is a truncated SHA256 digest used in log entries.
type Digest [C.YH_LOG_DIGEST_SIZE]C.uint8_t

// Algorithms understood by the device.
const (
	AlgoRSAPKCS1SHA1    = Algorithm(C.YH_ALGO_RSA_PKCS1_SHA1)
	AlgoRSAPKCS1SHA256  = Algorithm(C.YH_ALGO_RSA_PKCS1_SHA256)
	AlgoRSAPKCS1SHA384  = Algorithm(C.YH_ALGO_RSA_PKCS1_SHA384)
	AlgoRSAPKCS1SHA512  = Algorithm(C.YH_ALGO_RSA_PKCS1_SHA512)
	AlgoRSAPSSSHA1      = Algorithm(C.YH_ALGO_RSA_PSS_SHA1)
	AlgoRSAPSSSHA256    = Algorithm(C.YH_ALGO_RSA_PSS_SHA256)
	AlgoRSAPSSSHA384    = Algorithm(C.YH_ALGO_RSA_PSS_SHA384)
	AlgoRSAPSSSHA512    = Algorithm(C.YH_ALGO_RSA_PSS_SHA512)
	AlgoRSA2048         = Algorithm(C.YH_ALGO_RSA_2048)
	AlgoRSA3072         = Algorithm(C.YH_ALGO_RSA_3072)
	AlgoRSA4096         = Algorithm(C.YH_ALGO_RSA_4096)
	AlgoECP256          = Algorithm(C.YH_ALGO_EC_P256)  // secp256r1
	AlgoECP384          = Algorithm(C.YH_ALGO_EC_P384)  // secp384r1
	AlgoECP521          = Algorithm(C.YH_ALGO_EC_P521)  // secp521r1
	AlgoECK256          = Algorithm(C.YH_ALGO_EC_K256)  // secp256k1
	AlgoECBP256         = Algorithm(C.YH_ALGO_EC_BP256) // brainpool256r1
	AlgoECBP384         = Algorithm(C.YH_ALGO_EC_BP384) // brainpool384r1
	AlgoECBP512         = Algorithm(C.YH_ALGO_EC_BP512) // brainpool512r1
	AlgoHMACSHA1        = Algorithm(C.YH_ALGO_HMAC_SHA1)
	AlgoHMACSHA256      = Algorithm(C.YH_ALGO_HMAC_SHA256)
	AlgoHMACSHA384      = Algorithm(C.YH_ALGO_HMAC_SHA384)
	AlgoHMACSHA512      = Algorithm(C.YH_ALGO_HMAC_SHA512)
	AlgoECDSASHA1       = Algorithm(C.YH_ALGO_EC_ECDSA_SHA1)
	AlgoECECDH          = Algorithm(C.YH_ALGO_EC_ECDH)
	AlgoRSAOAEPSHA1     = Algorithm(C.YH_ALGO_RSA_OAEP_SHA1)
	AlgoRSAOAEPSHA256   = Algorithm(C.YH_ALGO_RSA_OAEP_SHA256)
	AlgoRSAOAEPSHA384   = Algorithm(C.YH_ALGO_RSA_OAEP_SHA384)
	AlgoRSAOAEPSHA512   = Algorithm(C.YH_ALGO_RSA_OAEP_SHA512)
	AlgoAES128CCMWrap   = Algorithm(C.YH_ALGO_AES128_CCM_WRAP)
	AlgoOpaqueData      = Algorithm(C.YH_ALGO_OPAQUE_DATA)
	AlgoOpaqueX509Cert  = Algorithm(C.YH_ALGO_OPAQUE_X509_CERT)
	AlgoMGF1SHA1        = Algorithm(C.YH_ALGO_MGF1_SHA1)
	AlgoMGF1SHA256      = Algorithm(C.YH_ALGO_MGF1_SHA256)
	AlgoMGF1SHA384      = Algorithm(C.YH_ALGO_MGF1_SHA384)
	AlgoMGF1SHA512      = Algorithm(C.YH_ALGO_MGF1_SHA512)
	AlgoSSHTemplate     = Algorithm(C.YH_ALGO_TEMPL_SSH)
	AlgoYubicoOTPAES128 = Algorithm(C.YH_ALGO_YUBICO_OTP_AES128)
	AlgoYubicoAESAuth   = Algorithm(C.YH_ALGO_YUBICO_AES_AUTH)
	AlgoYubicoOTPAES192 = Algorithm(C.YH_ALGO_YUBICO_OTP_AES192)
	AlgoYubicoOTPAES256 = Algorithm(C.YH_ALGO_YUBICO_OTP_AES256)
	AlgoAES192CCMWrap   = Algorithm(C.YH_ALGO_AES192_CCM_WRAP)
	AlgoAES256CCMWrap   = Algorithm(C.YH_ALGO_AES256_CCM_WRAP)
	AlgoECDSASHA256     = Algorithm(C.YH_ALGO_EC_ECDSA_SHA256)
	AlgoECDSASHA384     = Algorithm(C.YH_ALGO_EC_ECDSA_SHA384)
	AlgoECDSASHA512     = Algorithm(C.YH_ALGO_EC_ECDSA_SHA512)
	AlgoED25519         = Algorithm(C.YH_ALGO_EC_ED25519)
	AlgoECP224          = Algorithm(C.YH_ALGO_EC_P224)
)

// AlgorithmByName returns the algorithm with the given name, or 0 if there is
// no algorithm with that name.
func AlgorithmByName(name string) Algorithm {
	for _, v := range C.yh_algorithms {
		if name == C.GoString(v.name) {
			return Algorithm(v.algorithm)
		}
	}
	return 0
}

// String returns the name of a algorithm.
func (a Algorithm) String() string {
	for _, v := range C.yh_algorithms {
		if a == Algorithm(v.algorithm) {
			return C.GoString(v.name)
		}
	}
	return "<nil>"
}

// An Option is a global option.
type Option C.yh_option

// OptionByName returns the option with the given name, or 0 if there is
// no option with that name.
func OptionByName(name string) Option {
	for _, v := range C.yh_options {
		if name == C.GoString(v.name) {
			return Option(v.option)
		}
	}
	return 0
}

// String returns the name of a option.
func (o Option) String() string {
	for _, v := range C.yh_options {
		if o == Option(v.option) {
			return C.GoString(v.name)
		}
	}
	return "<nil>"
}

// Global options
const (
	// Forced audit mode
	OptionForceAudit = Option(C.YH_OPTION_FORCE_AUDIT)
	// Audit logging per command
	OptionCommandAudit = Option(C.YH_OPTION_COMMAND_AUDIT)
)

// A LogEntry is a log entry returned by the device.
type LogEntry struct {
	Number     int        // Number is a monotonically increasing index.
	Command    Command    // Command that was executed.
	Length     int        // Length of in-data.
	SessionKey int        // SessionKey is the ID of the authentication key used.
	TargetKey  int        // TargetKey is the ID of object used.
	SecondKey  int        // SecondKey is the ID of object used.
	Result     ReturnCode // Result of command.
	Systick    uint       // Systick at time of execution.
	Digest     Digest     // Digest of last digest + this entry.
}

// Unpack a C struct into this object.
func (e *LogEntry) unpack(c *C.yh_log_entry) {
	e.Number = int(c.number)
	e.Command = Command(c.command)
	e.Length = int(*C.yh_log_entry_length(c))
	e.SessionKey = int(*C.yh_log_entry_session_key(c))
	e.TargetKey = int(*C.yh_log_entry_target_key(c))
	e.SecondKey = int(*C.yh_log_entry_second_key(c))
	e.Result = ReturnCode(c.result)
	e.Systick = uint(c.systick)
	e.Digest = c.digest
}

// An Object is an object descriptor.
type Object struct {
	Capabilities Capabilities // Capabilities of the object.
	ID           int          // ID of the object.
	Length       int          // Length of the object.
	Domains      int          // Domains of the object.
	Type         ObjectType   // Type of the object.
	Algorithm    Algorithm    // Algorithm associated with the object.
	Sequence     byte         // Sequence number of object.
	Origin       byte         // Origin of object.
	Label        string       // Label of object.

	// DelegatedCapabilities are the object's delegated capabilities.
	DelegatedCapabilities Capabilities
}

// Unpack a C struct into this object.
func (o *Object) unpack(c *C.yh_object_descriptor) {
	o.Capabilities = Capabilities(c.capabilities)
	o.ID = int(c.id)
	o.Length = int(c.len)
	o.Domains = int(c.domains)
	o.Type = ObjectType(*C.yh_obj_desc_type(c))
	o.Algorithm = Algorithm(*C.yh_obj_desc_algorithm(c))
	o.Sequence = byte(c.sequence)
	o.Origin = byte(c.origin)
	o.Label = C.GoString(&c.label[0])
	o.DelegatedCapabilities = Capabilities(c.delegated_capabilities)
}

// DeviceInfo is information about a device.
type DeviceInfo struct {
	Major    uint8  // Major version.
	Minor    uint8  // Minor version.
	Patch    uint8  // Patch version.
	Serial   uint32 // Serial number.
	LogTotal uint8  // Total number of log entries.
	LogUsed  uint8  // Log entries used.

	Algorithms []Algorithm // Algorithms supported by device.
}

// Origin values
const (
	// Origin is generated
	OriginGenerated = C.YH_ORIGIN_GENERATED
	// Origin is imported
	OriginImported = C.YH_ORIGIN_IMPORTED
	// Origin is wrapped (note: this is used in combination with objects'
	// original origin)
	OriginImportedWrapper = C.YH_ORIGIN_IMPORTED_WRAPPED
)

// SetVerbosity sets the logging verbosity of the library.
func SetVerbosity(verbosity int) error {
	return rcerr(C.yh_set_verbosity(C.uint8_t(verbosity)))
}

// GetVerbosity gets the logging verbosity of the library.
func GetVerbosity() (int, error) {
	var v C.uint8_t
	rc := C.yh_get_verbosity(&v)
	return int(v), rcerr(rc)
}

// SetDebugOutput sets the file for debug output.
func SetDebugOutput(file *os.File) error {
	fd, err := syscall.Dup(int(file.Fd()))
	if err != nil {
		return err
	}
	mode := []C.char{'w', 'b', 0}
	cfile := C.fdopen(C.int(fd), &mode[0])
	C.yh_set_debug_output(cfile)
	return nil
}

// cstr creates a null-terminated *C.char from a string for use with C code.
// The reference must be held until the C code is done with it.
func cstr(s string) *C.char {
	cc := make([]C.char, len(s)+1)
	for i := range s {
		cc[i] = C.char(s[i])
	}
	return &cc[0]
}
