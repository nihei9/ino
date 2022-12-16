// Code generated by vartan-go. DO NOT EDIT.
package parser

import (
	"fmt"
	"io"
)

type ModeID int

func (id ModeID) Int() int {
	return int(id)
}

type StateID int

func (id StateID) Int() int {
	return int(id)
}

type KindID int

func (id KindID) Int() int {
	return int(id)
}

type ModeKindID int

func (id ModeKindID) Int() int {
	return int(id)
}

type LexSpec interface {
	InitialMode() ModeID
	Pop(mode ModeID, modeKind ModeKindID) bool
	Push(mode ModeID, modeKind ModeKindID) (ModeID, bool)
	ModeName(mode ModeID) string
	InitialState(mode ModeID) StateID
	NextState(mode ModeID, state StateID, v int) (StateID, bool)
	Accept(mode ModeID, state StateID) (ModeKindID, bool)
	KindIDAndName(mode ModeID, modeKind ModeKindID) (KindID, string)
}

// Token representes a token.
type Token struct {
	// ModeID is an ID of a lex mode.
	ModeID ModeID

	// KindID is an ID of a kind. This is unique among all modes.
	KindID KindID

	// ModeKindID is an ID of a lexical kind. This is unique only within a mode.
	// Note that you need to use KindID field if you want to identify a kind across all modes.
	ModeKindID ModeKindID

	// BytePos is a byte position where a token appears.
	BytePos int

	// ByteLen is a length of a token.
	ByteLen int

	// Row is a row number where a token appears.
	Row int

	// Col is a column number where a token appears.
	// Note that Col is counted in code points, not bytes.
	Col int

	// Lexeme is a byte sequence matched a pattern of a lexical specification.
	Lexeme []byte

	// When this field is true, it means the token is the EOF token.
	EOF bool

	// When this field is true, it means the token is an error token.
	Invalid bool
}

type LexerOption func(l *Lexer) error

// DisableModeTransition disables the active mode transition. Thus, even if the lexical specification has the push and pop
// operations, the lexer doesn't perform these operations. When the lexical specification has multiple modes, and this option is
// enabled, you need to call the Lexer.Push and Lexer.Pop methods to perform the mode transition. You can use the Lexer.Mode method
// to know the current lex mode.
func DisableModeTransition() LexerOption {
	return func(l *Lexer) error {
		l.passiveModeTran = true
		return nil
	}
}

type lexerState struct {
	srcPtr int
	row    int
	col    int
}

type Lexer struct {
	spec              LexSpec
	src               []byte
	state             lexerState
	lastAcceptedState lexerState
	tokBuf            []*Token
	modeStack         []ModeID
	passiveModeTran   bool
}

// NewLexer returns a new lexer.
func NewLexer(spec LexSpec, src io.Reader, opts ...LexerOption) (*Lexer, error) {
	b, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	l := &Lexer{
		spec: spec,
		src:  b,
		state: lexerState{
			srcPtr: 0,
			row:    0,
			col:    0,
		},
		lastAcceptedState: lexerState{
			srcPtr: 0,
			row:    0,
			col:    0,
		},
		modeStack: []ModeID{
			spec.InitialMode(),
		},
		passiveModeTran: false,
	}
	for _, opt := range opts {
		err := opt(l)
		if err != nil {
			return nil, err
		}
	}

	return l, nil
}

// Next returns a next token.
func (l *Lexer) Next() (*Token, error) {
	if len(l.tokBuf) > 0 {
		tok := l.tokBuf[0]
		l.tokBuf = l.tokBuf[1:]
		return tok, nil
	}

	tok, err := l.nextAndTransition()
	if err != nil {
		return nil, err
	}
	if !tok.Invalid {
		return tok, nil
	}
	errTok := tok
	for {
		tok, err = l.nextAndTransition()
		if err != nil {
			return nil, err
		}
		if !tok.Invalid {
			break
		}
		errTok.ByteLen += tok.ByteLen
		errTok.Lexeme = append(errTok.Lexeme, tok.Lexeme...)
	}
	l.tokBuf = append(l.tokBuf, tok)

	return errTok, nil
}

