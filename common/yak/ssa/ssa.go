package ssa

import (
	"github.com/yaklang/yaklang/common/sca/dxtypes"
	"sync"

	"github.com/samber/lo"
	"github.com/yaklang/yaklang/common/log"

	"github.com/yaklang/yaklang/common/utils/memedit"
	"github.com/yaklang/yaklang/common/utils/omap"
	"github.com/yaklang/yaklang/common/yak/ssa/ssautil"
)

type ErrorLogger interface {
	NewError(ErrorKind, ErrorTag, string)
}

type GetIdIF interface {
	GetId() int64
}

type Instruction interface {
	ErrorLogger

	GetOpcode() Opcode

	// function
	GetFunc() *Function
	SetFunc(*Function)
	// block
	GetBlock() *BasicBlock
	SetBlock(*BasicBlock)
	// program
	GetProgram() *Program

	GetName() string
	SetName(variable string)
	GetVerboseName() string
	GetShortVerboseName() string
	SetVerboseName(string)
	SetIsAnnotation(bool)
	IsAnnotation() bool

	GetIdIF
	SetId(int64)

	// position
	GetRange() *Range
	SetRange(*Range)
	GetSourceCode() string
	GetSourceCodeContext(n int) string

	// extern
	IsExtern() bool
	SetExtern(bool)

	SelfDelete()
	IsBlock(string) bool

	IsCFGEnterBlock() ([]Instruction, bool)

	// Self means lazy instruction will check and return
	// real instruction will return itself
	Self() Instruction

	// IsLazy means this instruction is lazy, not loaded all from db
	IsLazy() bool
}

type (
	Users  []User
	Values []Value
)

// data-flow
type Node interface {
	// string
	String() string

	// for graph
	HasUsers() bool
	GetUsers() Users
	HasValues() bool
	GetValues() Values
	IsUndefined() bool
}
type Typed interface {
	// Node
	// type
	GetType() Type
	SetType(Type)
}

type MemberCall interface {
	// object  member caller
	IsObject() bool
	AddMember(Value, Value)
	GetMember(Value) (Value, bool)
	GetIndexMember(int) (Value, bool)
	GetStringMember(string) (Value, bool)
	DeleteMember( /*key*/ Value)   // delete by key
	GetAllMember() map[Value]Value // map[key]value
	ForEachMember(func(k Value, v Value) bool)

	// ReplaceMember( /* key */ Value /* value */, Value) // replace old-value with new-value

	// member, member callee
	IsMember() bool
	SetObject(Value)
	SetKey(Value)
	GetKey() Value
	GetObject() Value

	// ReplaceObject(Value) // replace old-object to new-object
}

type AssignAble interface {
	GetVariable(string) *Variable
	GetLastVariable() *Variable
	GetAllVariables() map[string]*Variable
	AddVariable(*Variable)
}

// basic handle item (interface)
type Value interface {
	Node
	Instruction
	MemberCall
	Typed
	Maskable
	AssignAble
	AddUser(User)
	RemoveUser(User)

	// reference, this value same as other value
	// in use-def chain, this value use contain other value
	AddReference(Value)
	Reference() Values
}

type Maskable interface {
	AddMask(Value)
	GetMask() []Value
	Masked() bool
}

type User interface {
	Node
	Instruction
	ReplaceValue(Value, Value)
}

type Build func(string, *memedit.MemEditor, *FunctionBuilder) error

type (
	packagePath     []string
	packagePathList []packagePath
)

// both instruction and value
type Program struct {
	// package list
	Name        string
	SCAPackages []*dxtypes.Package
	Packages    map[string]*Package

	editorStack *omap.OrderedMap[string, *memedit.MemEditor]
	editorMap   *omap.OrderedMap[string, *memedit.MemEditor]

	Cache *Cache

	// offset
	OffsetMap         map[int]*OffsetItem
	OffsetSortedSlice []int

	// package Loader
	Loader *ssautil.PackageFileLoader
	Build  Build

	errors SSAErrors

	// for build
	buildOnce sync.Once

	// cache hitter
	packagePathList packagePathList

	// extern lib
	cacheExternInstance     map[string]Value // lib and value
	externType              map[string]Type
	externBuildValueHandler map[string]func(b *FunctionBuilder, id string, v any) (value Value)
	ExternInstance          map[string]any
	ExternLib               map[string]map[string]any
}

