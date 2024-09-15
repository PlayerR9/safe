package c_string

import (
	"github.com/gdamore/tcell"
)

// DefaultIndentationConfig is the default indentation configuration.
//
// Parameters:
//   - style: The style of the indentation.
//
// ==IndentConfig==
//   - DefaultIndentation
//   - 0
//
// Returns:
//   - *IndentConfig: A pointer to the default indentation configuration.
func DefaultIndentationConfig(style tcell.Style) *IndentConfig {
	return NewIndentConfig([]*Unit{NewUnit(DefaultIndentation, style)}, 0)
}

const (
	// DefaultIndentation is the default indentation string.
	DefaultIndentation string = "   "

	// DefaultSeparator is the default separator string.
	DefaultSeparator string = ", "
)

var (
	// DefaultSeparatorConfig is the default separator configuration.
	DefaultSeparatorConfig *SeparatorConfig = NewSeparator(DefaultSeparator, false)
)

// IndentConfig is a type that represents the configuration for indentation.
type IndentConfig struct {
	// str is the string that is used for indentation.
	units []*Unit

	// InitialLevel is the current indentation level.
	level int
}

// Copy is a method of uc.Copier interface.
//
// Returns:
//   - *IndentConfig: A copy of the indentation configuration.
func (c *IndentConfig) Copy() *IndentConfig {
	configCopy := &IndentConfig{
		level: c.level,
		units: make([]*Unit, 0, len(c.units)),
	}

	for _, unit := range c.units {
		configCopy.units = append(configCopy.units, unit.Copy())
	}

	return configCopy
}

// NewIndentConfig is a function that creates a new indentation configuration.
//
// Parameters:
//   - indentation: The string that is used for indentation.
//   - initialLevel: The initial indentation level.
//
// Returns:
//   - *IndentConfig: A pointer to the new indentation configuration.
//
// Default values:
//
//		==IndentConfig==
//	  - Indentation: DefaultIndentation
//	  - InitialLevel: 0
//
// Behaviors:
//   - If initialLevel is negative, it is set to 0.
func NewIndentConfig(units []*Unit, initialLevel int) *IndentConfig {
	if initialLevel < 0 {
		initialLevel = 0
	}

	config := &IndentConfig{
		units: ReduceUnitSequence(units),
		level: initialLevel,
	}

	return config
}

// GetIndentation is a method that returns the applied indentation.
//
// Returns:
//   - string: The applied indentation.
func (c *IndentConfig) GetIndentation() []*Unit {
	sequence := make([]*Unit, 0, len(c.units)*c.level)

	for i := 0; i < c.level; i++ {
		sequence = append(sequence, c.units...)
	}

	return ReduceUnitSequence(sequence)
}

// GetIndentStr is a method that returns the indentation string.
//
// Returns:
//   - string: The indentation string.
func (c *IndentConfig) GetIndentStr() []*Unit {
	return c.units
}

// SeparatorConfig is a type that represents the configuration for separators.
type SeparatorConfig struct {
	// str is the string that is used as a separator.
	str string

	// includeFinal specifies whether the last element should have a separator.
	includeFinal bool
}

// Copy is a method of uc.Copier interface.
//
// Returns:
//   - *SeparatorConfig: A copy of the separator configuration.
func (c *SeparatorConfig) Copy() *SeparatorConfig {
	return &SeparatorConfig{
		str:          c.str,
		includeFinal: c.includeFinal,
	}
}

// NewSeparator is a function that creates a new separator configuration.
//
// Parameters:
//   - separator: The string that is used as a separator.
//   - hasFinalSeparator: Whether the last element should have a separator.
//
// Returns:
//   - *SeparatorConfig: A pointer to the new separator configuration.
//
// Default values:
//
//		==SeparatorConfig==
//	  - Separator: DefaultSeparator
//	  - HasFinalSeparator: false
func NewSeparator(sep string, includeFinal bool) *SeparatorConfig {
	return &SeparatorConfig{
		str:          sep,
		includeFinal: includeFinal,
	}
}

// DelimiterConfig is a type that represents the configuration for delimiters.
type DelimiterConfig struct {
	// str is the string that is used as a delimiter.
	str string

	// isInline specifies whether the delimiter should be inline.
	isInline bool

	// left specifies whether the delimiter is on the left side.
	left bool
}

