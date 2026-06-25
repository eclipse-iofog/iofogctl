package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	dockerregistry "github.com/moby/moby/api/types/registry"
	mobyclient "github.com/moby/moby/client"
)

// PullOptions configures image pull behavior.
type PullOptions struct {
	Username string
	Password string
}

// PullImage pulls an image or verifies it exists locally when pull fails.
func (c *Client) PullImage(image string, opts PullOptions) error {
	ctx := c.context()
	pullOpts := mobyclient.ImagePullOptions{}
	if opts.Username != "" {
		authConfig := dockerregistry.AuthConfig{
			Username: opts.Username,
			Password: opts.Password,
		}
		authJSON, err := json.Marshal(authConfig) // #nosec G117 -- Moby registry auth JSON required by Docker API
		if err != nil {
			return err
		}
		pullOpts.RegistryAuth = base64.URLEncoding.EncodeToString(authJSON)
	}

	reader, err := c.cli.ImagePull(ctx, image, pullOpts)
	if err != nil {
		if found, listErr := c.imageExistsLocally(image); listErr != nil {
			return listErr
		} else if found {
			return nil
		}
		return err
	}
	defer reader.Close()
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return err
	}
	return waitForLocalImage(c, normalizeImageTag(image), 0)
}

func (c *Client) imageExistsLocally(image string) (bool, error) {
	images, err := c.cli.ImageList(c.context(), mobyclient.ImageListOptions{All: true})
	if err != nil {
		return false, err
	}
	target := normalizeImageTag(image)
	for _, img := range images.Items {
		for _, tag := range img.RepoTags {
			if tag == target || tag == image {
				return true, nil
			}
		}
	}
	return false, nil
}

func normalizeImageTag(image string) string {
	if strings.HasPrefix(image, "docker.io/") {
		return image[len("docker.io/"):]
	}
	return image
}

func waitForLocalImage(c *Client, image string, attempt int8) error {
	if attempt >= 18 {
		return util.NewInternalError("Could not find newly pulled image: " + image)
	}
	found, err := c.imageExistsLocally(image)
	if err != nil {
		return util.NewError(fmt.Sprintf("Could not list local images: %v", err))
	}
	if found {
		return nil
	}
	time.Sleep(10 * time.Second)
	return waitForLocalImage(c, image, attempt+1)
}
