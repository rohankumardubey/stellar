package xdr

import (
	"bytes"
	"fmt"

	"github.com/stellar/go/strkey"
	"github.com/stellar/go/support/errors"
)

// Address returns the strkey encoded form of this signer key.  This method will
// panic if the SignerKey is of an unknown type.
func (skey *SignerKey) Address() string {
	address, err := skey.GetAddress()
	if err != nil {
		panic(err)
	}
	return address
}

// GetAddress returns the strkey encoded form of this signer key, and an error if the
// SignerKey is of an unknown type.
func (skey *SignerKey) GetAddress() (string, error) {
	if skey == nil {
		return "", nil
	}

	vb := strkey.VersionByte(0)
	raw := make([]byte, 32)

	switch skey.Type {
	case SignerKeyTypeSignerKeyTypeEd25519:
		vb = strkey.VersionByteAccountID
		key := skey.MustEd25519()
		copy(raw, key[:])
	case SignerKeyTypeSignerKeyTypeHashX:
		vb = strkey.VersionByteHashX
		key := skey.MustHashX()
		copy(raw, key[:])
	case SignerKeyTypeSignerKeyTypePreAuthTx:
		vb = strkey.VersionByteHashTx
		key := skey.MustPreAuthTx()
		copy(raw, key[:])
	case SignerKeyTypeSignerKeyTypeEd25519SignedPayload:
		vb = strkey.VersionByteSignedPayload
		key := skey.MustEd25519SignedPayload()
		var err error
		raw, err = key.MarshalBinary()
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unknown signer key type: %v", skey.Type)
	}

	return strkey.Encode(vb, raw)
}

// Equals returns true if `other` is equivalent to `skey`
func (skey *SignerKey) Equals(other SignerKey) bool {
	if skey.Type != other.Type {
		return false
	}

	switch skey.Type {
	case SignerKeyTypeSignerKeyTypeEd25519:
		l := skey.MustEd25519()
		r := other.MustEd25519()
		return l == r
	case SignerKeyTypeSignerKeyTypeHashX:
		l := skey.MustHashX()
		r := other.MustHashX()
		return l == r
	case SignerKeyTypeSignerKeyTypePreAuthTx:
		l := skey.MustPreAuthTx()
		r := other.MustPreAuthTx()
		return l == r
	case SignerKeyTypeSignerKeyTypeEd25519SignedPayload:
		l := skey.MustEd25519SignedPayload()
		r := other.MustEd25519SignedPayload()
		return l.Ed25519 == r.Ed25519 && bytes.Equal(l.Payload, r.Payload)
	default:
		panic(fmt.Errorf("Unknown signer key type: %v", skey.Type))
	}
}

func MustSigner(address string) SignerKey {
	aid := SignerKey{}
	err := aid.SetAddress(address)
	if err != nil {
		panic(err)
	}
	return aid
}

// SetAddress modifies the receiver, setting it's value to the SignerKey form
// of the provided address.
func (skey *SignerKey) SetAddress(address string) error {
	if skey == nil {
		return nil
	}

	vb, err := strkey.Version(address)
	if err != nil {
		return errors.Wrap(err, "failed to extract address version")
	}

	var keytype SignerKeyType

	switch vb {
	case strkey.VersionByteAccountID:
		keytype = SignerKeyTypeSignerKeyTypeEd25519
	case strkey.VersionByteHashX:
		keytype = SignerKeyTypeSignerKeyTypeHashX
	case strkey.VersionByteHashTx:
		keytype = SignerKeyTypeSignerKeyTypePreAuthTx
	case strkey.VersionByteSignedPayload:
		keytype = SignerKeyTypeSignerKeyTypeEd25519SignedPayload
	default:
		return errors.Errorf("invalid version byte: %v", vb)
	}

	raw, err := strkey.Decode(vb, address)
	if err != nil {
		return err
	}

	switch vb {
	case strkey.VersionByteAccountID, strkey.VersionByteHashX, strkey.VersionByteHashTx:
		if len(raw) != 32 {
			return errors.New("invalid address")
		}
		var ui Uint256
		copy(ui[:], raw)
		*skey, err = NewSignerKey(keytype, ui)
	case strkey.VersionByteSignedPayload:
		var p SignerKeyEd25519SignedPayload
		err = p.UnmarshalBinary(raw)
		if err != nil {
			return err
		}
		*skey, err = NewSignerKey(keytype, p)
	default:
		return errors.Errorf("invalid version byte: %v", vb)
	}

	return err
}

// SetSignedPayload modifies the receiver, setting it's value to the SignerKey form
// of a signed payload.
func (skey *SignerKey) SetSignedPayload(address string, payload []byte) error {
	if skey == nil {
		return nil
	}

	addressRaw, err := strkey.Decode(strkey.VersionByteAccountID, address)
	if err != nil {
		return err
	}

	var p SignerKeyEd25519SignedPayload
	copy(p.Ed25519[:], addressRaw)
	p.Payload = payload

	*skey, err = NewSignerKey(SignerKeyTypeSignerKeyTypeEd25519SignedPayload, p)

	return err
}
