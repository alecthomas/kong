package kong

import (
	"strconv"
	"strings"
)

//go:generate stringer -type=TokenType

// TokenType is the type of a token.
type TokenType int

// Token types.
const (
	UntypedToken TokenType = iota
	EOLToken
	FlagToken               // --<flag>
	FlagValueToken          // =<value>
	ShortFlagToken          // -<short>[<tail]
	ShortFlagTailToken      // <tail>
	PositionalArgumentToken // <arg>
)

// Token created by Scanner.
type Token struct {
	Value string
	Type  TokenType
}

func (t Token) String() string {
	switch t.Type {
	case FlagToken:
		return "--" + t.Value

	case ShortFlagToken:
		return "-" + t.Value

	case EOLToken:
		return "EOL"

	default:
		return strconv.Quote(t.Value)
	}
}

// IsEOL returns true if this Token is past the end of the line.
func (t Token) IsEOL() bool {
	return t.Type == EOLToken
}

// IsAny returns true if the token's type is any of those provided.
func (t Token) IsAny(types ...TokenType) bool {
	for _, typ := range types {
		if t.Type == typ {
			return true
		}
	}
	return false
}

// IsValue returns true if token is usable as a parseable value.
//
// A parseable value is either a value typed token, or an untyped token NOT starting with a hyphen.
func (t Token) IsValue() bool {
	return t.IsAny(FlagValueToken, ShortFlagTailToken, PositionalArgumentToken) ||
		(t.Type == UntypedToken && !strings.HasPrefix(t.Value, "-"))
}

// Scanner is a stack-based scanner over command-line tokens.
//
// Initially all tokens are untyped. As the parser consumes tokens it assigns types, splits tokens, and pushes them back
// onto the stream.
//
// For example, the token "--foo=bar" will be split into the following by the parser:
//
// 		[{FlagToken, "foo"}, {FlagValueToken, "bar"}]
type Scanner struct {
	args []Token
}

// Scan creates a new Scanner from args with untyped tokens.
func Scan(args ...string) *Scanner {
	s := &Scanner{}
	for _, arg := range args {
		s.args = append(s.args, Token{Value: arg})
	}
	return s
}

// ScanFromTokens creates a new Scanner from a slice of tokens.
func ScanFromTokens(tokens ...Token) *Scanner {
	return &Scanner{args: tokens}
}

// Len returns the number of input arguments.
func (s *Scanner) Len() int {
	return len(s.args)
}

// Pop the front token off the Scanner.
func (s *Scanner) Pop() Token {
	if len(s.args) == 0 {
		return Token{Type: EOLToken}
	}
	arg := s.args[0]
	s.args = s.args[1:]
	return arg
}

// PopValue token, or panic with Error.
//
// "context" is used to assist the user if the value can not be popped, eg. "expected <context> value but got <type>"
func (s *Scanner) PopValue(context string) string {
	t := s.Pop()
	if !t.IsValue() {
		fail("expected %s value but got %s", context, t)
	}
	return t.Value
}

// PopWhile predicate returns true.
func (s *Scanner) PopWhile(predicate func(Token) bool) (values []Token) {
	for predicate(s.Peek()) {
		values = append(values, s.Pop())
	}
	return
}

// PopUntil predicate returns true.
func (s *Scanner) PopUntil(predicate func(Token) bool) (values []Token) {
	for !predicate(s.Peek()) {
		values = append(values, s.Pop())
	}
	return
}

// Peek at the next Token or return an EOLToken.
func (s *Scanner) Peek() Token {
	if len(s.args) == 0 {
		return Token{Type: EOLToken}
	}
	return s.args[0]
}

// Push an untyped Token onto the front of the Scanner.
func (s *Scanner) Push(arg string) *Scanner {
	s.PushToken(Token{Value: arg})
	return s
}

// PushTyped pushes a typed token onto the front of the Scanner.
func (s *Scanner) PushTyped(arg string, typ TokenType) *Scanner {
	s.PushToken(Token{Value: arg, Type: typ})
	return s
}

// PushToken pushes a preconstructed Token onto the front of the Scanner.
func (s *Scanner) PushToken(token Token) *Scanner {
	s.args = append([]Token{token}, s.args...)
	return s
}
