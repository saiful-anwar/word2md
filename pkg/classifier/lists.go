package classifier

import (
	"strings"

	"github.com/saiful-anwar/word2md/pkg/models"
)

// ListDetector identifies list items and groups consecutive list items.
type ListDetector struct{}

// NewListDetector creates a ListDetector.
func NewListDetector() *ListDetector {
	return &ListDetector{}
}

// DetectListItems scans document blocks and groups consecutive list items.
// Returns a new slice of blocks where list items are grouped into List blocks.
func (ld *ListDetector) DetectListItems(blocks []models.Block) []models.Block {
	var result []models.Block
	var currentList *models.List

	for i := range blocks {
		if blocks[i].Kind == models.KindParagraph {
			para := blocks[i].Value.(*models.Paragraph)
			if para.NumPr != nil {
				if currentList == nil {
					currentList = &models.List{
						NumID:    para.NumPr.NumID,
						IsBullet: para.NumPr.IsBullet,
					}
				}
				currentList.Items = append(currentList.Items, *para)
				continue
			}
		}

		// If we hit a non-list item, flush current list
		if currentList != nil {
			result = append(result, models.Block{Kind: models.KindList, Value: currentList})
			currentList = nil
		}
		result = append(result, blocks[i])
	}

	// Flush remaining list
	if currentList != nil {
		result = append(result, models.Block{Kind: models.KindList, Value: currentList})
	}

	return result
}

// ApplyStyleListInfo determines if a NumPr represents a bullet or numbered list.
// This requires reading numbering.xml; for now, we assume NumPr implies numbered
// and bullet lists are detected by bullet characters.
func (ld *ListDetector) ApplyStyleListInfo(blocks []models.Block) {
	for i := range blocks {
		if blocks[i].Kind == models.KindList {
			list := blocks[i].Value.(*models.List)
			// Check if first item has bullet-like prefix
			if len(list.Items) > 0 {
				text := extractParagraphText(&list.Items[0])
				if hasBulletPrefix(text) {
					list.IsBullet = true
				}
			}
		}
	}
}

func extractParagraphText(para *models.Paragraph) string {
	var sb strings.Builder
	for _, run := range para.Runs {
		sb.WriteString(run.Text)
	}
	return sb.String()
}

func hasBulletPrefix(text string) bool {
	return strings.HasPrefix(text, "•") ||
		strings.HasPrefix(text, "-") ||
		strings.HasPrefix(text, "*") ||
		strings.HasPrefix(text, "○")
}
