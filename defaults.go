package kong

// ApplyDefaults if they are not already set.
func ApplyDefaults(target interface{}, options ...Option) (err error) {
	var (
		app *Kong
		ctx *Context
	)

	if app, err = New(target, options...); err != nil {
		return
	}
	if ctx, err = Trace(app, nil); err != nil {
		return
	}
	if err = ctx.Resolve(); err != nil {
		return
	}
	if err = ctx.ApplyDefaults(); err != nil {
		return
	}
	err = ctx.Validate()

	return
}
