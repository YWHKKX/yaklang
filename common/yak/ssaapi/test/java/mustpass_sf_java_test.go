package java

import (
	"embed"
	_ "embed"
	"path/filepath"
	"testing"

	"golang.org/x/exp/slices"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/yaklang/yaklang/common/syntaxflow/sfvm"
	"github.com/yaklang/yaklang/common/utils/filesys"
	"github.com/yaklang/yaklang/common/yak/ssaapi"
)

//go:embed sample
var sample_code embed.FS

//go:embed mustpass
var sf_rules embed.FS

func Test_Debug(t *testing.T) {
	progs, err := ssaapi.ParseProject(
		filesys.NewEmbedFS(sample_code),
		ssaapi.WithLanguage(ssaapi.JAVA),
	)
	if err != nil {
		t.Fatalf("parse project error: %v", err)
	}
	Check(t, progs, "xxe.sf")
}

func TestCheckRuleInSource(t *testing.T) {
	// source
	prog, err := ssaapi.ParseProject(
		filesys.NewEmbedFS(sample_code),
		ssaapi.WithLanguage(ssaapi.JAVA),
	)
	if err != nil {
		t.Fatalf("parse project error: %v", err)
	}
	Check(t, prog)
}

func TestCheckRuleWithDatabase(t *testing.T) {
	programID := uuid.NewString()
	prog, err := ssaapi.ParseProject(
		filesys.NewEmbedFS(sample_code),
		ssaapi.WithLanguage(ssaapi.JAVA),
		ssaapi.WithDatabaseProgramName(programID),
	)
	if err != nil {
		t.Fatalf("parse project error: %v", err)
	}
	Check(t, prog)
}

func TestCheckRuleOnlyDatabase(t *testing.T) {
	programID := uuid.NewString()
	// compile with database
	{
		_, err := ssaapi.ParseProject(
			filesys.NewEmbedFS(sample_code),
			ssaapi.WithLanguage(ssaapi.JAVA),
			ssaapi.WithDatabaseProgramName(programID),
		)
		if err != nil {
			t.Fatalf("parse project error: %v", err)
		}
	}

	//  only database
	{
		prog, err := ssaapi.FromDatabase(programID)
		if err != nil {
			t.Fatalf("parse project error: %v", err)
		}
		Check(t, []*ssaapi.Program{prog})
	}
}

func Check(t *testing.T, progs []*ssaapi.Program, include ...string) {
	vs := sfvm.NewValues(lo.Map(progs, func(v *ssaapi.Program, _ int) sfvm.ValueOperator { return v }))
	vm := sfvm.NewSyntaxFlowVirtualMachine(sfvm.WithEnableDebug(false), sfvm.WithFailFast())
	entry, err := sf_rules.ReadDir("mustpass")
	if err != nil {
		t.Fatalf("no embed syntax files found: %v", err)
	}
	for _, f := range entry {
		if f.IsDir() {
			continue
		}
		path := filepath.Join("mustpass", f.Name())
		raw, err := sf_rules.ReadFile(path)
		if err != nil {
			t.Fatalf("cannot found syntax fs: %v", path)
		}
		if len(include) != 0 && !slices.Contains(include, f.Name()) {
			continue
		}
		frame, err := vm.Compile(string(raw))
		if err != nil {
			t.Fatalf("syntaxFlow compile error: %s", path)
		}
		t.Log("compile success: ", path)

		t.Run(f.Name(), func(t *testing.T) {
			res, err := frame.Feed(vs)
			if err != nil {
				t.Fatalf("feed error: %v", err)
			}
			res.Show()
		})
	}
}
