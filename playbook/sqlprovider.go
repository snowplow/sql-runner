package playbook

import (
	"io/ioutil"
	"path"
)

type SQLProvider interface {
	ResolveKey(key string) string
	GetSQL(key string) (string, error)
}

type FileSQLProvider struct {
	rootPath string
}

func NewFileSQLProvider(rootPath string) *FileSQLProvider {
	return &FileSQLProvider{
		rootPath: rootPath,
	}
}

func (p FileSQLProvider) GetSQL(scriptPath string) (string, error) {
	return readScript(p.ResolveKey(scriptPath))
}

func (p FileSQLProvider) ResolveKey(scriptPath string) string {
	return path.Join(p.rootPath, scriptPath)
}

// Reads the file ready for executing
func readScript(file string) (string, error) {
	scriptBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(scriptBytes), nil
}

type ConsulSQLProvider struct {
	consulAddress string
	keyPrefix     string
}

func NewConsulSQLProvider(consulAddress, keyPrefix string) *ConsulSQLProvider {
	return &ConsulSQLProvider{
		consulAddress: consulAddress,
		keyPrefix:     keyPrefix,
	}
}

func (p ConsulSQLProvider) GetSQL(key string) (string, error) {
	return GetStringValueFromConsul(p.consulAddress, p.ResolveKey(key))
}

func (p ConsulSQLProvider) ResolveKey(key string) string {
	return path.Join(p.keyPrefix, key)
}
