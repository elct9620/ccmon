package entity

// Token represents token usage for an API request
type Token struct {
	input         int64
	output        int64
	cacheRead     int64
	cacheCreation int64
}

// NewToken creates a new Token value object
func NewToken(input, output, cacheRead, cacheCreation int64) Token {
	return Token{
		input:         input,
		output:        output,
		cacheRead:     cacheRead,
		cacheCreation: cacheCreation,
	}
}

// Input returns the number of input tokens
func (t Token) Input() int64 {
	return t.input
}

// Output returns the number of output tokens
func (t Token) Output() int64 {
	return t.output
}

// CacheRead returns the number of cache read tokens
func (t Token) CacheRead() int64 {
	return t.cacheRead
}

// CacheCreation returns the number of cache creation tokens
func (t Token) CacheCreation() int64 {
	return t.cacheCreation
}

// Total returns the total number of tokens
func (t Token) Total() int64 {
	return t.input + t.output + t.cacheRead + t.cacheCreation
}

// Limited returns the number of tokens that count against limits (input + output)
func (t Token) Limited() int64 {
	return t.input + t.output
}

// Cache returns the total cache tokens (read + creation)
func (t Token) Cache() int64 {
	return t.cacheRead + t.cacheCreation
}

// Add returns a new Token with the sum of two token counts
func (t Token) Add(other Token) Token {
	return Token{
		input:         t.input + other.input,
		output:        t.output + other.output,
		cacheRead:     t.cacheRead + other.cacheRead,
		cacheCreation: t.cacheCreation + other.cacheCreation,
	}
}
