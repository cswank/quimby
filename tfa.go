package quimby

import (
	"crypto"
	"fmt"

	"github.com/sec51/twofactor"
)

type TFAer interface {
	Get(*User) ([]byte, []byte, error)
	Check(*User) error
}

type TFA struct {
	issuer string
}

func NewTFA(issuer string) *TFA {
	return &TFA{issuer}
}

//Get generates a otp for the user, stores it in the db,
//and returns the serialized otp (needs to be saved) and PNG
//data to display for google authenticator
func (t *TFA) Get(user *User) ([]byte, []byte, error) {
	otp, err := twofactor.NewTOTP(user.Username, t.issuer, crypto.SHA1, 8)
	if err != nil {
		return nil, nil, err
	}
	data, err := otp.ToBytes()
	if err != nil {
		return nil, nil, err
	}

	qr, err := otp.QR()
	return data, qr, err
}

//Check retuns nil if the user.TFA is valid for
//that user.
func (t *TFA) Check(user *User) error {
	if err := user.Fetch(); err != nil {
		return err
	}
	if len(user.TFAData) == 0 {
		return fmt.Errorf("user has no 2fa data")
	}
	otp, err := twofactor.TOTPFromBytes(user.TFAData, t.issuer)
	if err != nil {
		return err
	}
	return otp.Validate(user.TFA)
}
