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

func commandSignatureTagFromValue(v reflect.Value) (string, bool) {
	if !v.IsValid() {
		return "", false
	}
	var (
		sig string
		has bool
	)
	check := func(candidate any) {
		if candidate == nil {
			return
		}
		if provider, okProvider := candidate.(SignatureOverride); okProvider {
			sig = provider.Signature()
			has = true
		}
	}
	if v.CanInterface() {
		check(v.Interface())
	}
	if !has && v.CanAddr() && v.Addr().CanInterface() {
		check(v.Addr().Interface())
	}
	if !has {
		return "", false
	}
	sig = strings.TrimSpace(sig)
	if sig == "" {
		return "", false
	}
	return sig, true
}

func applyCommandSignature(v reflect.Value, ft reflect.StructField, fv reflect.Value, tag *Tag) error {
	sigTag, ok := commandSignatureTagFromValue(fv)
	if !ok {
		return nil
	}

	items, err := parseTagItems(sigTag, bareChars)
	if err != nil {
		return failField(v, ft, "signature: %s", err)
	}

	mergedItems := make(map[string][]string, len(tag.items)+len(items))
	for key, values := range tag.items {
		mergedItems[key] = append([]string(nil), values...)
	}
	// Signature tags are appended after field tags, so existing struct tags
	// take precedence for single-value options.
	for key, values := range items {
		mergedItems[key] = append(mergedItems[key], values...)
	}
	merged := &Tag{items: mergedItems}
	if err := hydrateTag(merged, ft.Type); err != nil {
		return failField(v, ft, "signature: %s", err)
	}
	*tag = *merged

	return nil
}
