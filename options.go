package goconf

import "fmt"

type (
	Option          func(*_options)
	newFileBehavior uint8
	_options        struct {
		indentSpaces    int
		newFileBehavior newFileBehavior
	}
)

const (
	NewFileBehaviorReject          newFileBehavior = iota // If the file doesn't exist, return an error
	NewFileBehaviorCreateAndReject                        // If the file doesn't exist, create it with any defaults, but still return an error
	NewFileBehaviorCreateAndTry                           // If the file doesn't exist, create it with any defaults, and then try to load it (which may still error if required fields are missing)
	newFileBehaviorBoundary
)

// Set the indentation spaces for the generated TOML file (default: 4)
func WithIndentSpaces(spaces int) Option {
	return func(opts *_options) {
		opts.indentSpaces = spaces
	}
}

// Set the behavior for when the config file doesn't exist (default: Reject)
func WithNewFileBehavior(behavior newFileBehavior) Option {
	return func(opts *_options) {
		opts.newFileBehavior = behavior
	}
}

// Apply options and validate them
func applyOptions(provided []Option) (opts _options, err error) {
	opts = _options{
		indentSpaces:    4,
		newFileBehavior: NewFileBehaviorReject,
	}

	for _, opt := range provided {
		opt(&opts)
	}

	if opts.indentSpaces < 0 {
		err = fmt.Errorf("indent spaces cannot be negative: %d", opts.indentSpaces)
		return
	}

	if opts.newFileBehavior >= newFileBehaviorBoundary {
		err = fmt.Errorf("invalid new file behavior: %d", opts.newFileBehavior)
		return
	}

	return
}