type Package struct {
	Name string
	// point to program
	Prog *Program
	// function list
	Funcs map[string]*Function

	// class blue print
	ClassBluePrint map[string]*ClassBluePrint
}

// implement Value
type Function struct {
	anValue

	isMethod   bool
	methodName string

	// package, double link
	Package *Package

	// Type
	Type *FunctionType

	// just function parameter
	Params      []Value
	ParamLength int
	// for closure function
	FreeValues map[string]Value // store the captured variable form parent-function, just contain name, and type is Parameter
	// parameter member call
	// ParameterMembers []*ParameterMember
	ParameterMembers []Value
	// function side effects
	SideEffects []*FunctionSideEffect

	// closure function double link. parentFunc <-> childFuncs
	parent     Value   // parent function;  can be nil if there is no parent function
	ChildFuncs []Value // child function within this function

	Return []Value

	// BasicBlock list
	Blocks []Instruction
	// First and End block
	EnterBlock Instruction
	ExitBlock  Instruction
	// For Defer  semantic
	// this block will always execute when the function exits,
	// regardless of whether the function returns normally or exits due to a panic.
	DeferBlock Instruction

	// ssa error
	errComment ErrorComment

	// ================  for build
	// builder
	builder *FunctionBuilder
	// this function is variadic parameter, for function type create
	hasEllipsis bool
	// generic
	isGeneric bool
}

func (f *Function) SetMethodName(name string) {
	f.isMethod = true
	f.methodName = name
}

func (f *Function) GetMethodName() string {
	return f.methodName
}

func (f *Function) FirstBlockInstruction() []Instruction {
	if len(f.Blocks) > 0 {
		firstBlock := f.Blocks[0]
		if block, ok := ToBasicBlock(firstBlock); ok {
			return block.Insts
		} else {
			log.Warnf("function %s first block is not a basic block", f.GetName())
		}
	}
	return nil
}

var (
	_ Node  = (*Function)(nil)
	_ Value = (*Function)(nil)
)

// implement Value
type BasicBlock struct {
	anValue

	Index int
	// BasicBlock graph
	Preds, Succs []Value

	/*
		if Condition == true: this block reach
	*/
	setReachable bool
	canBeReached int
	Condition    Value

	// instruction list
	Insts []Instruction
	Phis  []Value

	// error catch
	Handler *ErrorHandler

	// for build
	ScopeTable ScopeIF
	finish     bool // if emitJump finish!
}

func (b *BasicBlock) SetReachable(boolean bool) {
	b.setReachable = true
	if boolean {
		b.canBeReached = 1
	} else {
		b.canBeReached = -1
	}
}

func (b *BasicBlock) IsCFGEnterBlock() ([]Instruction, bool) {
	if len(b.Insts) <= 0 {
		return nil, false
	}
	jmp, err := lo.Last(b.Insts)
	if err != nil {
		return nil, false
	}

	_, ok := jmp.(*LazyInstruction)
	if ok {
		jmp = jmp.(*LazyInstruction).Self()
	}

	switch ret := jmp.(type) {
	case *Jump:
		if ret.To == nil {
			log.Warnf("Jump To is nil: %T", ret)
			return nil, false
		}

		toBlock, ok := ToBasicBlock(ret.To)
		if !ok {
			log.Warnf("Jump To is not *BasicBlock: %T", ret.To)
			return nil, false
		}

		last, err := lo.Last(toBlock.Insts)
		if err != nil {
			return nil, false
		}
		// fetch essential instructions via jump
		// if else(elif) condition
		// for loop condition
		// switch condition (label)
		if last.IsLazy() {
			last = last.Self()
		}
		switch ret := last.(type) {
		case *If:
			var ifs []*If
			ifs = append(ifs, ret)
			results := ret.GetSiblings()
			if len(results) > 0 {
				ifs = append(ifs, results...)
			}
			return lo.Map(ifs, func(a *If, i int) Instruction {
				return a
			}), true
		case *Switch:
			log.Warn("Swtich Statement (Condition/Label value should contains jmp) WARNING")
			return lo.Map(ret.Label, func(label SwitchLabel, i int) Instruction {
				var result Instruction = label.Value
				return result
			}), true
		case *Loop:
			log.Warn("Loop Statement (Condition/Label value should contains jmp) WARNING")
			return []Instruction{
				ret.Cond,
			}, true
		default:
			log.Warnf("unsupoorted CFG Entry Instruction: %T", ret)
		}
		return nil, false
	default:
		return nil, false
	}
}