func (l *Lexer) nextAndTransition() (*Token, error) {
	tok, err := l.next()
	if err != nil {
		return nil, err
	}
	if tok.EOF || tok.Invalid {
		return tok, nil
	}
	if l.passiveModeTran {
		return tok, nil
	}
	mode := l.Mode()
	if l.spec.Pop(mode, tok.ModeKindID) {
		err := l.PopMode()
		if err != nil {
			return nil, err
		}
	}
	if mode, ok := l.spec.Push(mode, tok.ModeKindID); ok {
		l.PushMode(mode)
	}
	// The checking length of the mode stack must be at after pop and push operations because those operations can be performed
	// at the same time. When the mode stack has just one element and popped it, the mode stack will be temporarily emptied.
	// However, since a push operation may be performed immediately after it, the lexer allows the stack to be temporarily empty.
	if len(l.modeStack) == 0 {
		return nil, fmt.Errorf("a mode stack must have at least one element")
	}
	return tok, nil
}

func (l *Lexer) next() (*Token, error) {
	mode := l.Mode()
	state := l.spec.InitialState(mode)
	buf := []byte{}
	startPos := l.state.srcPtr
	row := l.state.row
	col := l.state.col
	var tok *Token
	for {
		v, eof := l.read()
		if eof {
			if tok != nil {
				l.revert()
				return tok, nil
			}
			// When `buf` has unaccepted data and reads the EOF, the lexer treats the buffered data as an invalid token.
			if len(buf) > 0 {
				return &Token{
					ModeID:     mode,
					ModeKindID: 0,
					BytePos:    startPos,
					ByteLen:    l.state.srcPtr - startPos,
					Lexeme:     buf,
					Row:        row,
					Col:        col,
					Invalid:    true,
				}, nil
			}
			return &Token{
				ModeID:     mode,
				ModeKindID: 0,
				BytePos:    startPos,
				Row:        row,
				Col:        col,
				EOF:        true,
			}, nil
		}
		buf = append(buf, v)
		nextState, ok := l.spec.NextState(mode, state, int(v))
		if !ok {
			if tok != nil {
				l.revert()
				return tok, nil
			}
			return &Token{
				ModeID:     mode,
				ModeKindID: 0,
				BytePos:    startPos,
				ByteLen:    l.state.srcPtr - startPos,
				Lexeme:     buf,
				Row:        row,
				Col:        col,
				Invalid:    true,
			}, nil
		}
		state = nextState
		if modeKindID, ok := l.spec.Accept(mode, state); ok {
			kindID, _ := l.spec.KindIDAndName(mode, modeKindID)
			tok = &Token{
				ModeID:     mode,
				KindID:     kindID,
				ModeKindID: modeKindID,
				BytePos:    startPos,
				ByteLen:    l.state.srcPtr - startPos,
				Lexeme:     buf,
				Row:        row,
				Col:        col,
			}
			l.accept()
		}
	}
}

// Mode returns the current lex mode.
func (l *Lexer) Mode() ModeID {
	return l.modeStack[len(l.modeStack)-1]
}

// PushMode adds a lex mode onto the mode stack.
func (l *Lexer) PushMode(mode ModeID) {
	l.modeStack = append(l.modeStack, mode)
}

// PopMode removes a lex mode from the top of the mode stack.
func (l *Lexer) PopMode() error {
	sLen := len(l.modeStack)
	if sLen == 0 {
		return fmt.Errorf("cannot pop a lex mode from a lex mode stack any more")
	}
	l.modeStack = l.modeStack[:sLen-1]
	return nil
}

