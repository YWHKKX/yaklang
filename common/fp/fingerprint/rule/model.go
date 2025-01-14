package rule

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/fp/webfingerprint"
	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"golang.org/x/exp/maps"
	"strings"
)

type MatchResource struct {
	Data     []byte
	Port     int
	Path     string
	Protocol string
}

func NewHttpResource(data []byte) *MatchResource {
	return &MatchResource{
		Data:     data,
		Protocol: "http",
	}
}

type GeneralRule struct {
	gorm.Model
	*CPE
	WebPath         string
	ExtInfo         string
	MatchExpression string `gorm:"uniqueIndex"`
}

func (g *GeneralRule) String() string {
	items := []string{}
	cpeStr := g.CPE.String()
	items = append(items, fmt.Sprintf("cpe:%s", cpeStr))

	if g.WebPath != "" {
		items = append(items, fmt.Sprintf("webpath:%s", g.WebPath))
	}
	if g.ExtInfo != "" {
		items = append(items, fmt.Sprintf("info:%s", g.ExtInfo))
	}
	items = append(items, fmt.Sprintf("rule:%s", g.MatchExpression))
	return strings.Join(items, " ")
}

func NewEmptyGeneralRule() *GeneralRule {
	return &GeneralRule{
		CPE: &CPE{},
	}
}
func ParseGeneralRule(s string) (*GeneralRule, error) {
	rule := NewEmptyGeneralRule()
	infoItems := map[string]func(s string){"cpe:": func(s string) {
		cpe, err := ParseToCPE(s)
		if err != nil {
			log.Error(err)
		}
		rule.CPE = cpe
	}, "webpath:": func(s string) {
		rule.WebPath = s
	}, "info:": func(s string) {
		rule.ExtInfo = s
	}, "rule:": func(s string) {
		rule.MatchExpression = s
	}}
	keys := maps.Keys(infoItems)
	res := utils.IndexAllSubstrings(s, keys...)

	if len(res) > 0 {
		pre := res[0]
		for _, info := range res[1:] {
			v, ok := infoItems[keys[pre[0]]]
			if !ok {
				continue
			}
			v(s[pre[1]+len(keys[pre[0]]) : info[1]])
			pre = info
		}
		v, ok := infoItems[keys[pre[0]]]
		if ok {
			v(s[pre[1]+len(keys[pre[0]]):])
		}
	}

	if rule.MatchExpression == "" {
		return nil, errors.New("not set rule")
	}
	return rule, nil
}
func init() {
	db := consts.GetGormProjectDatabase()
	db.AutoMigrate(&GeneralRule{})
}

type MatchMethodParam struct {
	ExtParams map[string]any
	Info      *CPE

	// regexp
	RegexpPattern string
	Keyword       *webfingerprint.KeywordMatcher

	// complex
	Condition string
	SubRules  []*FingerPrintRule

	// http header
	HeaderKey       string
	HeaderMatchRule *FingerPrintRule

	//md5
	Md5 string

	// exp
	Params []any
	Op     string
}

type FingerPrintRule struct {
	opCodes    []*OpCode
	Method     string
	WebPath    string
	MatchParam *MatchMethodParam
}

func NewEmptyFingerPrintRule() *FingerPrintRule {
	return &FingerPrintRule{
		MatchParam: &MatchMethodParam{},
	}
}

type CPE struct {
	Part     string `yaml:"part,omitempty" json:"part"`
	Vendor   string `yaml:"vendor,omitempty" json:"vendor"`
	Product  string `yaml:"product,omitempty" json:"product"`
	Version  string `yaml:"version,omitempty" json:"version"`
	Update   string `yaml:"update,omitempty" json:"update"`
	Edition  string `yaml:"edition,omitempty" json:"edition"`
	Language string `yaml:"language,omitempty" json:"language"`
}

