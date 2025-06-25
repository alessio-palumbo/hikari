package decode

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

var embedRegexp = regexp.MustCompile(`^(?:\[(\d+)\])?<(.+)>$`)

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

func DecodeProtocol(yamlBytes []byte) (*ProtocolSpec, error) {
	var rootNode yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &rootNode); err != nil {
		return nil, fmt.Errorf("failed to parse protocol YAML: %w", err)
	}

	// Find top-level mapping keys
	if rootNode.Kind != yaml.DocumentNode || len(rootNode.Content) == 0 || rootNode.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("unexpected YAML format: expected document with mapping node")
	}
	top := rootNode.Content[0]

	var enumsNode, fieldsNode, unionsNode, packetsNode *yaml.Node
	for i := 0; i < len(top.Content); i += 2 {
		key := top.Content[i].Value
		val := top.Content[i+1]
		switch key {
		case "enums":
			enumsNode = val
		case "fields":
			fieldsNode = val
		case "unions":
			unionsNode = val
		case "packets":
			packetsNode = val
		}
	}

	enums, err := decodeNamedMap[*Enum](enumsNode)
	if err != nil {
		return nil, err
	}

	fieldGroups, err := decodeNamedMap[*FieldGroup](fieldsNode)
	if err != nil {
		return nil, err
	}

	unions, err := decodeNamedMap[*Union](unionsNode)
	if err != nil {
		return nil, err
	}

	packets, err := decodePacketGroups(packetsNode)
	if err != nil {
		return nil, err
	}

	enumMap := make(map[string]*Enum)
	for _, e := range enums {
		enumMap[e.Name] = e
	}
	fgMap := make(map[string]*FieldGroup)
	for _, fg := range fieldGroups {
		fgMap[fg.Name] = fg
	}
	unionMap := make(map[string]*Union)
	for _, u := range unions {
		unionMap[u.Name] = u
	}
	pktMap := make(map[string]*Packet)
	for _, p := range packets {
		pktMap[p.Name] = p
	}

	// Resolve embedded types in field groups and packets.
	for _, fg := range fieldGroups {
		resolved, err := resolveEmbeds("fields", fg.Fields, fgMap, unionMap, pktMap, enumMap)
		if err != nil {
			return nil, err
		}
		fg.Fields = resolved
	}

	for _, u := range unions {
		resolved, err := resolveEmbeds("fields", u.Fields, fgMap, unionMap, pktMap, enumMap)
		if err != nil {
			return nil, err
		}
		u.Fields = resolved
	}

	for _, pkt := range packets {
		resolved, err := resolveEmbeds("packets", pkt.Fields, fgMap, unionMap, pktMap, enumMap)
		if err != nil {
			return nil, err
		}
		pkt.Fields = resolved
	}

	return &ProtocolSpec{
		Enums:   derefSlice(enums),
		Fields:  derefSlice(fieldGroups),
		Unions:  derefSlice(unions),
		Packets: derefSlice(packets),
	}, nil
}

// decodeNamedMap decodes a YAML mapping node into a slice of named items.
func decodeNamedMap[T named](node *yaml.Node) ([]T, error) {
	if node == nil {
		return nil, fmt.Errorf("node is nil")
	}
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping node %v", node.Kind)
	}

	var items []T
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]

		item := new(T)
		if err := valNode.Decode(item); err != nil {
			return nil, fmt.Errorf("decoding map value for key %q: %w", keyNode.Value, err)
		}

		(*item).SetName(keyNode.Value)
		items = append(items, *item)
	}
	return items, nil
}

// decodePacketGroups decodes a YAML mapping node into a slice of Packets, each with its namespace.
func decodePacketGroups(node *yaml.Node) ([]*Packet, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping node for packet groups")
	}

	var packets []*Packet

	for i := 0; i < len(node.Content); i += 2 {
		nsNode := node.Content[i]
		nsValue := node.Content[i+1]
		namespace := nsNode.Value

		namedPkts, err := decodeNamedMap[*Packet](nsValue)
		if err != nil {
			return nil, fmt.Errorf("decoding packets in namespace %q: %w", namespace, err)
		}

		for _, pkt := range namedPkts {
			pkt.Namespace = namespace
			packets = append(packets, pkt)
		}
	}

	return packets, nil
}

func derefSlice[T any](ptrs []*T) []T {
	out := make([]T, len(ptrs))
	for i, v := range ptrs {
		out[i] = *v
	}
	return out
}

func resolveEmbeds(context string, fields []Field, fgMap map[string]*FieldGroup, unionMap map[string]*Union, pktMap map[string]*Packet, enumMap map[string]*Enum) ([]Field, error) {
	var resolved []Field

	for _, f := range fields {
		typ, err := resolveFieldType(context, f.Type, fgMap, unionMap, pktMap, enumMap)
		if err != nil {
			return nil, err
		}
		f.Type = typ
		resolved = append(resolved, f)
	}

	return resolved, nil
}

func resolveFieldType(context string, typeStr string, fgMap map[string]*FieldGroup, unionMap map[string]*Union, pktMap map[string]*Packet, enumMap map[string]*Enum) (string, error) {
	arrayPrefix, embed, ok := parseEmbedFields(typeStr)
	if !ok {
		return typeStr, nil
	}

	var ref string

	switch {
	case fgMap[embed] != nil, unionMap[embed] != nil, pktMap[embed] != nil:
	case enumMap[embed] != nil:
		ref = "enums."
	default:
		return "", fmt.Errorf("unknown embedded reference <%s>", embed)
	}

	return fmt.Sprintf("%s%s%s", arrayPrefix, ref, embed), nil
}

func parseEmbedFields(s string) (arrayPrefix string, embeddedType string, ok bool) {
	if m := embedRegexp.FindStringSubmatch(s); m != nil {
		if m[1] != "" {
			arrayPrefix = "[" + m[1] + "]"
		}
		return arrayPrefix, m[2], true
	}
	return "", "", false
}
