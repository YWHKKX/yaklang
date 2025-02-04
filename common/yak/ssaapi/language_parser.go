package ssaapi

import (
	"fmt"
	"io"
	"time"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/memedit"
	js2ssa "github.com/yaklang/yaklang/common/yak/JS2ssa"
	"github.com/yaklang/yaklang/common/yak/java/java2ssa"
	"github.com/yaklang/yaklang/common/yak/php/php2ssa"
	"github.com/yaklang/yaklang/common/yak/ssa"
	"github.com/yaklang/yaklang/common/yak/ssa4analyze"
	"github.com/yaklang/yaklang/common/yak/ssaapi/ssareducer"
	"github.com/yaklang/yaklang/common/yak/yak2ssa"
)

type Language string

const (
	Yak  Language = "yak"
	JS   Language = "js"
	PHP  Language = "php"
	JAVA Language = "java"
)

var LanguageBuilders = map[Language]ssa.Builder{
	Yak:  yak2ssa.Builder,
	JS:   js2ssa.Builder,
	PHP:  php2ssa.Builder,
	JAVA: java2ssa.Builder,
}

var AllLanguageBuilders = []ssa.Builder{
	php2ssa.Builder,
	java2ssa.Builder,

	yak2ssa.Builder,
	js2ssa.Builder,
}

func (c *config) parseProject() ([]*Program, error) {
	ret := make([]*Program, 0)

	programPath := c.programPath

	log.Infof("parse project in fs: %T, localpath: %v", c.fs, programPath)

	// parse project
	err := ssareducer.ReducerCompile(
		programPath, // base
		ssareducer.WithFileSystem(c.fs),
		ssareducer.WithEntryFiles(c.entryFile...),
		ssareducer.WithCompileMethod(func(path string, f io.Reader) (includeFiles []string, err error) {
			log.Debugf("start to compile from: %v", path)
			startTime := time.Now()

			raw, err := io.ReadAll(f)
			if err != nil {
				return nil, err
			}

			prog, err := c.parseSimple(path, memedit.NewMemEditor(string(raw)))
			endTime := time.Now()
			if err != nil {
				log.Debugf("parse %#v failed: %v", path, err)
				return nil, utils.Wrapf(err, "parse file %s error", path)
			}
			log.Infof("compile %s cost: %v", path, endTime.Sub(startTime))
			ret = append(ret, NewProgram(prog, c))
			exclude := prog.GetIncludeFiles()
			if len(exclude) > 0 {
				log.Infof("program include files: %v will not be as the entry from project", len(exclude))
			}
			return exclude, nil
		}),
	)
	if err != nil {
		return nil, utils.Wrap(err, "parse project error")
	}
	return ret, nil
}

func (c *config) parseFile() (ret *Program, err error) {
	prog, err := c.parseSimple("", c.originEditor)
	if err != nil {
		return nil, err
	}
	return NewProgram(prog, c), nil
}

func (c *config) feed(prog *ssa.Program, code *memedit.MemEditor) error {
	builder := prog.GetAndCreateFunctionBuilder("main", "main")
	if err := prog.Build("", code, builder); err != nil {
		return err
	}
	builder.Finish()
	ssa4analyze.RunAnalyzer(prog)
	return nil
}

func (c *config) parseSimple(path string, r *memedit.MemEditor) (ret *ssa.Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			ret = nil
			err = utils.Errorf("parse error with panic : %v", r)
			log.Errorf("parse error with panic : %v", err)
			utils.PrintCurrentGoroutineRuntimeStack()
		}
	}()

	prog, builder, err := c.init(path, r)
	if err != nil {
		return nil, err
	}
	// parse code
	if err := prog.Build(path, r, builder); err != nil {
		return nil, err
	}
	builder.Finish()
	ssa4analyze.RunAnalyzer(prog)
	prog.Finish()
	return prog, nil
}

var SkippedError = ssareducer.SkippedError

func (c *config) init(path string, editor *memedit.MemEditor) (*ssa.Program, *ssa.FunctionBuilder, error) {
	LanguageBuilder := c.Builder
	programName := c.DatabaseProgramName

	processBuilders := func(builders ...ssa.Builder) (ssa.Builder, error) {
		for _, instance := range builders {
			if instance.EnableExtraFileAnalyzer() {
				err := instance.ExtraFileAnalyze(c.fs, nil, path)
				if err != nil {
					return nil, err
				}
			}

			if instance.FilterFile(path) {
				return instance, nil
			}
		}
		return nil, utils.Wrapf(ssareducer.SkippedError, "file[%s] is not supported by any language builder, skip this file", path)
	}

	if path != "" {
		// TODO: whether to use the same programName for all program ?? when call ParseProject
		// programName += "-" + path
		var err error
		if LanguageBuilder != nil {
			LanguageBuilder, err = processBuilders(LanguageBuilder)
		} else {
			log.Warn("no language builder specified, try to use all language builders, but it may cause some error and extra file analyzing disabled")
			LanguageBuilder, err = processBuilders(AllLanguageBuilders...)
		}
		if err != nil {
			return nil, nil, err
		}
	} else {
		// path is empty, use language or YakLang as default
		if LanguageBuilder == nil {
			LanguageBuilder = LanguageBuilders[Yak]
			// log.Infof("use default language [%s] for empty path", Yak)
		}
	}

	prog := ssa.NewProgram(programName, c.fs, c.programPath)
	prog.Build = func(filePath string, src *memedit.MemEditor, fb *ssa.FunctionBuilder) error {
		// check builder
		if LanguageBuilder == nil {
			return utils.Errorf("not support language %s", c.language)
		}

		// get source code
		if src == nil {
			return fmt.Errorf("origin source code (MemEditor) is nil")
		}
		// backup old editor (source code)
		originEditor := fb.GetEditor()

		// include source code will change the context of the origin editor
		newCodeEditor := src
		newCodeEditor.SetUrl(filePath)
		fb.SetEditor(newCodeEditor) // set for current builder
		if originEditor != nil {
			originEditor.PushSourceCodeContext(newCodeEditor.SourceCodeMd5())
		}

		// push into program for recording what code is compiling
		prog.PushEditor(newCodeEditor)
		defer func() {
			// recover source code context
			fb.SetEditor(originEditor)
			prog.PopEditor()
		}()

		if ret := fb.GetEditor(); ret != nil {
			prog := fb.GetProgram()
			cache := prog.Cache
			progName, hash := prog.GetProgramName(), ret.SourceCodeMd5()
			if cache.IsExistedSourceCodeHash(progName, hash) {
				c.DatabaseProgramCacheHitter(fb)
			}
		} else {
			log.Warnf("(BUG or in DEBUG Mode)Range not found for %s", fb.GetName())
		}

		return LanguageBuilder.Build(src.GetSourceCode(), c.ignoreSyntaxErr, fb)
	}

	prog.PushEditor(editor)
	builder := prog.GetAndCreateFunctionBuilder("main", "main")
	// TODO: this extern info should be set in program
	builder.WithExternLib(c.externLib)
	builder.WithExternValue(c.externValue)
	builder.WithExternMethod(c.externMethod)
	builder.WithExternBuildValueHandler(c.externBuildValueHandler)
	builder.WithDefineFunction(c.defineFunc)
	builder.SetRangeInit(editor)
	return prog, builder, nil
}
