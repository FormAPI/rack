package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/convox/rack/pkg/helpers"
	"github.com/convox/stdcli"
)

func init() {
	CLI.Command("cp", "copy files", Cp, stdcli.CommandOptions{
		Flags:    []stdcli.Flag{flagApp, flagRack},
		Usage:    "<[pid:]src> <[pid:]dst>",
		Validate: stdcli.Args(2),
	})
}

func Cp(c *stdcli.Context) error {
	src := c.Arg(0)
	dst := c.Arg(1)

	r, err := cpSource(c, src)
	if err != nil {
		return err
	}

	if err := cpDestination(c, r, dst); err != nil {
		return err
	}

	return nil
}

func cpDestination(c *stdcli.Context, r io.Reader, dst string) error {
	parts := strings.SplitN(dst, ":", 2)

	switch len(parts) {
	case 1:
		abs, err := filepath.Abs(parts[0])
		if err != nil {
			return err
		}

		rr, err := helpers.RebaseArchive(r, "/base", abs)
		if err != nil {
			return err
		}

		return helpers.Unarchive(rr)
	case 2:
		if !strings.HasPrefix(parts[1], "/") {
			return fmt.Errorf("must specify absolute paths for processes")
		}

		rr, err := helpers.RebaseArchive(r, "/base", parts[1])
		if err != nil {
			return err
		}

		return provider(c).FilesUpload(app(c), parts[0], rr)
	default:
		return fmt.Errorf("unknown destination: %s", dst)
	}
}

func cpSource(c *stdcli.Context, src string) (io.Reader, error) {
	parts := strings.SplitN(src, ":", 2)

	switch len(parts) {
	case 1:
		abs, err := filepath.Abs(parts[0])
		if err != nil {
			return nil, err
		}

		r, err := helpers.Archive(abs)
		if err != nil {
			return nil, err
		}

		return helpers.RebaseArchive(r, abs, "/base")
	case 2:
		if !strings.HasPrefix(parts[1], "/") {
			return nil, fmt.Errorf("must specify absolute paths for processes")
		}

		r, err := provider(c).FilesDownload(app(c), parts[0], parts[1])
		if err != nil {
			return nil, err
		}

		return helpers.RebaseArchive(r, parts[1], "/base")
	default:
		return nil, fmt.Errorf("unknown source: %s", src)
	}
}