func (b *BasicBlock) GetType() Type {
	return nil
}

func (b *BasicBlock) SetType(ts Type) {
}

var (
	_ Node  = (*BasicBlock)(nil)
	_ Value = (*BasicBlock)(nil)
)

// =========================================  Value ===============================
// ================================= Spec Value

// ----------- Phi
type Phi struct {
	anValue

	CFGEntryBasicBlock Value

	Edge []Value // edge[i] from phi.Block.Preds[i]
	//	what instruction create this control-flow merge?
	// branch *Instruction // loop or if :
}

var (
	_ Node        = (*Phi)(nil)
	_ Value       = (*Phi)(nil)
	_ User        = (*Phi)(nil)
	_ Instruction = (*Phi)(nil)
)

// ----------- externLib
type ExternLib struct {
	anValue

	table   map[string]any
	builder *FunctionBuilder

	MemberMap map[string]Value
	Member    []Value
}

var (
	_ Node  = (*ExternLib)(nil)
	_ Value = (*ExternLib)(nil)
	_ User  = (*ExternLib)(nil)
)

type ParameterMemberCallKind int

const (
	NoMemberCall ParameterMemberCallKind = iota
	ParameterMemberCall
	FreeValueMemberCall
)

type parameterMemberInner struct {
	ObjectName            string
	MemberCallKind        ParameterMemberCallKind
	MemberCallObjectIndex int    // for Parameter
	MemberCallObjectName  string // for FreeValue
	MemberCallKey         Value
}

func newParameterMember(obj *Parameter, key Value) *parameterMemberInner {
	new := &parameterMemberInner{
		ObjectName:    obj.GetName(),
		MemberCallKey: key,
	}

	if obj.IsFreeValue {
		new.MemberCallKind = FreeValueMemberCall
		new.MemberCallObjectName = obj.GetName()
	} else {
		new.MemberCallKind = ParameterMemberCall
		new.MemberCallObjectIndex = obj.FormalParameterIndex
	}
	return new
}

func (p *parameterMemberInner) Get(c *Call) (obj Value, ok bool) {
	switch p.MemberCallKind {
	case NoMemberCall:
		return
	case ParameterMemberCall:
		if p.MemberCallObjectIndex >= len(c.Args) {
			// log.Errorf("handleCalleeFunction: memberCallObjectIndex out of range %d vs len: %d", p.MemberCallObjectIndex, len(c.Args))
			return
		}
		return c.Args[p.MemberCallObjectIndex], true
	case FreeValueMemberCall:
		obj, ok = c.Binding[p.MemberCallObjectName]
		return obj, ok
	}
	return
}

type ParameterMember struct {
	anValue
	FormalParameterIndex int
	*parameterMemberInner
}

var (
	_ Node  = (*Parameter)(nil)
	_ Value = (*Parameter)(nil)
)

// ----------- Parameter
type Parameter struct {
	anValue

	// for FreeValue
	IsFreeValue  bool
	defaultValue Value

	// Parameter Index
	FormalParameterIndex int
}

func (p *Parameter) GetDefault() Value {
	return p.defaultValue
}

