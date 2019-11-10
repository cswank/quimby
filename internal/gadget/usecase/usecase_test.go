package usecase_test

import (
	"testing"

	"github.com/cswank/quimby/internal/gadget/repository"
	"github.com/cswank/quimby/internal/schema"
	"github.com/cswank/quimby/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestUsecase(t *testing.T) {
	uc := usecase.New(usecase.Repo(&repository.Fake{
		DoGet: func(id int) (schema.Gadget, error) {
			return schema.Gadget{}, nil
		},
	}))

	_, err := uc.Get(1)
	assert.NoError(t, err)
}
