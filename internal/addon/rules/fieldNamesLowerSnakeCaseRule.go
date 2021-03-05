package rules

import (
	"fmt"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/parser"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/yoheimuta/protolint/internal/osutil"
	"github.com/yoheimuta/protolint/linter/report"
	"github.com/yoheimuta/protolint/linter/strs"
	"github.com/yoheimuta/protolint/linter/visitor"
)

// FieldNamesLowerSnakeCaseRule verifies that all field names are underscore_separated_names.
// See https://developers.google.com/protocol-buffers/docs/style#message-and-field-names.
type FieldNamesLowerSnakeCaseRule struct {
	fixMode bool
}

// NewFieldNamesLowerSnakeCaseRule creates a new FieldNamesLowerSnakeCaseRule.
func NewFieldNamesLowerSnakeCaseRule(fixMode bool) FieldNamesLowerSnakeCaseRule {
	return FieldNamesLowerSnakeCaseRule{
		fixMode: fixMode,
	}
}

// ID returns the ID of this rule.
func (r FieldNamesLowerSnakeCaseRule) ID() string {
	return "FIELD_NAMES_LOWER_SNAKE_CASE"
}

// Purpose returns the purpose of this rule.
func (r FieldNamesLowerSnakeCaseRule) Purpose() string {
	return "Verifies that all field names are underscore_separated_names."
}

// IsOfficial decides whether or not this rule belongs to the official guide.
func (r FieldNamesLowerSnakeCaseRule) IsOfficial() bool {
	return true
}

// Apply applies the rule to the proto.
func (r FieldNamesLowerSnakeCaseRule) Apply(proto *parser.Proto) ([]report.Failure, error) {
	/*v := &fieldNamesLowerSnakeCaseVisitor{
		BaseAddVisitor: visitor.NewBaseAddVisitor(r.ID()),
	}
	*/
	fileName := proto.Meta.Filename
	lines, err := osutil.ReadAllLines(fileName, "\n")
	if err != nil {
		return nil, err
	}
	v := &fieldNamesLowerSnakeCaseVisitor{
		BaseAddVisitor:   visitor.NewBaseAddVisitor(r.ID()),
		protoLines:       lines,
		fixMode:          r.fixMode,
		newline:          "\n",
		notInsertNewline: true,
		protoFileName:    fileName,
		fieldIdent:       make(map[int][]fieldIdent),
	}
	return visitor.RunVisitor(v, proto, r.ID())
}

type fieldIdent struct {
	Name string
	meta meta.Meta
}

type fieldNamesLowerSnakeCaseVisitor struct {
	*visitor.BaseAddVisitor
	fixMode bool

	style        string
	protoLines   []string
	currentLevel int

	newline          string
	notInsertNewline bool
	protoFileName    string
	fieldIdent       map[int][]fieldIdent
}

// VisitField checks the field.
func (v *fieldNamesLowerSnakeCaseVisitor) VisitField(field *parser.Field) bool {
	if !strs.IsLowerSnakeCase(field.FieldName) {
		v.Record(field.Meta.Pos.Line, fieldIdent{
			Name: field.FieldName,
			meta: field.Meta,
		})
		v.AddFailuref(field.Meta.Pos, "Field name %q must be underscore_separated_names", field.FieldName)
	}
	return false
}

// VisitMapField checks the map field.
func (v *fieldNamesLowerSnakeCaseVisitor) VisitMapField(field *parser.MapField) bool {
	if !strs.IsLowerSnakeCase(field.MapName) {
		v.Record(field.Meta.Pos.Line, fieldIdent{
			Name: field.MapName,
			meta: field.Meta,
		})
		v.AddFailuref(field.Meta.Pos, "Field name %q must be underscore_separated_names", field.MapName)
	}
	return false
}

// VisitOneofField checks the oneof field.
func (v *fieldNamesLowerSnakeCaseVisitor) VisitOneofField(field *parser.OneofField) bool {
	if !strs.IsLowerSnakeCase(field.FieldName) {
		v.Record(field.Meta.Pos.Line, fieldIdent{
			Name: field.FieldName,
			meta: field.Meta,
		})

		v.AddFailuref(field.Meta.Pos, "Field name %q must be underscore_separated_names", field.FieldName)
	}
	return false
}
func (v *fieldNamesLowerSnakeCaseVisitor) Record(line int, ident fieldIdent) {
	v.fieldIdent[line] = append(v.fieldIdent[line], ident)
}
func (v *fieldNamesLowerSnakeCaseVisitor) Finally() error {
	if v.fixMode {
		return v.fix()
	}
	return nil
}

func (v *fieldNamesLowerSnakeCaseVisitor) fix() error {
	var fixedLines []string
	for i, line := range v.protoLines {
		if fixes, ok := v.fieldIdent[i+1]; ok {
			lines := strings.Fields(line)
			for j := len(fixes) - 1; 0 <= j; j-- {
				text := strs.ToLowerSnake(lines[fixes[j].meta.Pos.Column-1])
				fmt.Println("fix : replace ", lines[fixes[j].meta.Pos.Column-1], " with ", text)
				//text := line[start : start+len(fixes[j].Name)]
				//fmt.Println("--------", i, lines, fixes, fixes[j].Name, "----->", strs.ToLowerSnake(text), fixes[j].pos)
				lines[fixes[j].meta.Pos.Column-1] = text
				fixedLine := ""
				for ii, l := range lines {
					if ii == 0 {
						fixedLine = fixedLine + " "
					}

					if l != " " {
						fixedLine = fixedLine + " " + l
					}
				}
				fixedLines = append(fixedLines, fixedLine)
			}
		} else {
			fixedLines = append(fixedLines, line)
		}
	}
	return osutil.WriteLinesToExistingFile(v.protoFileName, fixedLines, v.newline)
}
