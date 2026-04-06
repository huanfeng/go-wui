package editor

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// LoadXML reads an XML layout file and returns an EditorNode tree.
func LoadXML(path string) (*EditorNode, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseXML(f)
}

// ParseXML parses XML from a reader into an EditorNode tree.
func ParseXML(r io.Reader) (*EditorNode, error) {
	decoder := xml.NewDecoder(r)
	var stack []*EditorNode
	var root *EditorNode

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			node := NewEditorNode(t.Name.Local)
			for _, attr := range t.Attr {
				node.SetAttr(attr.Name.Local, attr.Value)
			}
			if len(stack) > 0 {
				stack[len(stack)-1].AddChild(node)
			}
			stack = append(stack, node)
			if root == nil {
				root = node
			}

		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}

	if root == nil {
		return nil, fmt.Errorf("empty XML document")
	}
	return root, nil
}

// SaveXML writes the EditorNode tree to an XML file.
func SaveXML(path string, root *EditorNode) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, `<?xml version="1.0" encoding="utf-8"?>`)
	if err != nil {
		return err
	}
	return writeNode(f, root, 0)
}

// MarshalXML returns the XML string representation of the node tree.
func MarshalXML(root *EditorNode) string {
	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	writeNodeToBuilder(&sb, root, 0)
	return sb.String()
}

func writeNode(w io.Writer, node *EditorNode, depth int) error {
	indent := strings.Repeat("    ", depth)

	if len(node.Children) == 0 && len(node.AttrKeys()) <= 3 {
		// Self-closing short form for leaf nodes with few attributes.
		fmt.Fprintf(w, "%s<%s", indent, node.Tag)
		writeAttrsInline(w, node)
		fmt.Fprintln(w, " />")
		return nil
	}

	// Opening tag with attributes.
	fmt.Fprintf(w, "%s<%s\n", indent, node.Tag)
	writeAttrsMultiline(w, node, depth)
	if len(node.Children) == 0 {
		fmt.Fprintf(w, "%s/>\n", indent+strings.Repeat(" ", len(node.Tag)+1))
		return nil
	}
	fmt.Fprintf(w, "%s>\n", indent+strings.Repeat(" ", len(node.Tag)+1))

	for _, child := range node.Children {
		if err := writeNode(w, child, depth+1); err != nil {
			return err
		}
	}

	fmt.Fprintf(w, "%s</%s>\n", indent, node.Tag)
	return nil
}

func writeNodeToBuilder(sb *strings.Builder, node *EditorNode, depth int) {
	indent := strings.Repeat("    ", depth)

	if len(node.Children) == 0 && len(node.AttrKeys()) <= 3 {
		fmt.Fprintf(sb, "%s<%s", indent, node.Tag)
		writeAttrsInlineBuilder(sb, node)
		sb.WriteString(" />\n")
		return
	}

	fmt.Fprintf(sb, "%s<%s\n", indent, node.Tag)
	writeAttrsMultilineBuilder(sb, node, depth)
	if len(node.Children) == 0 {
		fmt.Fprintf(sb, "%s/>\n", indent+strings.Repeat(" ", len(node.Tag)+1))
		return
	}
	fmt.Fprintf(sb, "%s>\n", indent+strings.Repeat(" ", len(node.Tag)+1))

	for _, child := range node.Children {
		writeNodeToBuilder(sb, child, depth+1)
	}
	fmt.Fprintf(sb, "%s</%s>\n", indent, node.Tag)
}

func writeAttrsInline(w io.Writer, node *EditorNode) {
	for _, key := range node.AttrKeys() {
		fmt.Fprintf(w, " %s=%q", key, node.Attrs[key])
	}
}

func writeAttrsInlineBuilder(sb *strings.Builder, node *EditorNode) {
	for _, key := range node.AttrKeys() {
		fmt.Fprintf(sb, " %s=%q", key, node.Attrs[key])
	}
}

func writeAttrsMultiline(w io.Writer, node *EditorNode, depth int) {
	attrIndent := strings.Repeat("    ", depth) + strings.Repeat(" ", len(node.Tag)+2)
	for _, key := range node.AttrKeys() {
		fmt.Fprintf(w, "%s%s=%q\n", attrIndent, key, node.Attrs[key])
	}
}

func writeAttrsMultilineBuilder(sb *strings.Builder, node *EditorNode, depth int) {
	attrIndent := strings.Repeat("    ", depth) + strings.Repeat(" ", len(node.Tag)+2)
	for _, key := range node.AttrKeys() {
		fmt.Fprintf(sb, "%s%s=%q\n", attrIndent, key, node.Attrs[key])
	}
}
