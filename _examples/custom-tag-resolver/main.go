package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/alecthomas/kong"
)

// MyResolver is a custom resolver that resolves the value of a flag from an HTTP endpoint.
func MyResolver(baseEndpoint string) kong.Resolver {
	return kong.ResolverFunc((func(_ *kong.Context, _ *kong.Path, flag *kong.Flag) (interface{}, error) {
		// NOTE: Since all flags are resolved using this resolver, you should check if the flag should be resolved by this resolver.
		// returning nil for the left value means that the flag should not be resolved by this resolver.
		if flag.Tag.Get("http_path") == "" {
			return nil, nil
		}
		resp, err := http.Get(baseEndpoint + flag.Tag.Get("http_path"))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return string(buf), nil
	}))
}

func main() {
	count := 0
	go func() {
		http.HandleFunc("/secret-path/1", func(w http.ResponseWriter, r *http.Request) {
			count++
			w.Write([]byte("secret-value1"))
		})
		http.HandleFunc("/secret-path/3", func(w http.ResponseWriter, r *http.Request) {
			count++
			w.Write([]byte("secret-value3"))
		})
		http.ListenAndServe(":8080", nil)
	}()

	var cli struct {
		Flag1 string `kong:"http_path='/secret-path/1'"`    // resolve the value from "localhost:8080/secret-path/1"
		Flag2 string `kong:"default='default-flag2-value'"` // this flag will not be resolved by MyResolver
		Flag3 string `kong:"http_path='/secret-path/3'"`    // resolve the value from "localhost:8080/secret-path/3"
	}
	parser := kong.Must(&cli, kong.Resolvers(MyResolver("http://localhost:8080")))
	_, err := parser.Parse([]string{})
	if err != nil {
		panic(err)
	}

	// print the resolved values
	fmt.Printf("%+v\n", cli)
	// print the number of requests, it should be 2
	fmt.Printf("number of requests: %d\n", count)
}
