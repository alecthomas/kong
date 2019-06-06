package kong

// ApplyDefaults applies defaults to a struct.
func ApplyDefaults(target interface{}, options ...Option) error {
	app, err := New(target, options...)
	if err != nil {
		return err
	}
	_, err = app.Parse(nil)
	return err
}
