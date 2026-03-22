package res

import "encoding/xml"

// xmlResources is the root XML element for Android-compatible resource files.
type xmlResources struct {
	XMLName xml.Name    `xml:"resources"`
	Strings []xmlString `xml:"string"`
	Colors  []xmlColor  `xml:"color"`
	Dimens  []xmlDimen  `xml:"dimen"`
	Arrays  []xmlArray  `xml:"string-array"`
	Styles  []xmlStyle  `xml:"style"`
}

type xmlString struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type xmlColor struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type xmlDimen struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type xmlArray struct {
	Name  string    `xml:"name,attr"`
	Items []xmlItem `xml:"item"`
}

type xmlItem struct {
	Value string `xml:",chardata"`
}

type xmlStyle struct {
	Name   string        `xml:"name,attr"`
	Parent string        `xml:"parent,attr"`
	Items  []xmlStyleItem `xml:"item"`
}

type xmlStyleItem struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

// ParseStringsXML parses string resources from XML data.
func ParseStringsXML(data []byte) (map[string]string, error) {
	var res xmlResources
	if err := xml.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, s := range res.Strings {
		m[s.Name] = s.Value
	}
	return m, nil
}

// ParseColorsXML parses color resources from XML data.
func ParseColorsXML(data []byte) (map[string]string, error) {
	var res xmlResources
	if err := xml.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, c := range res.Colors {
		m[c.Name] = c.Value
	}
	return m, nil
}

// ParseDimensXML parses dimension resources from XML data.
func ParseDimensXML(data []byte) (map[string]string, error) {
	var res xmlResources
	if err := xml.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, d := range res.Dimens {
		m[d.Name] = d.Value
	}
	return m, nil
}

// ParseStringArraysXML parses string-array resources from XML data.
func ParseStringArraysXML(data []byte) (map[string][]string, error) {
	var res xmlResources
	if err := xml.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	m := make(map[string][]string)
	for _, a := range res.Arrays {
		var items []string
		for _, item := range a.Items {
			items = append(items, item.Value)
		}
		m[a.Name] = items
	}
	return m, nil
}
