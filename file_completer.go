package kong

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CompleteDirs will search for directories in the given started to be typed path.
//
// If no path was started to be typed, it will complete to directories in the current working directory.
func CompleteDirs() CompleterFunc {
	return func(args CompleterArgs) []string {
		_, options := subdirOptions(args)

		// when there isn't exactly 1 option, we either have many results or have no results, so we return it.
		if len(options) != 1 {
			return options
		}

		// if there is only one option and it's a directory, try again with that directory as the last arg
		if len(options) == 1 && isExistingDir(options[0]) {
			if len(args) == 0 {
				args = append(args, "")
			}
			args[len(args)-1] = options[0]
			_, options = subdirOptions(args)
		}
		return options
	}
}

// CompleteFiles will search for files matching the given pattern in the started to be typed path.
//
// If no path was started to be typed, it will complete to files that match the pattern in the
// current working directory. To match any file, use "*" as pattern. To match go files use "*.go", and so on.
func CompleteFiles(pattern string) CompleterFunc {
	return func(args CompleterArgs) []string {
		options := fileOptions(args, pattern)

		// when there isn't exactly 1 option, we either have many results or have no results, so we return it.
		if len(options) != 1 {
			return options
		}

		// if there is only one option and it's a directory, try again with that directory as the last arg
		if len(options) == 1 && isExistingDir(options[0]) {
			if len(args) == 0 {
				args = append(args, "")
			}
			args[len(args)-1] = options[0]
			options = fileOptions(args, pattern)
		}
		return options
	}
}

// isExistingDir returns true if path points to a directory that exists.
// always returns false when os.Stat returns an error
func isExistingDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func fileOptions(args CompleterArgs, pattern string) []string {
	if strings.HasSuffix(args.Last(), "/..") {
		return nil
	}
	dir, results := subdirOptions(args)

	files, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return results
	}

	for _, file := range files {
		if !isExistingDir(file) {
			results = append(results, file)
		}
	}
	return CompleteFilesSet(results).Options(args)
}

func subdirOptions(args CompleterArgs) (dir string, subDirs []string) {
	dir = filepath.FromSlash("./")
	lastArg := args.Last()
	if strings.HasSuffix(lastArg, "/..") {
		return dir, nil
	}
	if isExistingDir(filepath.Dir(lastArg)) {
		dir = formatPathOption(lastArg, filepath.Dir(lastArg))
	}
	if isExistingDir(lastArg) {
		dir = formatPathOption(lastArg, lastArg)
	}
	contents, err := ioutil.ReadDir(dir)
	if err != nil {
		return dir, nil
	}
	subDirs = make([]string, 0, len(contents))
	for _, info := range contents {
		if info.IsDir() {
			subDirs = append(subDirs, filepath.Join(dir, info.Name()))
		}
	}
	return dir, CompleteFilesSet(append(subDirs, dir)).Options(args)
}

// CompleteFilesSet is like CompleteSet but for files
func CompleteFilesSet(files []string) CompleterFunc {
	return func(args CompleterArgs) []string {
		options := make([]string, 0, len(files))
		for _, f := range files {
			lastArg := args.Last()
			f = formatPathOption(lastArg, f)

			dotSlash := filepath.FromSlash("./")
			if f == dotSlash && (lastArg == "." || lastArg == "") {
				options = append(options, f)
				continue
			}

			if lastArg == "." && strings.HasPrefix(f, ".") {
				options = append(options, f)
				continue
			}

			if strings.HasPrefix(strings.TrimPrefix(f, dotSlash), strings.TrimPrefix(lastArg, dotSlash)) {
				options = append(options, f)
				continue
			}
		}
		return options
	}
}

// formatPathOption returns path in the form needed for a completion option
func formatPathOption(base string, path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	separator := string(filepath.Separator)

	// if base is absolute, return path as absolute
	if filepath.IsAbs(base) {
		if isExistingDir(abs) {
			abs = strings.TrimSuffix(abs, separator) + separator
		}
		return abs
	}

	wd, err := os.Getwd()
	if err != nil {
		return path
	}

	rel, err := filepath.Rel(wd, abs)
	if err != nil {
		return path
	}

	// the result should start with "./" when base starts with .
	if rel != "." && strings.HasPrefix(base, ".") {
		rel = "." + separator + rel
	}

	if isExistingDir(rel) {
		rel = strings.TrimSuffix(rel, separator) + separator
	}
	return rel
}
