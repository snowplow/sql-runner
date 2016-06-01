package playbook

type YAMLFilePlaybookProvider struct {
	playbookPath string
}

func NewYAMLFilePlaybookProvider(playbookPath string) *YAMLFilePlaybookProvider {
	return &YAMLFilePlaybookProvider{
		playbookPath: playbookPath,
	}
}

func (p YAMLFilePlaybookProvider) GetPlaybook() (*Playbook, error) {
	lines, err := loadLocalFile(p.playbookPath)
	if err != nil {
		return nil, err
	}

	playbook, pbErr := parsePlaybookYaml(lines)
	return &playbook, pbErr
}
