package mlxutils

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMstConfigReadData_ConnectX6(t *testing.T) {
	sut := New(makeStubCmd(t))

	sut.MstConfigReadData("0000:ca:00.0")
}

func makeStubCmd(t *testing.T) *stubCmd {
	return &stubCmd{
		baseFolder: "./testdata",
		t: t,
	}
}

type stubCmd struct {
	baseFolder string
	t *testing.T
}

func (s *stubCmd) RunCommand(cmd string, args ...string) (string, string, error) {
	stubOutputFile := path.Join(
		s.baseFolder,
		strings.Join([]string{cmd, args}),
	)

	data, err := os.ReadFile(stubOutputFile)
	assert.NoError(s.t, err)

	return string(data), "", nil
}

func (s *stubCmd) Chroot(_ string) (func() error, error) {
	panic("not implemented")
}
