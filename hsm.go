package seth

import (
	"sync"
)

type hsm struct {
	name  string
	probe func(...string) HSM
}

// HSM is the interface that an HSM (Hardware Security Module)
// driver has to impelement. This interface is deliberately Ethereum-specific.
type HSM interface {
	// Unlock unlocks the hardware security module.
	// It is acceptable for this function to be a no-op
	// if no software unlocking feature exists.
	Unlock(b []byte) error

	// Pubkeys returns the list of secp256k1 public keys
	// on the device, along with the ID of the key.
	Pubkeys() ([]HSMKey, error)

	// Signer should return a Signer function that
	// can be passed to functions like seth.SignTransaction
	Signer(key *HSMKey) (Signer, error)
}

// HSMKey represents a key on an HSM device.
type HSMKey struct {
	ID     string      // ID is an opaque, device-specific identifier
	Public PublicKey   // Public is the public part of the secp256k1 key-pair
	Aux    interface{} // Device-specific data
}

var hsmprobes struct {
	sync.Mutex
	hsms []hsm
}

// RegisterHSM registers an HSM under a name and a
// function to probe if the HSM is installed.
// The probe function can take hints passed from
// FindHSM to help it locate the module.
func RegisterHSM(name string, probe func(hints ...string) HSM) {
	hsmprobes.Lock()
	hsmprobes.hsms = append(hsmprobes.hsms, hsm{name: name, probe: probe})
	hsmprobes.Unlock()
}

// FindHSM finds an HSM based on an HSM name
// and probe hints. If no HSM is found, nil is returned.
func FindHSM(name string, probe ...string) HSM {
	hsmprobes.Lock()
	for i := range hsmprobes.hsms {
		if hsmprobes.hsms[i].name == name {
			p := hsmprobes.hsms[i].probe
			hsmprobes.Unlock()
			return p(probe...)
		}
	}
	hsmprobes.Unlock()
	return nil
}
