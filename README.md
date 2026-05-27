![Tests](https://img.shields.io/github/actions/workflow/status/z46-dev/goconf/tests.yml?branch=main&event=push&label=Tests&job=run-tests)
![Made with Golang](https://img.shields.io/badge/-Made_with_Golang-007d9c?logo=go&logoColor=white)

# goconf

A wrapper for a few modules I use for TOML-parsed configuration in my Go projects

## Modules Used

> You should really read the documentation for the following modules, as you will define the struct tags for your config struct using these modules.

- [BurntSushi/toml](https://github.com/BurntSushi/toml) for TOML parsing and encoding
- [go-playground/validator](https://github.com/go-playground/validator) for struct validation based on tags
- [creasty/defaults](https://github.com/creasty/defaults) for setting default values based on struct tags

---

## Usage

First, install the module:

```bash
go get -u github.com/z46-dev/goconf@latest
```

The most simple usage is just to load a config file into a struct, panicking if there are any errors:

```go
package main

import "github.com/z46-dev/goconf"

type Configuration struct {
    Name string `toml:"name" default:"John Doe" validate:"required"`
    Age  int    `toml:"age" default:"30" validate:"required,gte=21"`
}

func main() {
    var cfg Configuration = goconf.MustLoadConfig[Configuration]("./config.toml")

    // ...
}
```

You can also pass options to `LoadConfig` or `MustLoadConfig` to customize certain behaviors. For example, you can specify what to do if the config file doesn't exist:

```go
if cfg, err := goconf.LoadConfig[Configuration]("./config.toml", goconf.WithNewFileBehavior(goconf.NewFileBehaviorCreateAndTry)); err != nil {
    // handle error
} else {
    // use cfg
}
```

Currently, the `WithNewFileBehavior` option has three behaviors:
- `NewFileBehaviorReject`: if the file doesn't exist, return an error (this is the default behavior)
- `NewFileBehaviorCreateAndReject`: if the file doesn't exist, create it with default values based on the struct tags, but still return an error (so the user can fill it out before trying again)
- `NewFileBehaviorCreateAndTry`: if the file doesn't exist, create it with default values based on the struct tags, and then try to load it again immediately (so the user doesn't have to do anything on the first run, but it will still error if the defaults don't pass validation)

The other option is `WithIndentSpaces`, which allows you to specify how many spaces to use for indentation when encoding the config struct back to a file. By default, it uses 4 spaces.

```go
if cfg, err := goconf.LoadConfig[Configuration]("./config.toml", goconf.WithIndentSpaces(2)); err != nil {
    // handle error
} else {
    // use cfg
}
```

## Example

You can run the example in `example/main.go` to see how it works.

```bash
go run github.com/z46-dev/goconf/example@latest
```