// Copy is a method of uc.Copier interface.
//
// Returns:
//   - *DelimiterConfig: A copy of the delimiter configuration.
func (c *DelimiterConfig) Copy() *DelimiterConfig {
	return &DelimiterConfig{
		str:      c.str,
		isInline: c.isInline,
		left:     c.left,
	}
}

// NewDelimiterConfig is a function that creates a new delimiter configuration.
//
// Parameters:
//   - value: The string that is used as a delimiter.
//   - inline: Whether the delimiter should be inline.
//
// Returns:
//   - *DelimiterConfig: A pointer to the new delimiter configuration.
//
// Default values:
//   - ==DelimiterConfig==
//   - Value: ""
//   - Inline: true
func NewDelimiterConfig(str string, isInline, left bool) *DelimiterConfig {
	return &DelimiterConfig{
		str:      str,
		isInline: isInline,
		left:     left,
	}
}

// StyleConfig is a type that represents the configuration for styles.
type StyleConfig struct {
	// defaultStyle is the default style to use.
	defaultStyle tcell.Style
}

// Copy is a method of uc.Copier interface.
//
// Returns:
//   - *StyleConfig: A copy of the style configuration.
func (c *StyleConfig) Copy() *StyleConfig {
	return &StyleConfig{
		defaultStyle: c.defaultStyle,
	}
}

// NewStyleConfig is a function that creates a new style configuration.
//
// Parameters:
//   - defaultStyle: The default style to use.
//
// Returns:
//   - *StyleConfig: A pointer to the new style configuration.
func NewStyleConfig(defaultStyle tcell.Style) *StyleConfig {
	return &StyleConfig{
		defaultStyle: defaultStyle,
	}
}

//////////////////////////////////////////////////////////////

/*



func (config *IndentConfig) apply(values []string) []string {
	if len(values) == 0 {
		return []string{config.Indentation}
	}

	var builder strings.Builder

	result := make([]string, len(values))
	copy(result, values)

	for i := 0; i < len(result); i++ {
		builder.Reset()

		builder.WriteString(config.Indentation)
		builder.WriteString(result[i])

		result[i] = builder.String()
	}

	return result
}



func (config *SeparatorConfig) apply(values []string) []string {
	switch len(values) {
	case 0:
		if config.HasFinalSeparator {
			return []string{config.Separator}
		}

		return []string{}
	case 1:
		var builder strings.Builder

		builder.WriteString(values[0])

		if config.HasFinalSeparator {
			builder.WriteString(config.Separator)
		}

		return []string{builder.String()}
	default:
		result := make([]string, len(values))
		copy(result, values)

		var builder strings.Builder

		builder.WriteString(result[0])
		builder.WriteString(config.Separator)

		result[0] = builder.String()

		for i := 1; i < len(result)-1; i++ {
			builder.Reset()

			builder.WriteString(result[i])
			builder.WriteString(config.Separator)
			result[i] = builder.String()
		}

		if config.HasFinalSeparator {
			builder.Reset()

			builder.WriteString(result[len(result)-1])
			builder.WriteString(config.Separator)
			result[len(result)-1] = builder.String()
		}

		return result
	}
}


func (config *DelimiterConfig) applyOnLeft(values []string) []string {
	if len(values) == 0 {
		return []string{config.Value}
	}

	result := make([]string, len(values))
	copy(result, values)

	if config.Inline {
		var builder strings.Builder

		builder.WriteString(config.Value)
		builder.WriteString(values[0])

		result[0] = builder.String()
	} else {
		result = append([]string{config.Value}, result...)
	}

	return result
}

func (config *DelimiterConfig) applyOnRight(values []string) []string {
	if len(values) == 0 {
		return []string{config.Value}
	}

	result := make([]string, len(values))
	copy(result, values)

	if config.Inline {
		var builder strings.Builder

		builder.WriteString(values[len(values)-1])
		builder.WriteString(config.Value)

		result[len(values)-1] = builder.String()
	} else {
		result = append(result, config.Value)
	}

	return result
}
*/
