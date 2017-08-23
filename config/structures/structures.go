package structures

type DbConfig struct {
	Driver struct {
		Host   string `yaml:"host,omitempty"`
		User   string `yaml:"user,omitempty"`
		Pw     string `yaml:"pw,omitempty"`
		Port   int    `yaml:"port,omitempty"`
		Schema string `yaml:"schema,omitempty"`
	}
}

type CorreiosConfig struct {
	CartaoPostagem    string `yaml:"cartaoPostagem,omitempty"`
	CodAdministrativo string `yaml:"codAdministrativo,omitempty"`
	Contrato          string `yaml:"contrato,omitempty"`
	UserReverse       string `yaml:"userReversa,omitempty"`
	PwReverse         string `yaml:"pwReversa,omitempty"`
	UrlReverse        string `yaml:"urlReversa,omitempty"`
	MaxRetries        int64  `yaml:"maxRetries,omitempty"`
	UserTracking      string `yaml:"userTracking,omitempty"`
	PwTracking        string `yaml:"pwTracking,omitempty"`
	UrlTracking       string `yaml:"urlTracking,omitempty"`
}
