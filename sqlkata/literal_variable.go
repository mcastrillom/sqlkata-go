package sqlkata

// UnsafeLiteral is embedded raw SQL (SqlKata.UnsafeLiteral). Do not pass user input.
type UnsafeLiteral struct {
	Value string
}

// NewUnsafeLiteral mirrors new UnsafeLiteral(value, replaceQuotes).
func NewUnsafeLiteral(value string, replaceQuotes bool) *UnsafeLiteral {
	if replaceQuotes {
		value = replaceQuotesSingle(value)
	}
	return &UnsafeLiteral{Value: value}
}

func replaceQuotesSingle(s string) string {
	out := make([]byte, 0, len(s)+8)
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			out = append(out, '\'', '\'')
		} else {
			out = append(out, s[i])
		}
	}
	return string(out)
}

// Variable references a value defined via Query.Define (SqlKata.Variable).
type Variable struct {
	Name string
}

// NewVariable mirrors new Variable(name).
func NewVariable(name string) *Variable {
	return &Variable{Name: name}
}

// VariableExpr / UnsafeLiteralExpr mirror SqlKata.Expressions helpers.
func VariableExpr(name string) *Variable { return NewVariable(name) }

func UnsafeLiteralExpr(value string, replaceQuotes bool) *UnsafeLiteral {
	return NewUnsafeLiteral(value, replaceQuotes)
}