func (l *Lexer) read() (byte, bool) {
	if l.state.srcPtr >= len(l.src) {
		return 0, true
	}

	b := l.src[l.state.srcPtr]
	l.state.srcPtr++

	// Count the token positions.
	// The driver treats LF as the end of lines and counts columns in code points, not bytes.
	// To count in code points, we refer to the First Byte column in the Table 3-6.
	//
	// Reference:
	// - [Table 3-6] https://www.unicode.org/versions/Unicode13.0.0/ch03.pdf > Table 3-6.  UTF-8 Bit Distribution
	if b < 128 {
		// 0x0A is LF.
		if b == 0x0A {
			l.state.row++
			l.state.col = 0
		} else {
			l.state.col++
		}
	} else if b>>5 == 6 || b>>4 == 14 || b>>3 == 30 {
		l.state.col++
	}

	return b, false
}

// accept saves the current state.
func (l *Lexer) accept() {
	l.lastAcceptedState = l.state
}

// revert reverts the lexer state to the last accepted state.
//
// We must not call this function consecutively.
func (l *Lexer) revert() {
	l.state = l.lastAcceptedState
}

const (
	ModeIDNil     ModeID = 0
	ModeIDDefault ModeID = 1
)

const (
	ModeNameNil     = ""
	ModeNameDefault = "default"
)

// ModeIDToName converts a mode ID to a name.
func ModeIDToName(id ModeID) string {
	switch id {
	case ModeIDNil:
		return ModeNameNil
	case ModeIDDefault:
		return ModeNameDefault
	}
	return ""
}

const (
	KindIDNil       KindID = 0
	KindIDWs        KindID = 1
	KindIDNl        KindID = 2
	KindIDDef       KindID = 3
	KindIDOr        KindID = 4
	KindIDSemicolon KindID = 5
	KindIDKwData    KindID = 6
	KindIDId        KindID = 7
)

const (
	KindNameNil       = ""
	KindNameWs        = "ws"
	KindNameNl        = "nl"
	KindNameDef       = "def"
	KindNameOr        = "or"
	KindNameSemicolon = "semicolon"
	KindNameKwData    = "kw_data"
	KindNameId        = "id"
)

// KindIDToName converts a kind ID to a name.
func KindIDToName(id KindID) string {
	switch id {
	case KindIDNil:
		return KindNameNil
	case KindIDWs:
		return KindNameWs
	case KindIDNl:
		return KindNameNl
	case KindIDDef:
		return KindNameDef
	case KindIDOr:
		return KindNameOr
	case KindIDSemicolon:
		return KindNameSemicolon
	case KindIDKwData:
		return KindNameKwData
	case KindIDId:
		return KindNameId
	}
	return ""
}

type lexSpec struct {
	pop           [][]bool
	push          [][]ModeID
	modeNames     []string
	initialStates []StateID
	acceptances   [][]ModeKindID
	kindIDs       [][]KindID
	kindNames     []string
	initialModeID ModeID
	modeIDNil     ModeID
	modeKindIDNil ModeKindID
	stateIDNil    StateID

	rowNums           [][]int
	rowDisplacements  [][]int
	bounds            [][]int
	entries           [][]StateID
	originalColCounts []int
}

