package prosemirror

import (
	"fmt"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/texm/prosemirror-go"
)

var (
	textType     = prosemirror.NodeType{Text: true}
	sectionType  = prosemirror.NodeType{Name: "section"}
	headerType   = prosemirror.NodeType{Name: "header"}
	richTextType = prosemirror.NodeType{Name: "richText"}
	userType     = prosemirror.NodeType{Name: "user"}

	dividerNode = prosemirror.Node{Type: prosemirror.NodeType{Name: "divider"}}
)

func (s *DocumentsService) CreateOncallShiftHandoverMessage(sections []rez.OncallShiftHandoverSection, annotations []rez.OncallEventAnnotation, roster *ent.OncallRoster, endingShift *ent.OncallUserShift, startingShift *ent.OncallUserShift) (*rez.ContentNode, error) {
	var content []prosemirror.Node

	content = append(content, buildHandoverHeaderSection(roster, endingShift, startingShift), dividerNode)
	for _, section := range sections {
		if section.Kind == "annotations" {
			content = append(content, buildHandoverAnnotationsSection(annotations)...)
		} else {
			content = append(content, buildHandoverContentSection(section)...)
		}
	}
	content = append(content, dividerNode, buildHandoverFooterSection(endingShift))

	return &rez.ContentNode{Content: prosemirror.Fragment{Content: content}}, nil
}

func textNode(text string) prosemirror.Node {
	return prosemirror.Node{
		Type: textType,
		Text: text,
	}
}

func headerNode(content ...prosemirror.Node) prosemirror.Node {
	return prosemirror.Node{
		Type:    headerType,
		Content: prosemirror.Fragment{Content: content},
	}
}

func buildHandoverHeaderSection(roster *ent.OncallRoster, endingShift *ent.OncallUserShift, startingShift *ent.OncallUserShift) prosemirror.Node {
	titleNode := textNode(fmt.Sprintf(":pager: %s - Oncall Handover :pager:", roster.Name))

	usersNode := prosemirror.Node{
		Type: richTextType,
		Content: prosemirror.Fragment{Content: []prosemirror.Node{
			{Type: userType, Text: endingShift.Edges.User.ChatID},
			{Type: textType, Text: "to"},
			{Type: userType, Text: startingShift.Edges.User.ChatID},
		}},
	}

	contextNode := textNode(fmt.Sprintf("Shift Ending %s", endingShift.EndAt.Format(time.DateOnly)))

	return headerNode(titleNode, usersNode, contextNode)
}

func buildHandoverContentSection(section rez.OncallShiftHandoverSection) []prosemirror.Node {
	nodes := []prosemirror.Node{
		headerNode(textNode(section.Header)),
	}
	if section.Content != nil {
		nodes = append(nodes, *section.Content)
	} else {
		nodes = append(nodes, textNode("N/A"))
	}
	return nodes
}

func buildHandoverAnnotationsSection(annos []rez.OncallEventAnnotation) []prosemirror.Node {
	if len(annos) == 0 {
		return []prosemirror.Node{headerNode(textNode("Annotations")), textNode("N/A")}
	}
	return nil
}

func buildHandoverFooterSection(endingShift *ent.OncallUserShift) prosemirror.Node {
	endingShiftLink := fmt.Sprintf("%s/oncall/shifts/%s", rez.FrontendUrl, endingShift.ID)

	linkText := prosemirror.Node{
		Type:  richTextType,
		Text:  "View Shift Information in Rezible",
		Attrs: map[string]any{"href": endingShiftLink},
	}

	return prosemirror.Node{
		Type:    richTextType,
		Content: prosemirror.Fragment{Content: []prosemirror.Node{linkText}},
	}
}
