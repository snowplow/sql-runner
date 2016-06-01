package playbook

type ConsulPlaybookProvider struct {
	consulAddress string
	consulKey     string
}

func NewConsulPlaybookProvider(consulAddress, consulKey string) *ConsulPlaybookProvider {
	return &ConsulPlaybookProvider{
		consulAddress: consulAddress,
		consulKey:     consulKey,
	}
}

func (p ConsulPlaybookProvider) GetPlaybook() (*Playbook, error) {
	lines, err := GetBytesFromConsul(p.consulAddress, p.consulKey)
	if err != nil {
		return nil, err
	}

	playbook, pbErr := parsePlaybookYaml(lines)
	return &playbook, pbErr
}
