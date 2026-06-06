package parser

import (
	"encoding/xml"
	"fmt"

	"github.com/saiful-anwar/word2md/pkg/models"
)

// StylesReader parses styles.xml into a StyleSheet.
type StylesReader struct{}

// NewStylesReader creates a StylesReader.
func NewStylesReader() *StylesReader {
	return &StylesReader{}
}

// Read parses styles.xml data into a StyleSheet model.
func (sr *StylesReader) Read(data []byte) (*models.StyleSheet, error) {
	var raw rawStyles
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal styles.xml: %w", err)
	}

	sheet := &models.StyleSheet{Styles: make(map[string]models.Style)}
	for _, rs := range raw.Styles {
		style := models.Style{
			StyleID:  rs.StyleID,
			Name:     rs.Name.Val,
			Type:     rs.Type,
			FontSize: 0,
			BasedOn:  rs.BasedOn.Val,
			Next:     rs.Next.Val,
		}
		if rs.RPr != nil {
			style.Bold = rs.RPr.Bold != nil
			style.Italic = rs.RPr.Italic != nil
		}
		if rs.RPr != nil {
			if rs.RPr.Sz != nil {
				style.FontSize = rs.RPr.Sz.Val
			}
		}
		if rs.PPr != nil && rs.PPr.RPr != nil {
			if rs.PPr.RPr.Sz != nil {
				style.FontSize = rs.PPr.RPr.Sz.Val
			}
		}
		sheet.Styles[style.StyleID] = style
	}

	// Resolve inherited font sizes
	for id, style := range sheet.Styles {
		if style.FontSize == 0 && style.BasedOn != "" {
			if base, ok := sheet.Styles[style.BasedOn]; ok {
				style.FontSize = base.FontSize
				sheet.Styles[id] = style
			}
		}
	}

	return sheet, nil
}

type rawStyles struct {
	XMLName xml.Name   `xml:"styles"`
	Styles  []rawStyle `xml:"style"`
}

type rawStyle struct {
	StyleID string  `xml:"styleId,attr"`
	Type    string  `xml:"type,attr"`
	Name    valAttr `xml:"name"`
	BasedOn valAttr `xml:"basedOn"`
	Next    valAttr `xml:"next"`
	RPr     *rawRPr `xml:"rPr"`
	PPr     *rawPPr `xml:"pPr"`
}

type rawRPr struct {
	Sz     *intVal   `xml:"sz"`
	Bold   *struct{} `xml:"b"`
	Italic *struct{} `xml:"i"`
	Strike *struct{} `xml:"strike"`
}

type rawPPr struct {
	RPr *rawRPr `xml:"rPr"`
}

type valAttr struct {
	Val string `xml:"val,attr"`
}

type intVal struct {
	Val int `xml:"val,attr"`
}
