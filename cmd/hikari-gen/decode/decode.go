package decode

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

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

	var enumsNode, fieldsNode, packetsNode *yaml.Node
	for i := 0; i < len(top.Content); i += 2 {
		key := top.Content[i].Value
		val := top.Content[i+1]
		switch key {
		case "enums":
			enumsNode = val
		case "fields":
			fieldsNode = val
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

	packets, err := decodePacketGroups(packetsNode)
	if err != nil {
		return nil, err
	}

	return &ProtocolSpec{
		Enums:   derefSlice(enums),
		Fields:  derefSlice(fieldGroups),
		Packets: packets,
	}, nil
}

func decodeNamedMap[T named](node *yaml.Node) ([]T, error) {
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

func decodePacketGroups(node *yaml.Node) ([]Packet, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping node for packet groups")
	}

	var packets []Packet

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
			packets = append(packets, *pkt)
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
