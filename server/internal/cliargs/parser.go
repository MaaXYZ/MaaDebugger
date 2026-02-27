package cliargs

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type ValueParser func(raw string) (any, error)

type Option struct {
	Name        string
	Short       string
	Help        string
	Required    bool
	Default     any
	Parser      ValueParser
	TakesValue  bool
	Placeholder string
}

type Parser struct {
	Program     string
	Description string

	options    []*Option
	longIndex  map[string]*Option
	shortIndex map[string]*Option
}

type Result struct {
	Values        map[string]any
	Positionals   []string
	HelpRequested bool
}

func New(program, description string) *Parser {
	p := &Parser{
		Program:     program,
		Description: description,
		longIndex:   make(map[string]*Option),
		shortIndex:  make(map[string]*Option),
	}

	p.AddOption(Option{
		Name:        "help",
		Short:       "h",
		Help:        "show help message",
		Default:     false,
		TakesValue:  false,
		Placeholder: "BOOL",
		Parser:      BoolParser(),
	})

	return p
}

func (p *Parser) AddOption(opt Option) {
	if opt.Name == "" {
		panic("option name cannot be empty")
	}
	if strings.Contains(opt.Name, "=") {
		panic("option name cannot contain '='")
	}
	if opt.Short != "" {
		if len(opt.Short) != 1 {
			panic("short option must be exactly one character")
		}
		if strings.Contains(opt.Short, "-") {
			panic("short option cannot contain '-' ")
		}
	}
	if opt.Parser == nil {
		panic("option parser cannot be nil")
	}
	if opt.Placeholder == "" {
		opt.Placeholder = "VALUE"
	}

	if _, exists := p.longIndex[opt.Name]; exists {
		panic("duplicate long option: --" + opt.Name)
	}
	if opt.Short != "" {
		if _, exists := p.shortIndex[opt.Short]; exists {
			panic("duplicate short option: -" + opt.Short)
		}
	}

	optCopy := opt
	p.options = append(p.options, &optCopy)
	p.longIndex[opt.Name] = &optCopy
	if opt.Short != "" {
		p.shortIndex[opt.Short] = &optCopy
	}
}

func (p *Parser) AddString(name, short, help, def string, required bool) {
	p.AddOption(Option{
		Name:        name,
		Short:       short,
		Help:        help,
		Required:    required,
		Default:     def,
		Parser:      StringParser(),
		TakesValue:  true,
		Placeholder: "STRING",
	})
}

func (p *Parser) AddInt(name, short, help string, def int, required bool) {
	p.AddOption(Option{
		Name:        name,
		Short:       short,
		Help:        help,
		Required:    required,
		Default:     def,
		Parser:      IntParser(),
		TakesValue:  true,
		Placeholder: "INT",
	})
}

func (p *Parser) AddBool(name, short, help string, def bool) {
	p.AddOption(Option{
		Name:        name,
		Short:       short,
		Help:        help,
		Required:    false,
		Default:     def,
		Parser:      BoolParser(),
		TakesValue:  false,
		Placeholder: "BOOL",
	})
}

