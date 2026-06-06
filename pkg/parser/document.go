package parser

import (
	"encoding/xml"
	"fmt"

	"github.com/saiful-anwar/word2md/pkg/models"
)

// DocumentReader parses document.xml into a model Document.
type DocumentReader struct{}

// NewDocumentReader creates a DocumentReader.
func NewDocumentReader() *DocumentReader {
	return &DocumentReader{}
}

// Read parses the document XML and returns a model Document.
func (dr *DocumentReader) Read(data []byte) (*models.Document, error) {
	var raw rawDocument
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal document.xml: %w", err)
	}

	doc := &models.Document{}
	for _, bodyChild := range raw.Body.Children {
		block, err := dr.parseBlock(bodyChild)
		if err != nil {
			return nil, err
		}
		if block != nil {
			doc.Body = append(doc.Body, *block)
		}
	}

	return doc, nil
}

func (dr *DocumentReader) parseBlock(child xmlNode) (*models.Block, error) {
	switch child.XMLName.Local {
	case "p":
		para, err := dr.parseParagraph(child)
		if err != nil {
			return nil, err
		}
		return &models.Block{Kind: models.KindParagraph, Value: para}, nil
	case "tbl":
		tbl, err := dr.parseTable(child)
		if err != nil {
			return nil, err
		}
		return &models.Block{Kind: models.KindTable, Value: tbl}, nil
	default:
		return nil, nil
	}
}

func (dr *DocumentReader) parseParagraph(node xmlNode) (*models.Paragraph, error) {
	para := &models.Paragraph{}

	for _, child := range node.Children {
		switch child.XMLName.Local {
		case "pPr":
			if pStyleNode := child.FindChild("pStyle"); pStyleNode != nil {
				para.StyleID = pStyleNode.AttrVal("", "val")
			}
			if jcNode := child.FindChild("jc"); jcNode != nil {
				para.Alignment = jcNode.AttrVal("", "val")
			}
			if numPrNode := child.FindChild("numPr"); numPrNode != nil {
				numID := numPrNode.ChildInt("numId", "val")
				level := numPrNode.ChildInt("ilvl", "val")
				para.NumPr = &models.NumberingProperties{
					NumID: numID,
					Ilvl:  level,
				}
			}
			if indentNode := child.FindChild("ind"); indentNode != nil {
				para.Indent = &models.ParagraphIndent{
					Left:    indentNode.AttrInt("left"),
					Right:   indentNode.AttrInt("right"),
					First:   indentNode.AttrInt("firstLine"),
					Hanging: indentNode.AttrInt("hanging"),
				}
			}
		case "r":
			run, err := dr.parseRun(child)
			if err != nil {
				return nil, err
			}
			para.Runs = append(para.Runs, *run)
		case "hyperlink":
			for _, runNode := range child.FindChildren("r") {
				run, err := dr.parseRun(runNode)
				if err != nil {
					return nil, err
				}
				linkID := child.AttrVal("http://schemas.openxmlformats.org/officeDocument/2006/relationships", "id")
				if linkID == "" {
					linkID = child.AttrVal("", "id")
				}
				if linkID != "" {
					run.Hyperlink = &models.Hyperlink{URL: linkID} // resolved later
				}
				para.Runs = append(para.Runs, *run)
			}
		}
	}

	return para, nil
}

// parseRunProperties extracts formatting properties from an rPr XML node into a TextRun.
func (dr *DocumentReader) parseRunProperties(rPr xmlNode, run *models.TextRun) {
	if bNode := rPr.FindChild("b"); bNode != nil {
		val := bNode.AttrVal("", "val")
		run.Bold = val != "0" && val != "false"
	}
	if iNode := rPr.FindChild("i"); iNode != nil {
		val := iNode.AttrVal("", "val")
		run.Italic = val != "0" && val != "false"
	}
	if strikeNode := rPr.FindChild("strike"); strikeNode != nil {
		val := strikeNode.AttrVal("", "val")
		run.Strikethrough = val != "0" && val != "false"
	}
	if szNode := rPr.FindChild("sz"); szNode != nil {
		run.FontSize = szNode.AttrInt("val")
	}
	if colorNode := rPr.FindChild("color"); colorNode != nil {
		run.Color = colorNode.AttrVal("", "val")
	}
	if rFonts := rPr.FindChild("rFonts"); rFonts != nil {
		run.FontName = rFonts.AttrVal("", "ascii")
	}
	// Detect code style by font name
	if run.FontName == "Courier New" || run.FontName == "Consolas" || run.FontName == "Monospace" {
		run.Code = true
	}
}

func (dr *DocumentReader) parseRun(node xmlNode) (*models.TextRun, error) {
	run := &models.TextRun{}

	for _, child := range node.Children {
		switch child.XMLName.Local {
		case "rPr":
			dr.parseRunProperties(child, run)
		case "t":
			run.Text += child.Content
		case "br":
			run.Text += "\n"
		case "tab":
			run.Text += "\t"
		case "drawing":
			// Handle inline images (blip embed)
			if embed := extractDrawingEmbed(child); embed != "" {
				run.DrawingRID = embed
			}
		}
	}

	return run, nil
}

func (dr *DocumentReader) parseTable(node xmlNode) (*models.Table, error) {
	tbl := &models.Table{}

	for _, child := range node.Children {
		if child.XMLName.Local == "tr" {
			row, err := dr.parseTableRow(child)
			if err != nil {
				return nil, err
			}
			tbl.Rows = append(tbl.Rows, *row)
		}
	}

	return tbl, nil
}

func (dr *DocumentReader) parseTableRow(node xmlNode) (*models.TableRow, error) {
	row := &models.TableRow{}

	for _, child := range node.Children {
		if child.XMLName.Local == "tc" {
			cell, err := dr.parseTableCell(child)
			if err != nil {
				return nil, err
			}
			row.Cells = append(row.Cells, *cell)
		}
	}

	return row, nil
}

func (dr *DocumentReader) parseTableCell(node xmlNode) (*models.TableCell, error) {
	cell := &models.TableCell{}

	for _, child := range node.Children {
		block, err := dr.parseBlock(child)
		if err != nil {
			return nil, err
		}
		if block != nil {
			cell.Blocks = append(cell.Blocks, *block)
		}
	}

	return cell, nil
}

// extractDrawingEmbed extracts the rId from a drawing element.
func extractDrawingEmbed(node xmlNode) string {
	// Navigate: inline > graphic > graphicData > pic > blipFill > blip
	inline := node.FindChild("inline")
	if inline == nil {
		inline = node.FindChild("anchor")
		if inline == nil {
			return ""
		}
	}
	graphic := inline.FindChild("graphic")
	if graphic == nil {
		return ""
	}
	graphicData := graphic.FindChild("graphicData")
	if graphicData == nil {
		return ""
	}
	pic := graphicData.FindChild("pic")
	if pic == nil {
		return ""
	}
	blipFill := pic.FindChild("blipFill")
	if blipFill == nil {
		return ""
	}
	blip := blipFill.FindChild("blip")
	if blip == nil {
		return ""
	}
	return blip.AttrVal("http://schemas.openxmlformats.org/officeDocument/2006/relationships", "embed")
}

// raw types for XML unmarshalling

type rawDocument struct {
	XMLName xml.Name `xml:"document"`
	Body    rawBody  `xml:"body"`
}

type rawBody struct {
	Children []xmlNode `xml:",any"`
}
