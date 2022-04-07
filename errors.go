// Package kong
package kong

import "strings"

const (
	cUnknownError                          = "unknown error: %s"
	cNoCommandSelected                     = "no command selected"
	cExpected                              = "expected %s"
	cExpectedOneOf                         = "expected one of %s"
	cUnexpectedArgument                    = "unexpected argument %s"
	cUnexpectedFlagArgument                = "unexpected flag argument %q"
	cUnexpectedToken                       = "unexpected token %s"
	cUnknownFlag                           = "unknown flag %s"
	cMissingFlags                          = "missing flags: %s"
	cHelperErrorDidYouMean                 = "%s, did you mean %s?"
	cHelperErrorDidYouMeanOneOf            = "%s, did you mean one of %s?"
	cDefaultValueFor                       = "default value for %s: %s"
	cEnumValueFor                          = "enum value for %s: %s"
	cEnvValueFor                           = "env value for %s: %s"
	cHelpFor                               = "help for %s: %s"
	cIsRequired                            = "%s is required"
	cEnumSliceOrValue                      = "enum can only be applied to a slice or value"
	cMustBeOneOfButGot                     = "%s must be one of %s but got %q"
	cMissingPositionalArguments            = "missing positional arguments %s"
	cFailField                             = "%s.%s: %s"
	cRegexInputCannotBeEmpty               = "regex input cannot be empty"
	cRegexUnableToCompile                  = "unable to compile regex: %s"
	cKongMustBeConfiguredWithConfiguration = "kong must be configured with kong.Configuration"
	cUndefinedVariable                     = "undefined variable ${%s}"
)

var (
	errSingleton                             = &Error{}
	errUnknownError                          = err{tpl: cUnknownError}
	errNoCommandSelected                     = err{tpl: cNoCommandSelected}
	errExpected                              = err{tpl: cExpected}
	errExpectedOneOf                         = err{tpl: cExpectedOneOf}
	errUnexpectedArgument                    = err{tpl: cUnexpectedArgument}
	errUnexpectedFlagArgument                = err{tpl: cUnexpectedFlagArgument}
	errUnexpectedToken                       = err{tpl: cUnexpectedToken}
	errUnknownFlag                           = err{tpl: cUnknownFlag}
	errMissingFlags                          = err{tpl: cMissingFlags}
	errHelperErrorDidYouMean                 = err{tpl: cHelperErrorDidYouMean}
	errHelperErrorDidYouMeanOneOf            = err{tpl: cHelperErrorDidYouMeanOneOf}
	errDefaultValueFor                       = err{tpl: cDefaultValueFor}
	errEnumValueFor                          = err{tpl: cEnumValueFor}
	errEnvValueFor                           = err{tpl: cEnvValueFor}
	errHelpFor                               = err{tpl: cHelpFor}
	errIsRequired                            = err{tpl: cIsRequired}
	errEnumSliceOrValue                      = err{tpl: cEnumSliceOrValue}
	errMustBeOneOfButGot                     = err{tpl: cMustBeOneOfButGot}
	errMissingPositionalArguments            = err{tpl: cMissingPositionalArguments}
	errFailField                             = err{tpl: cFailField}
	errRegexInputCannotBeEmpty               = err{tpl: cRegexInputCannotBeEmpty}
	errRegexUnableToCompile                  = err{tpl: cRegexUnableToCompile}
	errKongMustBeConfiguredWithConfiguration = err{tpl: cKongMustBeConfiguredWithConfiguration}
	errUndefinedVariable                     = err{tpl: cUndefinedVariable}
)

// ERRORS: Implementation of errors with the ability to compare errors with each other

// UnknownError Unknown Error
func (e *Error) UnknownError(err error) Err { return newErr(&errUnknownError, err) }

// NoCommandSelected no command selected
func (e *Error) NoCommandSelected() Err { return newErr(&errNoCommandSelected) }

// Expected children ...
func (e *Error) Expected(s string) Err { return newErr(&errExpected, s) }

// ExpectedOneOf Expected one of ...
func (e *Error) ExpectedOneOf(s string) Err { return newErr(&errExpectedOneOf, s) }

// UnexpectedArgument Unexpected argument ...
func (e *Error) UnexpectedArgument(argument string) Err {
	return newErr(&errUnexpectedArgument, argument)
}

// UnexpectedFlagArgument unexpected flag argument
func (e *Error) UnexpectedFlagArgument(argument interface{}) Err {
	return newErr(&errUnexpectedFlagArgument, argument)
}

// UnexpectedToken Unexpected token ...
func (e *Error) UnexpectedToken(token Token) Err { return newErr(&errUnexpectedToken, token) }

// UnknownFlag unknown flag ...
func (e *Error) UnknownFlag(flag string) Err { return newErr(&errUnknownFlag, flag) }

// MissingFlags missing flags: ...
func (e *Error) MissingFlags(flag string) Err { return newErr(&errMissingFlags, flag) }

// HelperErrorDidYouMean ..., did you mean ...?
func (e *Error) HelperErrorDidYouMean(err error, hypothesis string) Err {
	return newErr(&errHelperErrorDidYouMean, err, hypothesis)
}

// HelperErrorDidYouMeanOneOf ..., did you mean one of ...?
func (e *Error) HelperErrorDidYouMeanOneOf(err error, hypothesis []string) Err {
	const hypothesisDelimiter = `, `
	return newErr(&errHelperErrorDidYouMeanOneOf, err, strings.Join(hypothesis, hypothesisDelimiter))
}

// DefaultValueFor default value for ...
func (e *Error) DefaultValueFor(s string, err error) Err { return newErr(&errDefaultValueFor, s, err) }

// EnumValueFor enum value for ...
func (e *Error) EnumValueFor(s string, err error) Err { return newErr(&errEnumValueFor, s, err) }

// EnvValueFor env value for ...
func (e *Error) EnvValueFor(s string, err error) Err { return newErr(&errEnvValueFor, s, err) }

// HelpFor help for ...
func (e *Error) HelpFor(s string, err error) Err { return newErr(&errHelpFor, s, err) }

// IsRequired ... is required
func (e *Error) IsRequired(required string) Err { return newErr(&errIsRequired, required) }

// EnumSliceOrValue Enum can only be applied to a slice or value
func (e *Error) EnumSliceOrValue() Err { return newErr(&errEnumSliceOrValue) }

// MustBeOneOfButGot ... must be one of ... but got ...
func (e *Error) MustBeOneOfButGot(summary string, enums string, item interface{}) Err {
	return newErr(&errMustBeOneOfButGot, summary, enums, item)
}

// MissingPositionalArguments missing positional arguments ...
func (e *Error) MissingPositionalArguments(arguments string) Err {
	return newErr(&errMissingPositionalArguments, arguments)
}

// FailField Error for check field
func (e *Error) FailField(parent, name, value string) Err {
	return newErr(&errFailField, parent, name, value)
}

// RegexInputCannotBeEmpty Regex input cannot be empty
func (e *Error) RegexInputCannotBeEmpty() Err { return newErr(&errRegexInputCannotBeEmpty) }

// RegexUnableToCompile Unable to compile regex: ...
func (e *Error) RegexUnableToCompile(err error) Err { return newErr(&errRegexUnableToCompile, err) }

// KongMustBeConfiguredWithConfiguration Kong must be configured with kong.Configuration
func (e *Error) KongMustBeConfiguredWithConfiguration() Err {
	return newErr(&errKongMustBeConfiguredWithConfiguration)
}

// UndefinedVariable undefined variable ...
func (e *Error) UndefinedVariable(name string) Err { return newErr(&errUndefinedVariable, name) }
