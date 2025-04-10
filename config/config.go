package config

import (
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Ferlab-Ste-Justine/gitsync/logger"
)

type ConfigFilesystem struct {
	Path           string
	FilesPermission       string `yaml:"files_permission"`
	DirectoriesPermission string `yaml:"directories_permission"`
}

type ConfigGitAuth struct {
	SshKey   string `yaml:"ssh_key"`
	KnownKey string `yaml:"known_key"`
	User     string
}

type ConfigGit struct {
	Repo               string
	Ref                string
	Path               string
	Auth               ConfigGitAuth
	AcceptedSignatures string        `yaml:"accepted_signatures"`
}

type ConfigGrpcAuth struct {
	CaCert     string `yaml:"ca_cert"`
	ClientCert string `yaml:"client_cert"`
	ClientKey  string `yaml:"client_key"`
}

type ConfigGrpcNotifications struct {
	Endpoint          string
	Filter            string
	FilterRegex       *regexp.Regexp `yaml:"-"`
	TrimKeyPath       bool           `yaml:"trim_key_path"`
	MaxChunkSize      uint64         `yaml:"max_chunk_size"`
	Auth              ConfigGrpcAuth
}

type Config struct {
	Filesystem                 ConfigFilesystem
	Git                        ConfigGit
	GrpcNotifications          []ConfigGrpcNotifications `yaml:"grpc_notifications"`
	NotificationCommand        []string                  `yaml:"notification_command"`
	NotificationCommandRetries uint64                    `yaml:"notification_command_retries"`
	LogLevel                   string                    `yaml:"log_level"`
}

func (c *Config) GetLogLevel() int64 {
	logLevel := strings.ToLower(c.LogLevel)
	switch logLevel {
	case "error":
		return logger.ERROR
	case "warning":
		return logger.WARN
	case "debug":
		return logger.DEBUG
	default:
		return logger.INFO
	}
}

func checkConfigIntegrity(c Config) error {
	if c.Filesystem.Path == "" {
		return errors.New("Configuration error: Filesystem path cannot be empty")
	}

	parsedPermission, err := strconv.ParseInt(c.Filesystem.FilesPermission, 8, 32)
	if err != nil || parsedPermission < 0 || parsedPermission > 511 {
		return errors.New("Configuration error: File permission must constitute a valid unix value for file permissions")
	}

	parsedPermission, err = strconv.ParseInt(c.Filesystem.DirectoriesPermission, 8, 32)
	if err != nil || parsedPermission < 0 || parsedPermission > 511 {
		return errors.New("Configuration error: Directory permission must constitute a valid unix value for file permissions")
	}

	if c.Git.Repo == "" {
		return errors.New("Configuration error: Git repo cannot be empty")
	}

	if c.Git.Ref == "" {
		return errors.New("Configuration error: Git reference cannot be empty")
	}

	return nil
}

func setGrpcEndpointsRegex(c *Config) error {
	for idx, notif := range c.GrpcNotifications {
		if notif.Filter != "" {
			exp, expErr := regexp.Compile(notif.Filter)
			if expErr != nil {
				return expErr
			}
			notif.FilterRegex = exp
			c.GrpcNotifications[idx] = notif
		}
	}

	return nil
}

func expandPath(fpath string, homedir string) string {
	if strings.HasPrefix(fpath, "~/") {
		fpath = path.Join(homedir, fpath[2:])
	}

	return fpath
}

func GetConfig(confFilePath string) (Config, error) {
	var c Config

	bs, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return Config{}, errors.New(fmt.Sprintf("Error reading configuration file: %s", err.Error()))
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return Config{}, errors.New(fmt.Sprintf("Error reading configuration file: %s", err.Error()))
	}

	if c.Filesystem.FilesPermission == "" {
		c.Filesystem.FilesPermission = "0550"
	}

	if c.Filesystem.DirectoriesPermission == "" {
		c.Filesystem.DirectoriesPermission = "0550"
	}

	absPath, absPathErr := filepath.Abs(c.Filesystem.Path)
	if absPathErr != nil {
		return Config{}, errors.New(fmt.Sprintf("Error conversion filesystem path to absolute path: %s", absPathErr.Error()))
	}

	c.Filesystem.Path = absPath

	expErr := setGrpcEndpointsRegex(&c)
	if expErr != nil {
		return c, expErr
	}

	err = checkConfigIntegrity(c)
	if err != nil {
		return Config{}, err
	}

	homeDir, homeDirErr := os.UserHomeDir()
	if homeDirErr == nil {
		c.Git.Auth.SshKey = expandPath(c.Git.Auth.SshKey, homeDir)
		c.Git.Auth.KnownKey = expandPath(c.Git.Auth.KnownKey, homeDir)
	}

	return c, nil
}