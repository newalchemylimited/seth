// +build yubihsm

package seth

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/newalchemylimited/seth/yubihsm"
)

const ConnectorURL = "http://127.0.0.1:12345"

func init() {
	RegisterHSM("yubihsm", yubiprobe)
}

type yubictx struct {
	conn *yubihsm.Connector
	ctx  *yubihsm.Context
	sess *yubihsm.Session

	// cache of keys on the device
	pks []HSMKey
}

// Unlock implements HSM.Unlock
func (y *yubictx) Unlock(b []byte) error {
	if y.sess != nil {
		err := y.sess.Destroy()
		y.sess = nil
		if err != nil {
			return err
		}
	}
	sess, err := y.conn.NewDerivedSession(1, b, false, y.ctx)
	if err != nil {
		return err
	}
	if err := sess.Authenticate(y.ctx); err != nil {
		sess.Destroy()
		return nil
	}
	y.sess = sess
	return nil
}

func (y *yubictx) loadk256() error {
	if y.sess == nil {
		return fmt.Errorf("hsm not unlocked")
	}
	if y.pks != nil {
		return nil
	}

	caps, err := yubihsm.CapabilitiesByName("asymmetric_sign_ecdsa")
	if err != nil {
		return err
	}

	objs, err := y.sess.ListObjects(&yubihsm.Filter{
		Type:         yubihsm.TypeAsymmetric,
		Capabilities: *caps,
		Algorithm:    yubihsm.AlgoECK256,
	})
	if err != nil {
		return err
	}

	for i := range objs {
		var spk PublicKey
		pk, err := y.sess.GetPublicKey(objs[i].ID)
		if err != nil {
			return err
		}
		copy(spk[:], pk)
		y.pks = append(y.pks, HSMKey{
			ID:     "yubihsm:" + strconv.Itoa(objs[i].ID),
			Public: spk,
			Aux:    objs[i],
		})
	}
	return nil
}

// Pubkeys implements HSM.Pubkeys
func (y *yubictx) Pubkeys() ([]HSMKey, error) {
	if err := y.loadk256(); err != nil {
		return nil, err
	}
	return y.pks, nil
}

func yubisigner(sess *yubihsm.Session, k *HSMKey) Signer {
	return func(sum *Hash) (*Signature, error) {
		r, s, err := sess.SignECDSA(k.Aux.(*yubihsm.Object).ID, sum[:])
		if err != nil {
			return nil, err
		}

		// Figure out what 'v' should be.
		sig := NewSignature(r, s, 0)
		if pk0, err := NewSignature(r, s, 0).Recover(sum); err != nil {
			return nil, err
		} else if *pk0 != k.Public {
			sig = NewSignature(r, s, 1)
		}

		return sig, nil
	}
}

// Signer implements HSM.Signer
func (y *yubictx) Signer(k *HSMKey) (Signer, error) {
	if y.sess == nil {
		return nil, fmt.Errorf("hsm not unlocked")
	}
	for i := range y.pks {
		if &y.pks[i] == k || y.pks[i].ID == k.ID {
			return yubisigner(y.sess, &y.pks[i]), nil
		}
	}
	return nil, fmt.Errorf("key ID %s not present on hsm", k.ID)
}

func yubiprobe(args ...string) HSM {
	url := ConnectorURL
	for i := range args {
		if strings.HasPrefix(args[i], "http") {
			url = args[i]
		}
	}

	conn, err := yubihsm.Connect(url)
	if err != nil {
		return nil
	}

	return &yubictx{
		conn: conn,
		ctx:  new(yubihsm.Context),
	}
}
