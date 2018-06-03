package kong

import (
	"os"
	"strings"
)

type ResolverInfo struct {
	Flag func(flag *Flag) string
}

//func JSONResolver(r io.Reader) *ResolverInfo {
//	return &ResolverInfo{
//		Grammar: func(grammar interface{}) error {
//			return json.NewDecoder(r).Decode(grammar)
//		},
//	}
//}

func EnvResolver(prefix string) *ResolverInfo {
	return &ResolverInfo{
		Flag: func(flag *Flag) string {
			env := envString(prefix, flag)
			return os.Getenv(env)

			//ctx := DecoderContext{Value: &flag.Value}
			//scan := Scanner{
			//	args: []Token{{Type: FlagValueToken, Value: s}},
			//}
			//flag.Decoder.Decode(&ctx, &scan, flag.Value.Value)
			//return nil
		},
	}
}

func envString(prefix string, flag *Flag) string {
	//if flag.Tag.Has("env") {
	//	env, ok := flag.Tag.Get("env")
	//	if ok {
	//		return env
	//	}
	//}

	env := strings.ToUpper(flag.Name)
	env = strings.Replace(env, "-", "_", -1)
	env = prefix + env

	return env
}