func NewLexSpec() *lexSpec {
	return &lexSpec{
		pop: [][]bool{
			nil,
			{
				false, false, false, false, false, false, false, false,
			},
		},
		push: [][]ModeID{
			nil,
			{
				0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		modeNames: []string{
			ModeNameNil,
			ModeNameDefault,
		},
		initialStates: []StateID{
			0,
			1,
		},
		acceptances: [][]ModeKindID{
			nil,
			{
				0, 0, 1, 2, 7, 7, 7, 6, 7, 3, 4, 5,
			},
		},
		kindIDs: [][]KindID{
			nil,
			{
				KindIDNil,
				KindIDWs,
				KindIDNl,
				KindIDDef,
				KindIDOr,
				KindIDSemicolon,
				KindIDKwData,
				KindIDId,
			},
		},
		kindNames: []string{
			KindNameNil,
			KindNameWs,
			KindNameNl,
			KindNameDef,
			KindNameOr,
			KindNameSemicolon,
			KindNameKwData,
			KindNameId,
		},
		initialModeID: ModeIDDefault,
		modeIDNil:     ModeIDNil,
		modeKindIDNil: 0,
		stateIDNil:    0,

		rowNums: [][]int{
			nil,
			{
				0, 1, 2, 3, 4, 5, 6, 7, 7, 0, 0, 0,
			},
		},
		rowDisplacements: [][]int{
			nil,
			{
				0, 339, 341, 343, 0, 75, 150, 225,
			},
		},
		bounds: [][]int{
			nil,
			{
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, -1, -1,
				-1, -1, -1, -1, -1, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
				4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, -1, -1, -1, -1, -1, -1, 4, 4, 4,
				4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
				4, 4, 4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, -1, -1, -1, -1, -1, -1, -1,
				5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
				5, 5, 5, 5, 5, 5, -1, -1, -1, -1, -1, -1, 5, 5, 5, 5, 5, 5, 5, 5,
				5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 6, 6,
				6, 6, 6, 6, 6, 6, 6, 6, -1, -1, -1, -1, -1, -1, -1, 6, 6, 6, 6, 6,
				6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
				6, -1, -1, -1, -1, -1, -1, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
				6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 7, 7, 7, 7, 7, 7, 7,
				7, 7, 7, -1, -1, -1, -1, -1, -1, -1, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
				7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, -1, -1, -1, -1,
				-1, -1, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
				7, 7, 7, 7, 7, 7, 7, 7, 1, 1, 2, -1, 1, 3, -1, -1, 3, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 1, -1, 2, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 1, -1,
				1, -1, -1, -1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, -1, -1, -1, -1, -1, -1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, -1, 1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
			},
		},
		entries: [][]StateID{
			nil,
			{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0, 0,
				0, 0, 0, 0, 0, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0, 0, 0, 0, 0, 0, 5, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0, 0, 0, 0, 0, 0, 0,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 0, 0, 0, 0, 0, 0, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 6, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 0, 0, 0, 0, 0, 0, 0, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 0, 0, 0, 0, 0, 0, 7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 0, 0, 0, 0, 0, 0, 0, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0, 0, 0, 0,
				0, 0, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 2, 3, 2, 0, 3, 3, 0, 0, 3, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 2, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 0,
				9, 0, 0, 0, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0, 0, 0, 0, 0, 0, 8, 8, 8, 4,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		originalColCounts: nil,
	}
}

func (s *lexSpec) InitialMode() ModeID {
	return s.initialModeID
}

func (s *lexSpec) Pop(mode ModeID, modeKind ModeKindID) bool {
	return s.pop[mode][modeKind]
}

func (s *lexSpec) Push(mode ModeID, modeKind ModeKindID) (ModeID, bool) {
	id := s.push[mode][modeKind]
	return id, id != s.modeIDNil
}

func (s *lexSpec) ModeName(mode ModeID) string {
	return s.modeNames[mode]
}

func (s *lexSpec) InitialState(mode ModeID) StateID {
	return s.initialStates[mode]
}

func (s *lexSpec) NextState(mode ModeID, state StateID, v int) (StateID, bool) {
	rowNum := s.rowNums[mode][state]
	d := s.rowDisplacements[mode][rowNum]
	if s.bounds[mode][d+v] != rowNum {
		return s.stateIDNil, false
	}
	return s.entries[mode][d+v], true
}

func (s *lexSpec) Accept(mode ModeID, state StateID) (ModeKindID, bool) {
	id := s.acceptances[mode][state]
	return id, id != s.modeKindIDNil
}

func (s *lexSpec) KindIDAndName(mode ModeID, modeKind ModeKindID) (KindID, string) {
	id := s.kindIDs[mode][modeKind]
	return id, s.kindNames[id]
}
