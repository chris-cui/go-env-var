# go-env-var
Load environment variables to a go struct, with some additional features like default value, validation and
converting string to other types.

***Note***, if Load function encounter any error then it returns error immediately without further processing.

## Add to project
```shell
go get github.com/chris-cui/go-env-var@v0.2.0
```


## Simple usage - single struct, string or string pointer type
```
type Book struct {
    Name   *string `env:"BOOK_NAME""`                    // nil if no BOOK_NAME env variable"
    Genre  string  `env:"BOOK_GENRE" default:"Fiction"`  // "Fiction" if no BOOK_GENRE env variable
    Isbn   string  `env:"BOOK_ISBN" required:"true"`     // return error if no BOOK_ISBN env variable
}

b := &Book{}
// load 3 env variables BOOK_NAME, BOOK_GENRE, BOOK_ISBN
err := envvar.Load(b)

```

## All supported tags(default)
- "env" - environment variable name
- "default" - default value if no given environment variable
- "required" - accept "true" only, return error if the filed value is zero value
- "converter" - convert string to any type, you have to register conversion function with consistent return type


## To override one or more default tags

Example:
```
envvar.TagEnvVar    = "z-env"
envvar.TagDefault   = "z-default"
envvar.TagRequired  = "z-required"
envvar.TagConverter = "z-converter"
```

## Nested and non-string type
- nested struct with env tags must be a pointer and must be initialized
- you have to provide your own converter for non-string type

Example:
```

type Book struct {
    Name   *string `env:"BOOK_NAME""`
    Genre  string  `env:"BOOK_GENRE" default:"Fiction"`
    Code   string  `env:"BOOK_ISBN" required:"true"`
    Pages  *int    `env:"BOOK_PAGES" converter:"toIntPtr"`
    Author *Author
}

type Author struct {
    Name string `env:"AUTHOR_NAME" required:"true"`
    Age  int    `env:"AUTHOR_AGE" required:"true" converter:"toInt"`
}

// convert string to int
envvar.Converter("toInt", func(s string) (any, error) {
    i, e := strconv.Atoi(s)
    return i, e
})

// convert string to *int
envvar.Converter("toIntPtr", func(s string) (any, error) {
    i, e := strconv.Atoi(s)
    return &i, e
})

b := &Book{
    Author: &Author{},
}
err := envvar.Load(b)

```

## Validation
Besides `required` there is no additional tag for custom validation, but you can use converter function for
any validation.

Example:
```

envvar.Converter("min10", func(s string) (any, error) {
    if len(s) < 10 {
        return s, errors.New("minimum 10 characters")
    }
    return s, nil
})

```

## To clear all converter functions
It is not mandatory, but after loading all environment variables you can clear converter functions with following code.
```
envvar.ClearConverters()

```
