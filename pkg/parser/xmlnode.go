package parser

import (
	"encoding/xml"
	"strconv"
	"strings"
)

// xmlNode is a generic XML node for flexible parsing.
type xmlNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  string     `xml:",chardata"`
	Children []xmlNode  `xml:",any"`
}

// AttrVal returns the value of an attribute by local name.
func (n *xmlNode) AttrVal(ns, local string) string {
	for _, a := range n.Attrs {
		if (ns == "" || a.Name.Space == ns) && a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

// AttrInt returns the integer value of an attribute.
func (n *xmlNode) AttrInt(local string) int {
	val := n.AttrVal("", local)
	if val == "" {
		return 0
	}
	v, _ := strconv.Atoi(val)
	return v
}

// HasChild checks if a child with the given local name exists.
func (n *xmlNode) HasChild(local string) bool {
	return n.FindChild(local) != nil
}

// FindChild finds the first child with the given local name.
func (n *xmlNode) FindChild(local string) *xmlNode {
	for i := range n.Children {
		if n.Children[i].XMLName.Local == local {
			return &n.Children[i]
		}
	}
	return nil
}

// ChildInt finds a child element by its own local name and returns its int attribute.
func (n *xmlNode) ChildInt(childName, attrName string) int {
	child := n.FindChild(childName)
	if child == nil {
		return 0
	}
	return child.AttrInt(attrName)
}

// FindChildren finds all children with the given local name.
func (n *xmlNode) FindChildren(local string) []xmlNode {
	var out []xmlNode
	for i := range n.Children {
		if n.Children[i].XMLName.Local == local {
			out = append(out, n.Children[i])
		}
	}
	return out
}

// UnmarshalXML implements custom XML unmarshalling.
func (n *xmlNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.XMLName = start.Name
	n.Attrs = start.Attr
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}
		switch t := token.(type) {
		case xml.StartElement:
			var child xmlNode
			if err := child.UnmarshalXML(d, t); err != nil {
				return err
			}
			n.Children = append(n.Children, child)
		case xml.EndElement:
			if t.Name == start.Name {
				// Trim whitespace from content
				n.Content = strings.TrimSpace(n.Content)
				return nil
			}
		case xml.CharData:
			n.Content += string(t)
		}
	}
}
