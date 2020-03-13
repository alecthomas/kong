package kong

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CompleteFilesSet is like CompleteSet but for files. It applies formatPathOption() before matching.
//  see formatPathOption() for more details.
func CompleteFilesSet(files []string) CompleterFunc {
	return func(args CompleterArgs) []string {
		return filterFilesByPrefix(args.Last(), files)
	}
}

// CompleteDirs will search for directories in the given started to be typed path.
//
// If no path was started to be typed, it will complete to directories in the current working directory.
func CompleteDirs() CompleterFunc {
	return func(args CompleterArgs) []string {
		return recurseDirectoryMatches(args.Last(), 1, subdirOptions)
	}
}

// CompleteFiles will search for files matching the given pattern in the started to be typed path.
//
// If no path was started to be typed, it will complete to files that match the pattern in the
// current working directory. To match any file, use "*" as pattern. To match go files use "*.go", and so on.
func CompleteFiles(pattern string) CompleterFunc {
	return func(args CompleterArgs) []string {
		return recurseDirectoryMatches(args.Last(), -1, func(prefix string) []string {
			//return fileOptions(prefix, pattern)
			if strings.HasSuffix(prefix, "/..") {
				return nil
			}
			dir := matcherBaseDirectory(prefix)
			results := subdirOptions(prefix)

			files, err := filepath.Glob(filepath.Join(dir, pattern))
			if err != nil {
				return results
			}

			for _, file := range files {
				if !isExistingDir(file) {
					results = append(results, file)
				}
			}
			return filterFilesByPrefix(prefix, results)
		})
	}
}

// recurseDirectoryMatches runs fn() on its output as long as the output is a single directory and it hasn't reached
//  maxDepth. When maxDepth is negative, there is no recursion limit.
func recurseDirectoryMatches(prefix string, maxDepth int, fn func(string) []string) []string {
	var options []string
	for {
		options = fn(prefix)
		if len(options) != 1 || !isExistingDir(options[0]) || options[0] == prefix || maxDepth == 0 {
			break
		}
		maxDepth--
		prefix = options[0]
	}
	return options
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

// matcherBaseDirectory returns the directory for file matchers to use for a given prefix.
//  If the prefix or its parent is a directory, that is returned.  Otherwise it returns "./"
func matcherBaseDirectory(prefix string) string {
	switch {
	case isExistingDir(prefix):
		return formatPathOption(prefix, prefix)
	case isExistingDir(filepath.Dir(prefix)):
		return formatPathOption(prefix, filepath.Dir(prefix))
	default:
		return filepath.FromSlash("./")
	}
}

// subdirOptions returns the directories that match prefix
func subdirOptions(prefix string) (subDirs []string) {
	dir := matcherBaseDirectory(prefix)
	contents, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}
	subDirs = make([]string, 0, len(contents))
	for _, info := range contents {
		if info.IsDir() {
			subDirs = append(subDirs, filepath.Join(dir, info.Name()))
		}
	}
	return filterFilesByPrefix(prefix, append(subDirs, dir))
}

// filterFilesByPrefix returns the members of files that with the given prefix
func filterFilesByPrefix(prefix string, files []string) []string {
	const dotSlash = "." + string(filepath.Separator)
	options := make([]string, 0, len(files))
	for _, file := range files {
		file = formatPathOption(prefix, file)

		// for matching purposes, "." is equivalent to "./"
		matchPrefix := prefix
		if matchPrefix == "." {
			matchPrefix = dotSlash
		}

		// strip "./" from the front of both strings before doing the prefix match
		matchPrefix = strings.TrimPrefix(matchPrefix, dotSlash)
		if strings.HasPrefix(strings.TrimPrefix(file, dotSlash), matchPrefix) {
			options = append(options, file)
		}
	}
	return options
}

// formatPathOption returns path in the form needed for a path completion
//  - when base is an absolute path, the returned value is an absolute path
//  - when base starts with ".", the returned value starts with "./"
//  - when base is a relative path, the retuned value is a relative path to the current working directory (not base)
//  - when path points to a directory, the returned value ends with "/"
func formatPathOption(base string, path string) string {
	// if base is absolute, return path as absolute
	if filepath.IsAbs(base) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return path
		}
		return appendSeparatorToDirectory(abs)
	}

	rel, err := filepath.Rel("", path)
	if err != nil {
		return path
	}

	// the result should start with "./" when base starts with .
	if rel != "." && strings.HasPrefix(base, ".") {
		rel = filepath.FromSlash("./") + rel
	}

	return appendSeparatorToDirectory(rel)
}

// appendSeparatorToDirectory adds "/" to the end of path if it isn't already present and path points to a directory
func appendSeparatorToDirectory(path string) string {
	if strings.HasSuffix(path, string(filepath.Separator)) || !isExistingDir(path) {
		return path
	}
	return path + string(filepath.Separator)
}
