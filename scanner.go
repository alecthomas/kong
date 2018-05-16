package kong

import (
	"fmt"
	"strconv"
)

//go:generate stringer -type=TokenType

type TokenType int

const (
	UntypedToken TokenType = iota
	EOLToken
	FlagToken               // --<flag>
	FlagValueToken          // =<value>
	ShortFlagToken          // -<short>[<tail]
	ShortFlagTailToken      // <tail>
	PositionalArgumentToken // <arg>
)

type TokenAssertionError struct{ err error }

func (t TokenAssertionError) Error() string {
	return t.err.Error()
}

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

func (t Token) IsAny(types ...TokenType) bool {
	for _, typ := range types {
		if t.Type == typ {
			return true
		}
	}
	return false
}

func (t Token) IsValue() bool {
	return t.IsAny(FlagValueToken, ShortFlagTailToken, PositionalArgumentToken, UntypedToken)
}

type Scanner struct {
	raw  []string
	args []Token
}

func Scan(args ...string) *Scanner {
	s := &Scanner{raw: args}
	for _, arg := range args {
		s.args = append(s.args, Token{Value: arg})
	}
	return s
}

func (s *Scanner) Pop() Token {
	if len(s.args) == 0 {
		return Token{Type: EOLToken}
	}
	arg := s.args[0]
	s.args = s.args[1:]
	return arg
}

// PopValue or panic with TokenAssertionError.
func (s *Scanner) PopValue(context string) string {
	t := s.Pop()
	if !t.IsValue() {
		panic(TokenAssertionError{fmt.Errorf("expected %s value but got %s", context, t)})
	}
	return t.Value
}

func (s *Scanner) Peek() Token {
	if len(s.args) == 0 {
		return Token{Type: EOLToken}
	}
	return s.args[0]
}

func (s *Scanner) Push(arg string) {
	s.PushToken(Token{Value: arg})
}

func (s *Scanner) PushTyped(arg string, typ TokenType) {
	s.PushToken(Token{Value: arg, Type: typ})
}

func (s *Scanner) PushToken(token Token) {
	s.args = append([]Token{token}, s.args...)
}
