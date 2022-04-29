package docker

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
)

type AuthConfig struct {
	Auth string `json:"auth"`
}

type DockerConfig struct {
	Auths map[string]AuthConfig `json:"auths"`
}

// TODO: This really has to exist somewhere in the docker SDK. We should try to find it.
func (c *DockerConfig) RegistryAuth(registry string) (string, error) {
	auth, ok := c.Auths[registry]

	if !ok {
		return "", errors.New("auth for registry not found")
	}

	decoded, err := base64.StdEncoding.DecodeString(auth.Auth)
	if err != nil {
		return "", err
	}

	s := strings.Split(string(decoded), ":")
	if len(s) != 2 {
		return "", errors.New("invalid docker auth format")
	}

	b, err := json.Marshal(types.AuthConfig{
		Username: s[0],
		Password: s[1],
	})

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func DefaultConfig() (*DockerConfig, error) {
	var (
		dockerConfig = os.Getenv("DOCKER_CONFIG")
		home         = os.Getenv("HOME")
		dirs         = []string{}
	)

	if dockerConfig != "" {
		dirs = append(dirs, filepath.Join(dockerConfig, "config.json"))
	}
	if home != "" {
		dirs = append(dirs, filepath.Join(home, ".docker", "config.json"))
	}

	cfg := &DockerConfig{}
	for _, v := range dirs {
		// Skip it if the file is not found or is a directory
		if info, err := os.Stat(v); err != nil {
			continue
		} else {
			if info.IsDir() {
				continue
			}
		}

		f, err := os.Open(v)
		if err != nil {
			return nil, err
		}

		if err := json.NewDecoder(f).Decode(cfg); err != nil {
			return nil, err
		}

		return cfg, nil
	}

	return nil, errors.New("no docker config found")
}
