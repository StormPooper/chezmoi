package cmd

import (
	"archive/tar"
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/go-vfs/vfst"
)

func TestImportCmd(t *testing.T) {
	b := &bytes.Buffer{}
	w := tar.NewWriter(b)
	assert.NoError(t, w.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     "dir",
		Mode:     0755,
	}))
	assert.NoError(t, w.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "dir/file",
		Size:     int64(len("contents")),
		Mode:     0644,
	}))
	_, err := w.Write([]byte("contents"))
	assert.NoError(t, err)
	assert.NoError(t, w.WriteHeader(&tar.Header{
		Typeflag: tar.TypeSymlink,
		Name:     "symlink",
		Linkname: "target",
	}))
	assert.NoError(t, w.Close())

	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user/.local/share/chezmoi": &vfst.Dir{Perm: 0700},
	})
	require.NoError(t, err)
	defer cleanup()

	c := &Config{
		fs:        fs,
		mutator:   chezmoi.NewVerboseMutator(os.Stdout, chezmoi.NewFSMutator(fs), false),
		SourceDir: "/home/user/.local/share/chezmoi",
		stdin:     b,
	}
	assert.NoError(t, c.runImportCmd(nil, nil))

	vfst.RunTests(t, fs, "test",
		vfst.TestPath("/home/user/.local/share/chezmoi/dir",
			vfst.TestIsDir,
		),
		vfst.TestPath("/home/user/.local/share/chezmoi/dir/file",
			vfst.TestModeIsRegular,
			vfst.TestContentsString("contents"),
		),
		vfst.TestPath("/home/user/.local/share/chezmoi/symlink_symlink",
			vfst.TestModeIsRegular,
			vfst.TestContentsString("target"),
		),
	)
}