func (p *Parameter) SetDefault(v Value) {
	p.defaultValue = v
	v.AddReference(p)
}

var (
	_ Node  = (*Parameter)(nil)
	_ Value = (*Parameter)(nil)
)

// ================================= Normal Value

// ----------- Const
// ConstInst also have block pointer, which block set this const to variable
type ConstInst struct {
	*Const
	anValue
	Unary      int
	isIdentify bool // field key
	Origin     User
}

// ConstInst cont set Type
func (c *ConstInst) GetType() Type   { return c.anValue.GetType() }
func (c *ConstInst) SetType(ts Type) { c.anValue.SetType(ts) }

var (
	_ Node        = (*ConstInst)(nil)
	_ Value       = (*ConstInst)(nil)
	_ User        = (*ConstInst)(nil)
	_ Instruction = (*ConstInst)(nil)
)

// ----------- Undefined

// UndefinedKind : mark undefined value type
type UndefinedKind int

const (
	// UndefinedValueInValid normal undefined value
	UndefinedValueInValid UndefinedKind = iota
	// UndefinedValueValid is variable only declare
	UndefinedValueValid
	// UndefinedMemberInValid member call but not this key
	UndefinedMemberInValid
	// UndefinedMemberValid member call, has this key, but not this value, this shouldn't mark error
	UndefinedMemberValid
)

type Undefined struct {
	anValue
	Kind UndefinedKind
}

func (u *Undefined) IsUndefined() bool { return true }

var (
	_ Node        = (*Undefined)(nil)
	_ Value       = (*Undefined)(nil)
	_ Instruction = (*Undefined)(nil)
)

// ----------- BinOp
type BinOp struct {
	anValue
	Op   BinaryOpcode
	X, Y Value
}

var (
	_ Value       = (*BinOp)(nil)
	_ User        = (*BinOp)(nil)
	_ Node        = (*BinOp)(nil)
	_ Instruction = (*BinOp)(nil)
)

type UnOp struct {
	anValue

	Op UnaryOpcode
	X  Value
}

var (
	_ Value       = (*UnOp)(nil)
	_ User        = (*UnOp)(nil)
	_ Node        = (*UnOp)(nil)
	_ Instruction = (*UnOp)(nil)
)

// ================================= Function Call

// ----------- Call
// call instruction call method function  with args as argument
type Call struct {
	// call is a value
	anValue

	// for call function
	Method    Value
	Args      []Value
	Binding   map[string]Value
	ArgMember []Value

	// go function
	Async  bool
	Unpack bool

	// caller
	// caller Value
	// ~ drop error
	IsDropError bool
	IsEllipsis  bool
}

var (
	_ Node        = (*Call)(nil)
	_ Value       = (*Call)(nil)
	_ User        = (*Call)(nil)
	_ Instruction = (*Call)(nil)
)

// ----------- SideEffect
type SideEffect struct {
	anValue
	CallSite Value // call instruction
	Value    Value // modify to this value
}

var (
	_ Node        = (*SideEffect)(nil)
	_ Value       = (*SideEffect)(nil)
	_ User        = (*SideEffect)(nil)
	_ Instruction = (*SideEffect)(nil)
)

// ----------- Return
// The Return instruction returns values and control back to the calling
// function.
type Return struct {
	anValue
	Results []Value
}

var (
	_ Node        = (*Return)(nil)
	_ User        = (*Return)(nil)
	_ Value       = (*Return)(nil)
	_ Instruction = (*Return)(nil)
)

// ================================= Memory Value

// ----------- Make
type Make struct {
	anValue

	// when slice
	low, high, step Value

	parentI Value // parent interface

	// when slice or map
	Len, Cap Value
}

var (
	_ Node        = (*Make)(nil)
	_ Value       = (*Make)(nil)
	_ User        = (*Make)(nil)
	_ Instruction = (*Make)(nil)
)

// ------------- Next
type Next struct {
	anValue
	Iter   Value
	InNext bool // "in" grammar
}

