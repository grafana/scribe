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

var (
	ErrorNoDockerConfig = errors.New("no docker config found")
)

// AuthConfig represents a single item in the config.json. Typically this is the value of a key in the root of the JSON object which contains a base64 encoded string
// which contains a string with the authentication details, like username/password or access token.
type AuthConfig struct {
	Auth string `json:"auth"`
}

// DockerConfig represents the config.json file.
type DockerConfig struct {
	Auths map[string]AuthConfig `json:"auths"`
}

// RegistryAuth converts the base64 encoded string found in the docker config for the given registry from the <username>:<password> format to a base64 encoded string in the JSON format, with keys for username/password.
// This is needed because authenticating to the Docker registry from the SDK is required in this format.
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

// DefaultConfig attempts to load the docker config given default assumptions, in this order:
// 1. A 'config.json' in the '$DOCKER_CONFIG' directory.
// 2. A '.docker/config.json' file in the $HOME directory.
// TODO: We might want to also look in the current working directory for a config.json
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

	return nil, ErrorNoDockerConfig
}
