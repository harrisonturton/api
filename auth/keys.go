package auth

import (
	"bytes"
	"fmt"
	"os/exec"
)

const (
	cmdGenPrivateKey = "openssl genrsa -out %s 2048"
	cmdGenPublicKey  = "openssl rsa -in %s -pubout > %s"
)

func GenerateKeys(privateKeyPath, publicKeyPath string) error {
	err := genPrivateKey(privateKeyPath)
	if err != nil {
		return err
	}
	err = genPublicKey(privateKeyPath, publicKeyPath)
	return err
}

func genPublicKey(privateKeyPath, publicKeyPath string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(cmdGenPublicKey, privateKeyPath, publicKeyPath))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate public key: %v\nStderr: %s", err, stderr.String())
	}
	return nil
}

func genPrivateKey(privateKeyPath string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(cmdGenPrivateKey, privateKeyPath))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v\nStderr: %s", err, stderr.String())
	}
	return nil
}
