package slack

import (
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
)

type (
	blockConverter struct {
		prefix string

		sections   []slack.RichTextElement
		sectionEls []slack.RichTextSectionElement

		listMarkers []nodeListMarker
	}

	nodeListMarker struct {
		depth    int
		listType slack.RichTextListElementType
		start    int
		end      int
	}
)

func plainTextBlock(text string) *slack.TextBlockObject {
	return slack.NewTextBlockObject("plain_text", text, false, false)
}

func convertContentToBlocks(content *rez.ContentNode, prefix string) []slack.Block {
	c := &blockConverter{}
	if prefix != "" {
		c.prefix = prefix + "_"
	}
	return c.convertContent(content)
}

func (c *blockConverter) convertContent(doc *rez.ContentNode) []slack.Block {
	if doc == nil {
		return nil
	}

	c.crawlNode(doc, mapset.NewSet[string](), 0)

	blocks := make([]slack.Block, len(c.sections))
	for i, s := range c.sections {
		blockId := fmt.Sprintf("%sblock_%d", c.prefix, i)
		blocks[i] = slack.NewRichTextBlock(blockId, s)
	}

	return blocks
}

func (c *blockConverter) crawlNode(node *rez.ContentNode, marks mapset.Set[string], depth int) {
	addedMarks := mapset.NewSet[string]()
	for _, m := range node.Marks {
		markName := string(m.Type.Name)
		if marks.Add(markName) {
			addedMarks.Add(markName)
		}
	}

	c.convertNode(node, marks)

	c.crawlChildren(node, marks, depth)

	if depth == 1 {
		c.flushSections()
	}

	// remove marks added by this node
	marks = marks.Difference(addedMarks)
}

func (c *blockConverter) convertNode(node *rez.ContentNode, marks mapset.Set[string]) {
	// TODO: support links etc, not just plain text
	style := &slack.RichTextSectionTextStyle{
		Bold:   marks.Contains("bold"),
		Italic: marks.Contains("italic"),
	}

	if node.Type.Name == "user" {
		c.sectionEls = append(c.sectionEls, slack.NewRichTextSectionUserElement(node.Text, style))
	}

	if node.IsText() {
		if href, isLink := node.Attrs["href"]; isLink {
			if hrefStr := href.(string); hrefStr != "" {
				c.sectionEls = append(c.sectionEls, slack.NewRichTextSectionLinkElement(hrefStr, node.Text, style))
				return
			}
		}

		c.sectionEls = append(c.sectionEls, slack.NewRichTextSectionTextElement(node.Text, style))
	}
}

func (c *blockConverter) crawlChildren(node *rez.ContentNode, marks mapset.Set[string], depth int) {
	nodeTypeName := string(node.Type.Name)
	isList := nodeTypeName == "bulletList" || nodeTypeName == "orderedList" // todo: check if 'container'
	var listIdx int
	if isList {
		listIdx = len(c.listMarkers)
		marker := nodeListMarker{
			depth:    depth,
			start:    len(c.sectionEls),
			end:      -1,
			listType: slack.RTEListOrdered,
		}
		if nodeTypeName == "bulletList" {
			marker.listType = slack.RTEListBullet
		}
		c.listMarkers = append(c.listMarkers, marker)
	}

	for _, child := range node.Content.Content {
		c.crawlNode(&child, marks, depth+1)
	}

	if isList {
		c.listMarkers[listIdx].end = len(c.sectionEls)
	}
}

func (c *blockConverter) flushSections() {
	wasInList, rootEls := c.flushLists()

	if len(rootEls) == 0 && !wasInList {
		rootEls = []slack.RichTextSectionElement{
			slack.NewRichTextSectionTextElement("N/A", nil),
		}
	}

	if len(rootEls) > 0 {
		c.sections = append(c.sections, slack.NewRichTextSection(rootEls...))
	}
	c.sectionEls = make([]slack.RichTextSectionElement, 0)
}

func (c *blockConverter) flushLists() (bool, []slack.RichTextSectionElement) {
	var wasInList bool
	var rootEls []slack.RichTextSectionElement
	var listEls []slack.RichTextSectionElement
	lastIndent := 0
	for i := range len(c.sectionEls) {
		el := c.sectionEls[i]
		deepestIndent := -1
		for indent, split := range c.listMarkers {
			if i >= split.start && i < split.end && indent > deepestIndent {
				deepestIndent = indent
			}
		}
		if deepestIndent == -1 {
			rootEls = append(rootEls, el)
			continue
		}
		wasInList = true
		if lastIndent != deepestIndent && len(listEls) > 0 {
			marker := c.listMarkers[deepestIndent]
			section := slack.NewRichTextSection(listEls...)
			c.sections = append(c.sections, slack.NewRichTextList(marker.listType, lastIndent, section))
			listEls = []slack.RichTextSectionElement{}
		}
		listEls = append(listEls, el)
		lastIndent = deepestIndent
	}
	if len(listEls) > 0 {
		marker := c.listMarkers[lastIndent]
		section := slack.NewRichTextSection(listEls[0])
		c.sections = append(c.sections, slack.NewRichTextList(marker.listType, lastIndent, section))
	}
	c.listMarkers = make([]nodeListMarker, 0)
	return wasInList, rootEls
}
