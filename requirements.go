package kong

import (
	"fmt"
	"strings"
	"unicode"
)

type flagRequirement struct {
	owner      *Flag
	node       *Node
	text       string
	expression *requirementExpr
}

type requirementExpr struct {
	operator rune
	flag     *Flag
	left     *requirementExpr
	right    *requirementExpr
}

func (e *requirementExpr) evaluate(present map[*Flag]bool) bool {
	switch e.operator {
	case '!':
		return !e.left.evaluate(present)
	case '&':
		return e.left.evaluate(present) && e.right.evaluate(present)
	case '|':
		return e.left.evaluate(present) || e.right.evaluate(present)
	default:
		return present[e.flag]
	}
}

type requirementParser struct {
	input    []rune
	position int
	flags    map[string]*Flag
}

func (p *requirementParser) peek() rune {
	for p.position < len(p.input) && unicode.IsSpace(p.input[p.position]) {
		p.position++
	}
	if p.position == len(p.input) {
		return -1
	}
	return p.input[p.position]
}

func (p *requirementParser) expression() (*requirementExpr, error) {
	left, err := p.conjunction()
	if err != nil {
		return nil, err
	}
	for p.peek() == '|' {
		p.position++
		right, err := p.conjunction()
		if err != nil {
			return nil, err
		}
		left = &requirementExpr{operator: '|', left: left, right: right}
	}
	return left, nil
}

func (p *requirementParser) conjunction() (*requirementExpr, error) {
	left, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.peek() == '&' {
		p.position++
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		left = &requirementExpr{operator: '&', left: left, right: right}
	}
	return left, nil
}

func (p *requirementParser) unary() (*requirementExpr, error) {
	switch p.peek() {
	case '!':
		p.position++
		operand, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &requirementExpr{operator: '!', left: operand}, nil
	case '(':
		p.position++
		expression, err := p.expression()
		if err != nil {
			return nil, err
		}
		if p.peek() != ')' {
			return nil, fmt.Errorf("expected ')' at position %d", p.position)
		}
		p.position++
		return expression, nil
	case -1, ')', '&', '|':
		return nil, fmt.Errorf("expected flag name at position %d", p.position)
	}
	start := p.position
	for p.position < len(p.input) {
		r := p.input[p.position]
		if unicode.IsSpace(r) || strings.ContainsRune("!&|()", r) {
			break
		}
		p.position++
	}
	name := string(p.input[start:p.position])
	flag := p.flags[name]
	if flag == nil {
		return nil, fmt.Errorf("unknown flag %q in requirement", name)
	}
	return &requirementExpr{flag: flag}, nil
}

func (k *Kong) buildRequirements() error {
	return Visit(k.Model.Node, func(value Visitable, next Next) error {
		switch value := value.(type) {
		case *Node:
			if value.Tag != nil && value.Tag.Has("requires") {
				return fmt.Errorf("%s: requires is only valid on flags", value.Path())
			}
			if err := k.nodeRequirements(value); err != nil {
				return err
			}
		case *Value:
			if value.Flag == nil && value.Tag.Has("requires") {
				return fmt.Errorf("%s: requires is only valid on flags", value.Name)
			}
		}
		return next(nil)
	})
}

func (k *Kong) nodeRequirements(node *Node) error {
	visible := map[string]*Flag{}
	for scope := node; scope != nil; scope = scope.Parent {
		for _, flag := range scope.Flags {
			if visible[flag.Name] == nil {
				visible[flag.Name] = flag
			}
		}
	}
	for _, flag := range node.Flags {
		for _, text := range flag.Tag.GetAll("requires") {
			parser := requirementParser{input: []rune(text), flags: visible}
			expression, err := parser.expression()
			if err == nil && parser.peek() != -1 {
				err = fmt.Errorf("unexpected token at position %d", parser.position)
			}
			if err != nil {
				return fmt.Errorf("--%s requires %q: %w", flag.Name, text, err)
			}
			k.requirements = append(k.requirements, flagRequirement{
				owner: flag, node: node, text: text, expression: expression,
			})
		}
	}
	return nil
}

func (c *Context) validateRequirements() error {
	present := map[*Flag]bool{}
	for flag := range c.requirementEnv {
		present[flag] = true
	}
	active := map[*Node]bool{}
	for _, path := range c.Path {
		if path.Flag != nil {
			present[path.Flag] = true
		}
		if node := path.Node(); node != nil {
			active[node] = true
		}
	}
	for _, requirement := range c.requirements {
		if active[requirement.node] && present[requirement.owner] &&
			!requirement.expression.evaluate(present) {
			return fmt.Errorf("--%s requires %s", requirement.owner.Name, requirement.text)
		}
	}
	return nil
}

func (c *Context) recordRequirementEnvironment(value *Value) {
	if value.Flag == nil || !atLeastOneEnvSet(value.Tag.Envs) {
		return
	}
	if c.requirementEnv == nil {
		c.requirementEnv = map[*Flag]bool{}
	}
	c.requirementEnv[value.Flag] = true
}
