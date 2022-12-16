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

	// Row is a row number where a lexeme appears.
	Row int

	// Col is a column number where a lexeme appears.
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

type Lexer struct {
	spec            LexSpec
	src             []byte
	srcPtr          int
	row             int
	col             int
	prevRow         int
	prevCol         int
	tokBuf          []*Token
	modeStack       []ModeID
	passiveModeTran bool
}

// NewLexer returns a new lexer.
func NewLexer(spec LexSpec, src io.Reader, opts ...LexerOption) (*Lexer, error) {
	b, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	l := &Lexer{
		spec:   spec,
		src:    b,
		srcPtr: 0,
		row:    0,
		col:    0,
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
	unfixedBufLen := 0
	row := l.row
	col := l.col
	var tok *Token
	for {
		v, eof := l.read()
		if eof {
			if tok != nil {
				l.unread(unfixedBufLen)
				return tok, nil
			}
			// When `buf` has unaccepted data and reads the EOF, the lexer treats the buffered data as an invalid token.
			if len(buf) > 0 {
				return &Token{
					ModeID:     mode,
					ModeKindID: 0,
					Lexeme:     buf,
					Row:        row,
					Col:        col,
					Invalid:    true,
				}, nil
			}
			return &Token{
				ModeID:     mode,
				ModeKindID: 0,
				Row:        0,
				Col:        0,
				EOF:        true,
			}, nil
		}
		buf = append(buf, v)
		unfixedBufLen++
		nextState, ok := l.spec.NextState(mode, state, int(v))
		if !ok {
			if tok != nil {
				l.unread(unfixedBufLen)
				return tok, nil
			}
			return &Token{
				ModeID:     mode,
				ModeKindID: 0,
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
				Lexeme:     buf,
				Row:        row,
				Col:        col,
			}
			unfixedBufLen = 0
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
	if l.srcPtr >= len(l.src) {
		return 0, true
	}

	b := l.src[l.srcPtr]
	l.srcPtr++

	l.prevRow = l.row
	l.prevCol = l.col

	// Count the token positions.
	// The driver treats LF as the end of lines and counts columns in code points, not bytes.
	// To count in code points, we refer to the First Byte column in the Table 3-6.
	//
	// Reference:
	// - [Table 3-6] https://www.unicode.org/versions/Unicode13.0.0/ch03.pdf > Table 3-6.  UTF-8 Bit Distribution
	if b < 128 {
		// 0x0A is LF.
		if b == 0x0A {
			l.row++
			l.col = 0
		} else {
			l.col++
		}
	} else if b>>5 == 6 || b>>4 == 14 || b>>3 == 30 {
		l.col++
	}

	return b, false
}

// We must not call this function consecutively to record the token position correctly.
func (l *Lexer) unread(n int) {
	l.srcPtr -= n

	l.row = l.prevRow
	l.col = l.prevCol
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
	KindIDNil    KindID = 0
	KindIDWs     KindID = 1
	KindIDNl     KindID = 2
	KindIDDef    KindID = 3
	KindIDColon  KindID = 4
	KindIDOr     KindID = 5
	KindIDAdd    KindID = 6
	KindIDSub    KindID = 7
	KindIDMul    KindID = 8
	KindIDDiv    KindID = 9
	KindIDMod    KindID = 10
	KindIDKwData KindID = 11
	KindIDId     KindID = 12
	KindIDInt    KindID = 13
	KindIDString KindID = 14
)

const (
	KindNameNil    = ""
	KindNameWs     = "ws"
	KindNameNl     = "nl"
	KindNameDef    = "def"
	KindNameColon  = "colon"
	KindNameOr     = "or"
	KindNameAdd    = "add"
	KindNameSub    = "sub"
	KindNameMul    = "mul"
	KindNameDiv    = "div"
	KindNameMod    = "mod"
	KindNameKwData = "kw_data"
	KindNameId     = "id"
	KindNameInt    = "int"
	KindNameString = "string"
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
	case KindIDColon:
		return KindNameColon
	case KindIDOr:
		return KindNameOr
	case KindIDAdd:
		return KindNameAdd
	case KindIDSub:
		return KindNameSub
	case KindIDMul:
		return KindNameMul
	case KindIDDiv:
		return KindNameDiv
	case KindIDMod:
		return KindNameMod
	case KindIDKwData:
		return KindNameKwData
	case KindIDId:
		return KindNameId
	case KindIDInt:
		return KindNameInt
	case KindIDString:
		return KindNameString
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
				false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
			},
		},
		push: [][]ModeID{
			nil,
			{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
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
				0, 0, 1, 2, 12, 12, 12, 11, 12, 13, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 4, 5, 6, 7, 8, 9, 10, 14,
			},
		},
		kindIDs: [][]KindID{
			nil,
			{
				KindIDNil,
				KindIDWs,
				KindIDNl,
				KindIDDef,
				KindIDColon,
				KindIDOr,
				KindIDAdd,
				KindIDSub,
				KindIDMul,
				KindIDDiv,
				KindIDMod,
				KindIDKwData,
				KindIDId,
				KindIDInt,
				KindIDString,
			},
		},
		kindNames: []string{
			KindNameNil,
			KindNameWs,
			KindNameNl,
			KindNameDef,
			KindNameColon,
			KindNameOr,
			KindNameAdd,
			KindNameSub,
			KindNameMul,
			KindNameDiv,
			KindNameMod,
			KindNameKwData,
			KindNameId,
			KindNameInt,
			KindNameString,
		},
		initialModeID: ModeIDDefault,
		modeIDNil:     ModeIDNil,
		modeKindIDNil: 0,
		stateIDNil:    0,

		rowNums: [][]int{
			nil,
			{
				0, 1, 2, 3, 4, 5, 6, 7, 7, 8, 9, 10, 11, 10, 12, 10, 13, 10, 14, 10,
				15, 16, 10, 17, 18, 10, 19, 20, 10, 21, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		rowDisplacements: [][]int{
			nil,
			{
				0, 236, 1264, 1265, 765, 840, 915, 990, 1215, 0, 237, 1023, 301, 1087, 365, 991, 429, 493, 557, 1119,
				621, 1266,
			},
		},
		bounds: [][]int{
			nil,
			{
				9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
				9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
				9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
				9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
				9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
				9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
				9, 9, 9, 9, 9, 9, 9, 9, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 9, 9, 9, 9, 9, 9,
				9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
				9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
				9, 9, 9, 9, 9, 1, 1, -1, -1, 1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, 1, -1, 1, -1, -1, 1, -1, -1, -1, -1, 1, 1,
				-1, 1, -1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, -1, -1, 1, -1, -1,
				-1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, -1, -1, -1, -1, -1, -1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, -1,
				1, -1, -1, -1, -1, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 10, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12,
				12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12,
				12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12,
				12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 14, 14, 14, 14, 14, 14, 14,
				14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14,
				14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14,
				14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 16, 16, 16,
				16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,
				16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,
				16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16,
				16, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17,
				17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17,
				17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17,
				17, 17, 17, 17, 17, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18,
				18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18,
				18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18,
				18, 18, 18, 18, 18, 18, 18, 18, 18, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20,
				20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20,
				20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20,
				20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 4, 4, 4, 4, 4, 4, 4,
				4, 4, 4, -1, -1, -1, -1, -1, -1, -1, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
				4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, -1, -1, -1, -1,
				-1, -1, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
				4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, -1, -1,
				-1, -1, -1, -1, -1, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
				5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, -1, -1, -1, -1, -1, -1, 5, 5, 5,
				5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
				5, 5, 5, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, -1, -1, -1, -1, -1, -1, -1,
				6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
				6, 6, 6, 6, 6, 6, -1, -1, -1, -1, -1, -1, 6, 6, 6, 6, 6, 6, 6, 6,
				6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 7, 7,
				7, 7, 7, 7, 7, 7, 7, 7, -1, -1, -1, -1, -1, -1, -1, 7, 7, 7, 7, 7,
				7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
				7, -1, -1, -1, -1, -1, -1, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
				7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 15, 15, 15, 15, 15,
				15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
				15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
				15, 15, 15, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11,
				11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 13, 13, 13, 13, 13,
				13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13,
				13, 13, 13, 13, 13, 13, 13, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19,
				19, 19, 19, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 2, -1, 3, -1, -1, 3, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 2, -1, -1, -1,
				21, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 21, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
				-1, -1,
			},
		},
		entries: [][]StateID{
			nil,
			{
				10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 38, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 29, 10, 10, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 11, 11, 11, 11, 11,
				11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11,
				11, 11, 11, 11, 12, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 16, 18, 18,
				20, 23, 23, 23, 26, 2, 3, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 10, 0, 0, 37, 0, 0, 0, 0, 35, 33,
				0, 34, 0, 36, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 31, 0, 0, 30, 0, 0,
				0, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 0, 0, 0, 0, 0, 0, 8, 8, 8, 4, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0,
				32, 0, 0, 0, 0, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
				10, 10, 10, 10, 10, 10, 10, 10, 10, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
				15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
				15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15,
				15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 19, 19, 19, 19, 19, 19, 19,
				19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19,
				19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19,
				19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 22, 22, 22,
				22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
				22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
				22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
				22, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
				24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
				24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
				24, 24, 24, 24, 24, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25,
				25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25,
				25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25,
				25, 25, 25, 25, 25, 25, 25, 25, 25, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28,
				28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28,
				28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28,
				28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 0, 0, 0, 0, 0, 0, 0, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0, 0, 0, 0,
				0, 0, 5, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0, 0,
				0, 0, 0, 0, 0, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0, 0, 0, 0, 0, 0, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 6, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0, 0, 0, 0, 0, 0, 0,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 0, 0, 0, 0, 0, 0, 7, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 0, 0, 0, 0, 0, 0, 0, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 0, 0, 0, 0, 0, 0, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 21, 21, 21, 21, 21,
				21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
				21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
				21, 21, 21, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13,
				13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 17, 17, 17, 17, 17,
				17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17,
				17, 17, 17, 17, 17, 17, 17, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27,
				27, 27, 27, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 2, 0, 3, 0, 0, 3, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0,
				10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0,
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