var (
	_ Node        = (*Next)(nil)
	_ User        = (*Next)(nil)
	_ Value       = (*Next)(nil)
	_ Instruction = (*Next)(nil)
)

// ================================= Assert Value

// ----------- assert
type Assert struct {
	anInstruction

	Cond     Value
	Msg      string
	MsgValue Value
}

var (
	_ Node        = (*Assert)(nil)
	_ User        = (*Assert)(nil)
	_ Instruction = (*Assert)(nil)
)

// ----------- Type-cast
// cast value -> type
type TypeCast struct {
	anValue

	Value Value
}

var (
	_ Node        = (*TypeCast)(nil)
	_ Value       = (*TypeCast)(nil)
	_ User        = (*TypeCast)(nil)
	_ Instruction = (*TypeCast)(nil)
)

// ------------- type value
type TypeValue struct {
	anValue
}

var (
	_ Node        = (*TypeValue)(nil)
	_ Value       = (*TypeValue)(nil)
	_ Instruction = (*TypeValue)(nil)
)

// ================================= Error Handler

// ------------- ErrorHandler
type ErrorHandler struct {
	anInstruction
	catchs           []*BasicBlock
	try, final, done *BasicBlock
}

var _ Instruction = (*ErrorHandler)(nil)

// -------------- PANIC
type Panic struct {
	anValue
	Info Value
}

var (
	_ Node        = (*Panic)(nil)
	_ User        = (*Panic)(nil)
	_ Instruction = (*Panic)(nil)
)

// --------------- RECOVER
type Recover struct {
	anValue
}

var (
	_ Node        = (*Recover)(nil)
	_ Value       = (*Recover)(nil)
	_ Instruction = (*Recover)(nil)
)

// control-flow instructions  ----------------------------------------
// jump / if / return / call / switch

// ----------- Jump
// The Jump instruction transfers control to the sole successor of its
// owning block.
//
// the block containing Jump instruction only have one successor block
type Jump struct {
	anInstruction
	To Value
}

var _ Instruction = (*Jump)(nil)

// ----------- IF
// The If instruction transfers control to one of the two successors
// of its owning block, depending on the boolean Cond: the first if
// true, the second if false.
type If struct {
	anInstruction

	Cond  Value
	True  Value
	False Value
}

func (i *If) GetSiblings() []*If {
	return i.getSiblings(nil)
}

func (i *If) getSiblings(m map[int64]struct{}) []*If {
	if m == nil {
		m = make(map[int64]struct{})
	}
	_, visited := m[i.GetId()]
	if visited {
		return nil
	}

	var ifs []*If
	if i.False == nil {
		return nil
	}

	falseBlock, ok := ToBasicBlock(i.False)
	if !ok {
		return nil
	}
	raw := falseBlock.LastInst()
	lastIf, ok := ToIfInstruction(raw)
	if ok {
		m[lastIf.GetId()] = struct{}{}
		ifs = append(ifs, lastIf)
		ifs = append(ifs, lastIf.getSiblings(m)...)
	}
	return ifs
}

var (
	_ Node        = (*If)(nil)
	_ User        = (*If)(nil)
	_ Instruction = (*If)(nil)
)

// ----------- For
// for loop
type Loop struct {
	anInstruction

	Body, Exit *BasicBlock

	Init, Cond, Step Value
	Key              Value
}

var (
	_ Node        = (*Loop)(nil)
	_ User        = (*Loop)(nil)
	_ Instruction = (*Loop)(nil)
)

// ----------- Switch
type SwitchLabel struct {
	Value Value
	Dest  *BasicBlock
}

func NewSwitchLabel(v Value, dest *BasicBlock) SwitchLabel {
	return SwitchLabel{
		Value: v,
		Dest:  dest,
	}
}

type Switch struct {
	anInstruction

	Cond         Value
	DefaultBlock *BasicBlock

	Label []SwitchLabel
}

var (
	_ Node        = (*Switch)(nil)
	_ User        = (*Switch)(nil)
	_ Instruction = (*Switch)(nil)
)
