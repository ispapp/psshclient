package codeditor

import (
	"regexp"
	"unicode"
)

// Lexer handles tokenization of source code for syntax highlighting
type Lexer struct {
	language string
	rules    []LexerRule
}

// LexerRule defines a pattern and token type for lexical analysis
type LexerRule struct {
	Pattern   *regexp.Regexp
	TokenType TokenType
	Priority  int // Higher priority rules are checked first
}

// NewLexer creates a new lexer for the specified language
func NewLexer(language string) *Lexer {
	lexer := &Lexer{
		language: language,
		rules:    []LexerRule{},
	}

	lexer.initializeRules()
	return lexer
}

// initializeRules sets up the lexer rules based on the language
func (l *Lexer) initializeRules() {
	switch l.language {
	case "go":
		l.initializeGoRules()
	case "python":
		l.initializePythonRules()
	case "javascript", "js":
		l.initializeJavaScriptRules()
	case "java":
		l.initializeJavaRules()
	case "c", "cpp", "c++":
		l.initializeCRules()
	default:
		l.initializeGenericRules()
	}
}

// initializeGoRules sets up syntax highlighting rules for Go
func (l *Lexer) initializeGoRules() {
	// Comments (higher priority)
	l.addRule(`//.*`, TokenComment, 10)
	l.addRule(`/\*[\s\S]*?\*/`, TokenComment, 10)

	// Strings
	l.addRule(`"(?:[^"\\]|\\.)*"`, TokenString, 9)
	l.addRule("`[^`]*`", TokenString, 9)
	l.addRule(`'(?:[^'\\]|\\.)*'`, TokenString, 9)

	// Numbers
	l.addRule(`\b\d+\.?\d*([eE][+-]?\d+)?\b`, TokenNumber, 8)
	l.addRule(`\b0[xX][0-9a-fA-F]+\b`, TokenNumber, 8)

	// Keywords
	keywords := []string{
		"break", "case", "chan", "const", "continue", "default", "defer",
		"else", "fallthrough", "for", "func", "go", "goto", "if", "import",
		"interface", "map", "package", "range", "return", "select", "struct",
		"switch", "type", "var",
	}
	for _, keyword := range keywords {
		l.addRule(`\b`+keyword+`\b`, TokenKeyword, 7)
	}

	// Types
	types := []string{
		"bool", "byte", "complex64", "complex128", "error", "float32", "float64",
		"int", "int8", "int16", "int32", "int64", "rune", "string",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
	}
	for _, t := range types {
		l.addRule(`\b`+t+`\b`, TokenType_, 6)
	}

	// Functions (identifiers followed by parentheses) - simplified without lookahead
	l.addRule(`\b[a-zA-Z_][a-zA-Z0-9_]*\s*\(`, TokenFunction, 5)

	// Operators
	l.addRule(`[+\-*/%=!<>&|^~]+`, TokenOperator, 4)
	l.addRule(`[{}()\[\];,.]`, TokenOperator, 4)

	// Identifiers
	l.addRule(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`, TokenIdentifier, 1)
}

// initializePythonRules sets up syntax highlighting rules for Python
func (l *Lexer) initializePythonRules() {
	// Comments
	l.addRule(`#.*`, TokenComment, 10)

	// Strings
	l.addRule(`"""[\s\S]*?"""`, TokenString, 9)
	l.addRule(`'''[\s\S]*?'''`, TokenString, 9)
	l.addRule(`"(?:[^"\\]|\\.)*"`, TokenString, 9)
	l.addRule(`'(?:[^'\\]|\\.)*'`, TokenString, 9)

	// Numbers
	l.addRule(`\b\d+\.?\d*([eE][+-]?\d+)?\b`, TokenNumber, 8)
	l.addRule(`\b0[xX][0-9a-fA-F]+\b`, TokenNumber, 8)

	// Keywords
	keywords := []string{
		"and", "as", "assert", "break", "class", "continue", "def", "del",
		"elif", "else", "except", "finally", "for", "from", "global", "if",
		"import", "in", "is", "lambda", "nonlocal", "not", "or", "pass",
		"raise", "return", "try", "while", "with", "yield",
	}
	for _, keyword := range keywords {
		l.addRule(`\b`+keyword+`\b`, TokenKeyword, 7)
	}

	// Built-in types and functions
	builtins := []string{
		"bool", "int", "float", "str", "list", "dict", "tuple", "set",
		"print", "len", "range", "enumerate", "zip", "map", "filter",
	}
	for _, builtin := range builtins {
		l.addRule(`\b`+builtin+`\b`, TokenType_, 6)
	}

	// Functions
	l.addRule(`\b[a-zA-Z_][a-zA-Z0-9_]*\s*\(`, TokenFunction, 5)

	// Operators
	l.addRule(`[+\-*/%=!<>&|^~]+`, TokenOperator, 4)
	l.addRule(`[{}()\[\];,.]`, TokenOperator, 4)

	// Identifiers
	l.addRule(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`, TokenIdentifier, 1)
}

// initializeJavaScriptRules sets up syntax highlighting rules for JavaScript
func (l *Lexer) initializeJavaScriptRules() {
	// Comments
	l.addRule(`//.*`, TokenComment, 10)
	l.addRule(`/\*[\s\S]*?\*/`, TokenComment, 10)

	// Strings
	l.addRule(`"(?:[^"\\]|\\.)*"`, TokenString, 9)
	l.addRule(`'(?:[^'\\]|\\.)*'`, TokenString, 9)
	l.addRule("`(?:[^`\\]|\\.)*`", TokenString, 9)

	// Numbers
	l.addRule(`\b\d+\.?\d*([eE][+-]?\d+)?\b`, TokenNumber, 8)
	l.addRule(`\b0[xX][0-9a-fA-F]+\b`, TokenNumber, 8)

	// Keywords
	keywords := []string{
		"break", "case", "catch", "class", "const", "continue", "debugger",
		"default", "delete", "do", "else", "export", "extends", "finally",
		"for", "function", "if", "import", "in", "instanceof", "new",
		"return", "super", "switch", "this", "throw", "try", "typeof",
		"var", "void", "while", "with", "yield", "let", "async", "await",
	}
	for _, keyword := range keywords {
		l.addRule(`\b`+keyword+`\b`, TokenKeyword, 7)
	}

	// Built-in types and objects
	builtins := []string{
		"Array", "Object", "String", "Number", "Boolean", "Date", "RegExp",
		"Math", "JSON", "console", "window", "document",
	}
	for _, builtin := range builtins {
		l.addRule(`\b`+builtin+`\b`, TokenType_, 6)
	}

	// Functions
	l.addRule(`\b[a-zA-Z_$][a-zA-Z0-9_$]*\s*\(`, TokenFunction, 5)

	// Operators
	l.addRule(`[+\-*/%=!<>&|^~?:]+`, TokenOperator, 4)
	l.addRule(`[{}()\[\];,.]`, TokenOperator, 4)

	// Identifiers
	l.addRule(`\b[a-zA-Z_$][a-zA-Z0-9_$]*\b`, TokenIdentifier, 1)
}

// initializeJavaRules sets up syntax highlighting rules for Java
func (l *Lexer) initializeJavaRules() {
	// Comments
	l.addRule(`//.*`, TokenComment, 10)
	l.addRule(`/\*[\s\S]*?\*/`, TokenComment, 10)

	// Strings
	l.addRule(`"(?:[^"\\]|\\.)*"`, TokenString, 9)
	l.addRule(`'(?:[^'\\]|\\.)*'`, TokenString, 9)

	// Numbers
	l.addRule(`\b\d+\.?\d*([eE][+-]?\d+)?[fFdD]?\b`, TokenNumber, 8)
	l.addRule(`\b0[xX][0-9a-fA-F]+[lL]?\b`, TokenNumber, 8)

	// Keywords
	keywords := []string{
		"abstract", "assert", "boolean", "break", "byte", "case", "catch",
		"char", "class", "const", "continue", "default", "do", "double",
		"else", "enum", "extends", "final", "finally", "float", "for",
		"goto", "if", "implements", "import", "instanceof", "int", "interface",
		"long", "native", "new", "package", "private", "protected", "public",
		"return", "short", "static", "strictfp", "super", "switch", "synchronized",
		"this", "throw", "throws", "transient", "try", "void", "volatile", "while",
	}
	for _, keyword := range keywords {
		l.addRule(`\b`+keyword+`\b`, TokenKeyword, 7)
	}

	// Types
	types := []string{
		"String", "Integer", "Boolean", "Character", "Byte", "Short",
		"Long", "Float", "Double", "Object", "List", "Map", "Set",
	}
	for _, t := range types {
		l.addRule(`\b`+t+`\b`, TokenType_, 6)
	}

	// Functions
	l.addRule(`\b[a-zA-Z_][a-zA-Z0-9_]*\s*\(`, TokenFunction, 5)

	// Operators
	l.addRule(`[+\-*/%=!<>&|^~?:]+`, TokenOperator, 4)
	l.addRule(`[{}()\[\];,.]`, TokenOperator, 4)

	// Identifiers
	l.addRule(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`, TokenIdentifier, 1)
}

// initializeCRules sets up syntax highlighting rules for C/C++
func (l *Lexer) initializeCRules() {
	// Comments
	l.addRule(`//.*`, TokenComment, 10)
	l.addRule(`/\*[\s\S]*?\*/`, TokenComment, 10)

	// Preprocessor directives
	l.addRule(`#.*`, TokenKeyword, 10)

	// Strings
	l.addRule(`"(?:[^"\\]|\\.)*"`, TokenString, 9)
	l.addRule(`'(?:[^'\\]|\\.)*'`, TokenString, 9)

	// Numbers
	l.addRule(`\b\d+\.?\d*([eE][+-]?\d+)?[fFlL]?\b`, TokenNumber, 8)
	l.addRule(`\b0[xX][0-9a-fA-F]+[lL]?\b`, TokenNumber, 8)

	// Keywords
	keywords := []string{
		"auto", "break", "case", "char", "const", "continue", "default", "do",
		"double", "else", "enum", "extern", "float", "for", "goto", "if",
		"int", "long", "register", "return", "short", "signed", "sizeof",
		"static", "struct", "switch", "typedef", "union", "unsigned", "void",
		"volatile", "while",
	}

	// C++ specific keywords
	if l.language == "cpp" || l.language == "c++" {
		cppKeywords := []string{
			"class", "private", "protected", "public", "virtual", "template",
			"typename", "namespace", "using", "new", "delete", "try", "catch",
			"throw", "bool", "true", "false", "inline", "operator", "friend",
		}
		keywords = append(keywords, cppKeywords...)
	}

	for _, keyword := range keywords {
		l.addRule(`\b`+keyword+`\b`, TokenKeyword, 7)
	}

	// Functions
	l.addRule(`\b[a-zA-Z_][a-zA-Z0-9_]*\s*\(`, TokenFunction, 5)

	// Operators
	l.addRule(`[+\-*/%=!<>&|^~?:]+`, TokenOperator, 4)
	l.addRule(`[{}()\[\];,.]`, TokenOperator, 4)

	// Identifiers
	l.addRule(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`, TokenIdentifier, 1)
}

// initializeGenericRules sets up basic syntax highlighting for unknown languages
func (l *Lexer) initializeGenericRules() {
	// Comments
	l.addRule(`//.*`, TokenComment, 10)
	l.addRule(`/\*[\s\S]*?\*/`, TokenComment, 10)
	l.addRule(`#.*`, TokenComment, 10)

	// Strings
	l.addRule(`"(?:[^"\\]|\\.)*"`, TokenString, 9)
	l.addRule(`'(?:[^'\\]|\\.)*'`, TokenString, 9)

	// Numbers
	l.addRule(`\b\d+\.?\d*\b`, TokenNumber, 8)

	// Operators
	l.addRule(`[+\-*/%=!<>&|^~]+`, TokenOperator, 4)
	l.addRule(`[{}()\[\];,.]`, TokenOperator, 4)

	// Identifiers
	l.addRule(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`, TokenIdentifier, 1)
}

// addRule adds a lexer rule with the given pattern and token type
func (l *Lexer) addRule(pattern string, tokenType TokenType, priority int) {
	regex := regexp.MustCompile(pattern)
	rule := LexerRule{
		Pattern:   regex,
		TokenType: tokenType,
		Priority:  priority,
	}

	// Insert rule maintaining priority order (higher priority first)
	inserted := false
	for i, existingRule := range l.rules {
		if priority > existingRule.Priority {
			l.rules = append(l.rules[:i], append([]LexerRule{rule}, l.rules[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		l.rules = append(l.rules, rule)
	}
}

// Tokenize breaks down the input text into tokens based on the language rules
func (l *Lexer) Tokenize(content string) []Token {
	var tokens []Token
	position := 0

	for position < len(content) {
		// Skip whitespace but preserve it for positioning
		if unicode.IsSpace(rune(content[position])) {
			start := position
			for position < len(content) && unicode.IsSpace(rune(content[position])) {
				position++
			}

			// Add whitespace as plain text to preserve formatting
			tokens = append(tokens, Token{
				Type:  TokenPlain,
				Value: content[start:position],
				Start: start,
				End:   position,
			})
			continue
		}

		// Try to match against rules
		matched := false
		for _, rule := range l.rules {
			match := rule.Pattern.FindStringIndex(content[position:])
			if match != nil && match[0] == 0 { // Match at current position
				start := position
				end := position + match[1]

				tokens = append(tokens, Token{
					Type:  rule.TokenType,
					Value: content[start:end],
					Start: start,
					End:   end,
				})

				position = end
				matched = true
				break
			}
		}

		// If no rule matched, treat as plain text
		if !matched {
			start := position
			position++

			tokens = append(tokens, Token{
				Type:  TokenPlain,
				Value: content[start:position],
				Start: start,
				End:   position,
			})
		}
	}

	return tokens
}