func (f *FingerPrintRule) preToOpCodes() []*OpCode {
	res := []*OpCode{}
	pushCode := func(code *OpCode) {
		res = append(res, code)
	}
	switch f.Method {
	case "exp":
		if len(f.MatchParam.Params) != 2 {
			return nil
		}
		params := []any{}
		for _, param := range f.MatchParam.Params {
			params = append(params, param)
		}
		ref := params[0].(string)
		value := params[1]
		if strings.HasPrefix(ref, "header_") {
			pushCode(&OpCode{Op: OpExtractData, data: []any{f.WebPath, "header_item", strings.TrimLeft(ref, "header_")}})
			pushCode(&OpCode{Op: OpPush, data: []any{value}})
		} else {
			pushCode(&OpCode{Op: OpExtractData, data: []any{f.WebPath, ref}})
			pushCode(&OpCode{Op: OpPush, data: []any{value}})
		}
		switch f.MatchParam.Op {
		case "=":
			pushCode(&OpCode{Op: OpContains})
		case "!=":
			pushCode(&OpCode{Op: OpContains})
			pushCode(&OpCode{Op: OpNot})
		case "==":
			pushCode(&OpCode{Op: OpEqual})
		case "!==":
			pushCode(&OpCode{Op: OpEqual})
			pushCode(&OpCode{Op: OpNot})
		case "~=":
			pushCode(&OpCode{Op: OpRegexpMatch})
		default:
			log.Errorf("not supported op: %s", f.MatchParam.Op)
			return nil
		}
	case "regexp":
		pushCode(&OpCode{Op: OpData, data: []any{f.WebPath}})
		pushCode(&OpCode{Op: OpPush, data: []any{f.MatchParam.RegexpPattern}})
		extGroup := []any{f.MatchParam.Keyword.VersionIndex, f.MatchParam.Keyword.ProductIndex, f.MatchParam.Keyword.VersionIndex, f.MatchParam.Keyword.UpdateIndex, f.MatchParam.Keyword.EditionIndex, f.MatchParam.Keyword.LanguageIndex}
		if !funk.Any(extGroup...) {
			extGroup = nil
		}
		pushCode(&OpCode{Op: OpRegexpMatch, data: extGroup})
	case "md5":
		pushCode(&OpCode{Op: OpExtractData, data: []any{f.WebPath, "md5"}})
		pushCode(&OpCode{Op: OpPush, data: []any{f.MatchParam.Md5}})
		pushCode(&OpCode{Op: OpEqual})
	case "http_header":
		pushCode(&OpCode{Op: OpExtractData, data: []any{f.WebPath, "header_item", f.MatchParam.HeaderKey}})
		subParam := f.MatchParam.HeaderMatchRule.MatchParam
		pushCode(&OpCode{Op: OpPush, data: []any{subParam.RegexpPattern}})
		extGroup := []any{subParam.Keyword.VersionIndex, subParam.Keyword.ProductIndex, subParam.Keyword.VersionIndex, subParam.Keyword.UpdateIndex, subParam.Keyword.EditionIndex, subParam.Keyword.LanguageIndex}
		if !funk.Any(extGroup...) {
			extGroup = nil
		}
		pushCode(&OpCode{Op: OpRegexpMatch, data: extGroup})
	case "complex":
		jmpPoint := map[*OpCode]int{}
		codes := []*OpCode{}
		switch f.MatchParam.Condition {
		case "or":
			for i, rule := range f.MatchParam.SubRules {
				codes = append(codes, rule.preToOpCodes()...)
				if i == len(f.MatchParam.SubRules)-1 {
					continue
				}
				code := &OpCode{Op: OpOr}
				jmpPoint[code] = len(codes)
				codes = append(codes, code)
			}
			res = append(res, codes...)
			for opCode, i := range jmpPoint {
				opCode.data = []any{len(codes) - i}
			}
		case "and":
			for i, rule := range f.MatchParam.SubRules {
				codes = append(codes, rule.preToOpCodes()...)
				if i == len(f.MatchParam.SubRules)-1 {
					continue
				}
				code := &OpCode{Op: OpAnd}
				jmpPoint[code] = len(codes)
				codes = append(codes, code)
			}
			res = append(res, codes...)
			for opCode, i := range jmpPoint {
				opCode.data = []any{len(codes) - i}
			}
		}
	default:
		return nil
	}
	return res
}
func (f *FingerPrintRule) ToOpCodes() []*OpCode {
	if f.opCodes != nil {
		return f.opCodes
	}
	codes := f.preToOpCodes()
	codes = append(codes, &OpCode{Op: OpInfo, data: []any{f.MatchParam.Info}})
	f.opCodes = codes
	return codes
}

func (c *CPE) Init() {
	if c.Part == "" {
		c.Part = "a"
	}

	setWildstart := func(raw *string) {
		if *raw == "" {
			*raw = "*"
		}
	}

	setWildstart(&c.Vendor)
	setWildstart(&c.Product)
	setWildstart(&c.Version)
	setWildstart(&c.Update)
	setWildstart(&c.Edition)
	setWildstart(&c.Language)
}

func (c *CPE) String() string {
	c.Init()
	raw := fmt.Sprintf("cpe:/%s:%s:%s:%s:%s:%s:%s", c.Part, c.Vendor, c.Product, c.Version, c.Update, c.Edition, c.Language)
	raw = strings.ReplaceAll(raw, " ", "_")
	raw = strings.ToLower(raw)
	return raw
}
func ParseToCPE(cpe string) (*CPE, error) {
	if (!strings.HasPrefix(cpe, "cpe:/")) && (!strings.HasPrefix(cpe, "cpe:2.3:")) {
		return nil, utils.Errorf("raw [%s] is not a valid cpe", cpe)
	}

	if strings.HasPrefix(cpe, "cpe:2.3:") {
		cpe = strings.Replace(cpe, "cpe:2.3:", "cpe:/", 1)
	}

	var cpeArgs [7]string
	s := strings.Split(cpe, ":")
	for i := 1; i <= len(s)-1; i++ {
		ret := strings.ReplaceAll(s[i], "%", "")
		cpeArgs[i-1] = ret
		if i == 7 {
			break
		}
	}
	cpeArgs[0] = cpeArgs[0][1:]
	cpeModel := &CPE{
		Part:     cpeArgs[0],
		Vendor:   cpeArgs[1],
		Product:  cpeArgs[2],
		Version:  cpeArgs[3],
		Update:   cpeArgs[4],
		Edition:  cpeArgs[5],
		Language: cpeArgs[6],
	}
	cpeModel.Init()
	return cpeModel, nil
}
