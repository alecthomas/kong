package kong

import (
	"reflect"
	"strings"
)

// SignatureOverride overrides command metadata via a bare tag string.
// Example: `name:"migrate" help:"Run migrations" aliases:"mig,mg"`.
type SignatureOverride interface {
	Signature() string
}

func commandSignatureTagFromValue(v reflect.Value) (*Tag, bool, error) {
	if !v.IsValid() {
		return nil, false, nil
	}
	var (
		sig string
		ok  bool
	)
	check := func(candidate any) {
		if candidate == nil {
			return
		}
		if provider, okProvider := candidate.(SignatureOverride); okProvider {
			sig = provider.Signature()
			ok = true
		}
	}
	if v.CanInterface() {
		check(v.Interface())
	}
	if !ok && v.CanAddr() && v.Addr().CanInterface() {
		check(v.Addr().Interface())
	}
	if !ok {
		return nil, false, nil
	}
	sig = strings.TrimSpace(sig)
	if sig == "" {
		return nil, true, nil
	}
	tag, err := parseTagString(sig)
	if err != nil {
		return nil, true, err
	}
	return tag, true, nil
}

func applyCommandSignature(node *Node, v reflect.Value, ft reflect.StructField, fv reflect.Value, tag *Tag, name *string) error {
	sigTag, ok, err := commandSignatureTagFromValue(fv)
	if err != nil {
		return failField(v, ft, "signature: %s", err)
	}
	if !ok || sigTag == nil {
		return nil
	}
	if sigTag.Name != "" {
		*name = sigTag.Name
		tag.Name = sigTag.Name
	}
	if len(sigTag.Aliases) > 0 {
		tag.Aliases = sigTag.Aliases
	}
	if sigTag.Help != "" {
		tag.Help = sigTag.Help
	}
	if sigTag.Has("hidden") {
		tag.Hidden = sigTag.Hidden
	}
	if sigTag.Group != "" {
		tag.Group = sigTag.Group
	}
	return nil
}
