package docker

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	mobyclient "github.com/moby/moby/client"
)

// CopyToContainer archives a host directory and extracts it inside a container path.
func (c *Client) CopyToContainer(name, source, dest string) error {
	cont, err := c.GetContainerByName(name)
	if err != nil {
		return err
	}
	var content bytes.Buffer
	if err := compressDir(source, &content); err != nil {
		return err
	}
	if _, err := c.ExecuteCmd(name, []string{"mkdir", "-p", dest}); err != nil {
		return err
	}
	_, err = c.cli.CopyToContainer(c.context(), cont.ID, mobyclient.CopyToContainerOptions{
		DestinationPath: dest,
		Content:         &content,
	})
	return err
}

func compressDir(src string, buf io.Writer) error {
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	srcLength := len(filepath.ToSlash(src))
	err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if file == src {
			return nil
		}
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}
		name := string([]rune(filepath.ToSlash(file))[srcLength:])
		header.Name = name
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			defer data.Close()
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := zr.Close(); err != nil {
		return err
	}
	return nil
}

// WaitForCommand polls a container command until output matches condition.
func (c *Client) WaitForCommand(containerName string, match func(stdout string) bool, command ...string) error {
	for iteration := 0; iteration < 120; iteration++ {
		output, err := c.ExecuteCmd(containerName, command)
		if err == nil && match(output.StdOut) {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timed out waiting for container command %v", command)
}