func (p *Parser) Parse(args []string) (*Result, error) {
	res := &Result{
		Values:      make(map[string]any),
		Positionals: make([]string, 0),
	}

	for _, opt := range p.options {
		if opt.Default != nil {
			res.Values[opt.Name] = opt.Default
		}
	}

	for i := 0; i < len(args); i++ {
		tok := args[i]

		if tok == "--" {
			res.Positionals = append(res.Positionals, args[i+1:]...)
			break
		}

		if !strings.HasPrefix(tok, "-") || tok == "-" {
			res.Positionals = append(res.Positionals, tok)
			continue
		}

		if strings.HasPrefix(tok, "--") {
			name, rawValue, hasInlineValue := splitLongToken(tok[2:])
			opt, ok := p.longIndex[name]
			if !ok {
				return nil, fmt.Errorf("unknown option: --%s", name)
			}

			if !opt.TakesValue {
				if hasInlineValue {
					parsed, err := opt.Parser(rawValue)
					if err != nil {
						return nil, fmt.Errorf("invalid value for --%s: %w", name, err)
					}
					res.Values[opt.Name] = parsed
				} else {
					res.Values[opt.Name] = true
				}
				continue
			}

			var valueToken string
			if hasInlineValue {
				valueToken = rawValue
			} else {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("option --%s requires a value", name)
				}
				i++
				valueToken = args[i]
			}

			parsed, err := opt.Parser(valueToken)
			if err != nil {
				return nil, fmt.Errorf("invalid value for --%s: %w", name, err)
			}
			res.Values[opt.Name] = parsed
			continue
		}

		shortCluster := tok[1:]
		for idx := 0; idx < len(shortCluster); idx++ {
			key := string(shortCluster[idx])
			opt, ok := p.shortIndex[key]
			if !ok {
				return nil, fmt.Errorf("unknown option: -%s", key)
			}

			if !opt.TakesValue {
				res.Values[opt.Name] = true
				continue
			}

			var valueToken string
			if idx < len(shortCluster)-1 {
				valueToken = shortCluster[idx+1:]
				idx = len(shortCluster)
			} else {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("option -%s requires a value", key)
				}
				i++
				valueToken = args[i]
			}

			parsed, err := opt.Parser(valueToken)
			if err != nil {
				return nil, fmt.Errorf("invalid value for -%s: %w", key, err)
			}
			res.Values[opt.Name] = parsed
			break
		}
	}

	if hv, ok := res.Values["help"].(bool); ok && hv {
		res.HelpRequested = true
		return res, nil
	}

	for _, opt := range p.options {
		if !opt.Required {
			continue
		}
		if _, ok := res.Values[opt.Name]; !ok {
			return nil, fmt.Errorf("required option missing: --%s", opt.Name)
		}
	}

	return res, nil
}

func (p *Parser) Help() string {
	var b strings.Builder

	if p.Program == "" {
		p.Program = "app"
	}

	b.WriteString("Usage:\n")
	b.WriteString("  ")
	b.WriteString(p.Program)
	b.WriteString(" [options]\n")

	if p.Description != "" {
		b.WriteString("\n")
		b.WriteString(p.Description)
		b.WriteString("\n")
	}

	b.WriteString("\nOptions:\n")

	sorted := make([]*Option, 0, len(p.options))
	sorted = append(sorted, p.options...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	for _, opt := range sorted {
		left := buildOptionSignature(opt)
		line := fmt.Sprintf("  %-26s %s", left, opt.Help)
		if opt.Required {
			line += " (required)"
		}
		if opt.Default != nil {
			line += fmt.Sprintf(" [default: %v]", opt.Default)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func (r *Result) String(name string) (string, bool) {
	v, ok := r.Values[name]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func (r *Result) Int(name string) (int, bool) {
	v, ok := r.Values[name]
	if !ok {
		return 0, false
	}
	n, ok := v.(int)
	return n, ok
}

func (r *Result) Bool(name string) (bool, bool) {
	v, ok := r.Values[name]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func StringParser() ValueParser {
	return func(raw string) (any, error) {
		return raw, nil
	}
}

func IntParser() ValueParser {
	return func(raw string) (any, error) {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("expected integer, got %q", raw)
		}
		return n, nil
	}
}

func BoolParser() ValueParser {
	return func(raw string) (any, error) {
		v := strings.TrimSpace(strings.ToLower(raw))
		switch v {
		case "1", "true", "yes", "y", "on":
			return true, nil
		case "0", "false", "no", "n", "off":
			return false, nil
		default:
			return nil, fmt.Errorf("expected boolean, got %q", raw)
		}
	}
}

func splitLongToken(raw string) (name string, value string, hasValue bool) {
	parts := strings.SplitN(raw, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return raw, "", false
}

func buildOptionSignature(opt *Option) string {
	parts := make([]string, 0, 2)
	if opt.Short != "" {
		parts = append(parts, "-"+opt.Short)
	}
	long := "--" + opt.Name
	if opt.TakesValue {
		long += " <" + opt.Placeholder + ">"
	}
	parts = append(parts, long)
	return strings.Join(parts, ", ")
}
