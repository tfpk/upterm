package utils

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"regexp"
	"errors"

	"github.com/dchest/uniuri"
	"golang.org/x/crypto/ssh"
)

const (
	logFile = "upterm.log"
)

func UptermDir() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homedir, ".upterm"), nil
}

func CreateUptermDir() (string, error) {
	dir, err := UptermDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return dir, nil
}

func OpenHostLogFile() (*os.File, error) {
	dir, err := CreateUptermDir()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(dir, logFile), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
}

func DefaultLocalhost(defaultPort string) string {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	return fmt.Sprintf("127.0.0.1:%s", port)
}

func CreateSigners(privateKeys [][]byte) ([]ssh.Signer, error) {
	var signers []ssh.Signer

	for _, pk := range privateKeys {
		signer, err := ssh.ParsePrivateKey(pk)
		if err != nil {
			return nil, err
		}

		signers = append(signers, signer)
	}

	// generate one if no signer
	if len(signers) == 0 {
		_, epk, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, err
		}

		signer, err := ssh.NewSignerFromKey(epk)
		if err != nil {
			return nil, err
		}

		signers = append(signers, signer)

	}

	return signers, nil
}

func ReadFiles(paths []string) ([][]byte, error) {
	var files [][]byte

	for _, p := range paths {
		b, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", p, err)
		}

		files = append(files, b)
	}

	return files, nil
}

func GenerateSessionID(sessionName string) (string, error) {
	var serverNameRegexp = regexp.MustCompile("^[a-zA-Z0-9_-]{3,}$")

	// Empty SessionName == generated randomly
	if sessionName == "" {
		return uniuri.NewLen(uniuri.UUIDLen), nil
	}

	if serverNameRegexp.MatchString(sessionName) {
		return sessionName, nil
	}

	return "", errors.New("Could not generate session ID from the given code.")
}

func FingerprintSHA256(key ssh.PublicKey) string {
	hash := sha256.Sum256(key.Marshal())
	b64hash := base64.StdEncoding.EncodeToString(hash[:])
	return fmt.Sprintf("SHA256:%s", strings.TrimRight(b64hash, "="))
}
