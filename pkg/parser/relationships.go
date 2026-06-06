package parser

import (
	"encoding/xml"
	"fmt"

	"github.com/saiful-anwar/word2md/pkg/models"
)

// RelationshipsReader parses .rels files.
type RelationshipsReader struct{}

// NewRelationshipsReader creates a RelationshipsReader.
func NewRelationshipsReader() *RelationshipsReader {
	return &RelationshipsReader{}
}

// Read parses .rels XML data into a Relationships model.
func (rr *RelationshipsReader) Read(data []byte) (*models.Relationships, error) {
	var raw rawRelationships
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal relationships: %w", err)
	}

	rels := &models.Relationships{Items: make(map[string]models.Relationship)}
	for _, r := range raw.Relationships {
		rels.Items[r.ID] = models.Relationship{
			ID:     r.ID,
			Type:   r.Type,
			Target: r.Target,
		}
	}

	return rels, nil
}

type rawRelationships struct {
	XMLName       xml.Name          `xml:"Relationships"`
	Relationships []rawRelationship `xml:"Relationship"`
}

type rawRelationship struct {
	ID     string `xml:"Id,attr"`
	Type   string `xml:"Type,attr"`
	Target string `xml:"Target,attr"`
}
