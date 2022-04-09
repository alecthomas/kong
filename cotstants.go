// Package kong
package kong

const (
	anonymousStruct         = `<anonymous struct>`
	interpolateValueDefault = `default`
	interpolateValueEnum    = `enum`
	panicUnsupportedPath    = `unsupported Path`
	keyBeforeResolve        = `BeforeResolve`
	keyBeforeApply          = `BeforeApply`
	keyAfterApply           = `AfterApply`
	keyVersion              = `version`
	keyEnv                  = `env`
	keyValue                = `value`
	keyString               = `string`
	onyOther                = `...`
	delimiterComma          = `,`
	delimiterCommaSpace     = `, `
	delimiterPoint          = `.`
	delimiterDash           = `-`
	delimiterDoubleDash     = `--`
	delimiterUnderscore     = `_`
	delimiterSpace          = ` `
	delimiterDollar         = `$`
	cmdWithArgs             = `withargs`
	keyError                = `error`

	// help constants
	helpName             = `help`
	helpHelp             = `Show context-sensitive help.`
	helpOrigin           = `Show context-sensitive help.`
	helpShort            = 'h'
	helpDefaultValue     = false
	defaultIndent        = 2
	defaultColumnPadding = 4
	labelCommands        = `Commands:`
	labelFlags           = `Flags:`
	labelArguments       = `Arguments:`
)

const (
	shortUsage usageOnError = iota + 1
	fullUsage
)
