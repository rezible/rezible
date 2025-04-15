package prosemirror

/*
var (
	testHandoverSpecSchema = `{"marks":{"link":{"inclusive":true,"attrs":{"href":{"default":null},"target":{"default":"_blank"},"rel":{"default":"noopener noreferrer nofollow"},"class":{"default":null}},"parseDOM":[{"tag":"a[href]"}]},"bold":{"parseDOM":[{"tag":"strong"},{"tag":"b"},{"style":"font-weight"}]},"code":{"excludes":"_","code":true,"parseDOM":[{"tag":"code"}]},"italic":{"parseDOM":[{"tag":"em"},{"tag":"i"},{"style":"font-style=italic"}]},"strike":{"parseDOM":[{"tag":"s"},{"tag":"del"},{"tag":"strike"},{"style":"text-decoration","consuming":false}]}},"nodes":{"paragraph":{"content":"inline*","group":"block","parseDOM":[{"tag":"p"}]},"blockquote":{"content":"block+","group":"block","defining":true,"parseDOM":[{"tag":"blockquote"}]},"bulletList":{"content":"listItem+","group":"block list","parseDOM":[{"tag":"ul"}]},"codeBlock":{"content":"text*","marks":"","group":"block","code":true,"defining":true,"attrs":{"language":{"default":null}},"parseDOM":[{"tag":"pre","preserveWhitespace":"full"}]},"doc":{"content":"block+"},"hardBreak":{"group":"inline","inline":true,"selectable":false,"parseDOM":[{"tag":"br"}]},"heading":{"content":"inline*","group":"block","defining":true,"attrs":{"level":{"default":1}},"parseDOM":[{"tag":"h1","attrs":{"level":1}},{"tag":"h2","attrs":{"level":2}},{"tag":"h3","attrs":{"level":3}},{"tag":"h4","attrs":{"level":4}},{"tag":"h5","attrs":{"level":5}},{"tag":"h6","attrs":{"level":6}}]},"horizontalRule":{"group":"block","parseDOM":[{"tag":"hr"}]},"listItem":{"content":"paragraph block*","defining":true,"parseDOM":[{"tag":"li"}]},"orderedList":{"content":"listItem+","group":"block list","attrs":{"start":{"default":1},"type":{}},"parseDOM":[{"tag":"ol"}]},"text":{"group":"inline"},"image":{"group":"block","inline":false,"draggable":true,"attrs":{"src":{"default":null},"alt":{"default":null},"title":{"default":null}},"parseDOM":[{"tag":"img[src]:not([src^=\"data:\"])"}]}},"topNode":"doc"}`
	testHandoverContent2   = `[{"header":"Overview","kind":"regular","jsonContent":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"foo bar"}]}]}},{"header":"Handoff Tasks","kind":"regular","jsonContent":{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"a task"}]}]}]}]}},{"header":"Things to Monitor","kind":"regular","jsonContent":{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph"}]}]}]}},{"header":"Event Annotations","kind":"annotations"}]`
)

func registerSchema(t *testing.T) {
	var spec prosemirror.SchemaSpec
	if err := json.Unmarshal([]byte(testHandoverSpecSchema), &spec); err != nil {
		t.Errorf("failed to unmarshal: %s", err)
		t.FailNow()
	}
	schema, schemaErr := prosemirror.NewSchema(spec)
	if schemaErr != nil {
		t.Errorf("failed to load schema: %s", schemaErr)
		t.FailNow()
	}
	prosemirror.RegisterSchema(schema)
}

func TestConvertContent(t *testing.T) {
	registerSchema(t)

	var content []rez.OncallShiftHandoverSection
	if err := json.Unmarshal([]byte(testHandoverContent2), &content); err != nil {
		t.Errorf("failed to unmarshal: %s", err)
		t.FailNow()
	}

	var annos []*ent.OncallAnnotation
	if err := json.Unmarshal([]byte(testAnnotations), &annos); err != nil {
		t.Errorf("failed to unmarshal: %s", err)
		t.FailNow()
	}

	builder := handoverMessageBuilder{
		roster: &ent.OncallRoster{
			Name: "roster name",
		},
		senderId:    "foo",
		receiverId:  "foo",
		endingShift: &ent.OncallUserShift{},
	}

	if convErr := builder.build(content); convErr != nil {
		t.Errorf("failed to convertOncallHandoverToBlocks: %s", convErr)
		t.FailNow()
	}

	js, _ := json.Marshal(builder.blocks)
	fmt.Printf(`{"blocks": %s}`, js)
	fmt.Println()
}
*/
