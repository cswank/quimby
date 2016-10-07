package mocks

type TFA struct{}

func NewTFA() *TFA {
	return &TFA{}
}

func (t *TFA) Get(username string) ([]byte, []byte, error) {
	return []byte("fake somethign"), []byte("fake other"), nil
}

func (t *TFA) Check(tfaData []byte, token string) error {
	return nil
}
