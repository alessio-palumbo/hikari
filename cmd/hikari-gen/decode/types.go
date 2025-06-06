package decode

type ProtocolSpec struct {
	Enums   []Enum       `yaml:"enums"`
	Fields  []FieldGroup `yaml:"fields"`
	Unions  []Union      `yaml:"unions"`
	Packets []Packet     `yaml:"packets"`
}

type Enum struct {
	Name   string      `yaml:"-"`
	Type   string      `yaml:"type"`
	Values []EnumValue `yaml:"values"`
}

type EnumValue struct {
	Name  string `yaml:"name"`
	Value int    `yaml:"value"`
}

type FieldGroup struct {
	Name      string  `yaml:"-"`
	SizeBytes int     `yaml:"size_bytes"`
	Fields    []Field `yaml:"fields"`
}

type Union struct {
	Name      string  `yaml:"-"`
	SizeBytes int     `yaml:"size_bytes"`
	Fields    []Field `yaml:"fields"`
}

type Packet struct {
	Name      string  `yaml:"-"`
	Namespace string  `yaml:"-"`
	PktType   int     `yaml:"pkt_type"`
	SizeBytes int     `yaml:"size_bytes"`
	Fields    []Field `yaml:"fields"`
}

type Field struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	SizeBytes int    `yaml:"size_bytes"`
}

type named interface {
	SetName(string)
}

func (p *Enum) SetName(name string) {
	p.Name = name
}

func (p *FieldGroup) SetName(name string) {
	p.Name = name
}

func (p *Union) SetName(name string) {
	p.Name = name
}

func (p *Packet) SetName(name string) {
	p.Name = name
}

func (p *Packet) SetNamespace(ns string) {
	p.Namespace = ns
}
