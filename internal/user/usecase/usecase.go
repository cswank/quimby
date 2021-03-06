package usecase

import (
	"crypto"

	"github.com/cswank/quimby/internal/schema"
	"github.com/cswank/quimby/internal/user"
	"github.com/cswank/quimby/internal/user/repository"
	"github.com/sec51/twofactor"
	"golang.org/x/crypto/bcrypt"
)

// Usecase does nondatabase-y things
type Usecase struct {
	repo user.Repository
}

func New() *Usecase {
	return &Usecase{
		repo: repository.New(),
	}
}

func (u Usecase) GetAll() ([]schema.User, error) {
	return u.repo.GetAll()
}

func (u Usecase) Get(username string) (schema.User, error) {
	return u.repo.Get(username)
}

func (u Usecase) Create(name, pws string) (*schema.User, []byte, error) {
	pw, err := bcrypt.GenerateFromPassword([]byte(pws), 10)
	if err != nil {
		return nil, nil, err
	}

	tfa, qr, err := u.tfa(name)
	user, err := u.repo.Create(name, pw, tfa)
	return user, qr, err
}

func (u Usecase) Delete(name string) error {
	return u.repo.Delete(name)
}

func (u Usecase) Check(username, pw, token string) error {
	usr, err := u.repo.Get(username)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword(usr.Password, []byte(pw)); err != nil {
		return err
	}

	otp, err := twofactor.TOTPFromBytes(usr.TFA, "quimby")
	if err != nil {
		return err
	}

	return otp.Validate(token)
}

func (u Usecase) tfa(username string) ([]byte, []byte, error) {
	otp, err := twofactor.NewTOTP(username, "quimby", crypto.SHA1, 6)
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
// func (t *TFA) Check(tfaData []byte, token string) error {
// 	if len(tfaData) == 0 {
// 		return fmt.Errorf("no 2fa data")
// 	}
// 	otp, err := twofactor.TOTPFromBytes(tfaData, t.issuer)
// 	if err != nil {
// 		return err
// 	}
// 	return otp.Validate(token)
// }
