package types

type EnumValue struct {
	Name  string `yaml:"name"`
	Value int    `yaml:"value"`
}

type Enum struct {
	Type   string      `yaml:"type"`
	Values []EnumValue `yaml:"values"`
}

type Field struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	SizeBytes int    `yaml:"size_bytes"`
}

type FieldGroup struct {
	SizeBytes int     `yaml:"size_bytes"`
	Fields    []Field `yaml:"fields"`
}

type Packet struct {
	PktType   int     `yaml:"pkt_type"`
	SizeBytes int     `yaml:"size_bytes"`
	Fields    []Field `yaml:"fields"`
}

type ProtocolSpec struct {
	Enums   map[string]Enum       `yaml:"enums"`
	Fields  map[string]FieldGroup `yaml:"fields"`
	Packets map[string]Packet     `yaml:"packets"`
}
