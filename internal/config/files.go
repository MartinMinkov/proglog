package config

import (
	"os"
	"path/filepath"
)

var (
	CAFile               = configFile("CERT_DIR", "ca.pem")
	ServerCertFile       = configFile("CERT_DIR", "server.pem")
	ServerKeyFile        = configFile("CERT_DIR", "server-key.pem")
	RootClientCertFile   = configFile("CERT_DIR", "root-client.pem")
	RootClientKeyFile    = configFile("CERT_DIR", "root-client-key.pem")
	NobodyClientCertFile = configFile("CERT_DIR", "nobody-client.pem")
	NobodyClientKeyFile  = configFile("CERT_DIR", "nobody-client-key.pem")
	ACLModelFile         = configFile("AUTH_DIR", "model.conf")
	ACLPolicyFile        = configFile("AUTH_DIR", "policy.csv")
)

func configFile(key, filename string) string {
	if dir := os.Getenv(key); dir != "" {
		return filepath.Join(dir, filename)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDir, ".proglog", filename)
}
