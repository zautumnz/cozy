# cozy

This is a WIP. See [TODO](./TODO.md).

[![Support with PayPal](https://img.shields.io/badge/paypal-donate-yellow.png)](https://paypal.me/zacanger) [![Patreon](https://img.shields.io/badge/patreon-donate-yellow.svg)](https://www.patreon.com/zacanger) [![ko-fi](https://img.shields.io/badge/donate-KoFi-yellow.svg)](https://ko-fi.com/U7U2110VB)

Simple, high-ish-level interpreted programming language that sits somewhere
between scripting and general-purpose programming. Dynamically and strongly
typed, with some with semantics that work well with pseudo-functional
programming but syntax similar to Python, Go, JavaScript, and Shell; no OOP
constructs like classes; instead we have first-class functions, closures, and
macros.

---

## Example

```cozy
let reduce = fn(fun, xs, init) {
    mutable acc = init

    foreach i, x in xs {
        acc = fun(x, acc, i)
    }

    return acc
}

let ints? = fn(xs) {
    foreach x in xs {
        if type(x) != "integer" && type(x) != "float" {
            return false
        }
    }

    return true
}

let sum = fn(xs) {
    assert(ints?(xs), "expected only numbers!")
    return reduce(
        fn(x, acc) {
            return x + acc
        }, xs, 0)
}

print(sum([1, 2, 3, 4]) == 10) # true
```

For more examples and documentation, see the [examples](./examples) and
[stdlib](./stdlib) directories. For vim, CLOC, Ctags, and Emacs support, see the
[editor](./editor) directory.

## About

Originally designed by writing a bunch of examples and a small stdlib.
Implementation started as a fork from [skx's
version](https://github.com/skx/monkey) of the language from the [Go
Interpreters Book](https://interpreterbook.com), and also includes some pieces
of [prologic's](https://github.com/prologic/monkey-lang) upstream version.
Written in pure Go with no third-party dependencies, with a large amount of the
standard library implemented in cozy itself.

This is the first large Go program I've worked on, so contributions, especially
in areas where I didn't write idiomatic Go, are definitely welcome. See
[CONTRIBUTING](.github/CONTRIBUTING.md) for contribution guidelines.

The majority of the upstream versions of this code are MIT licensed. This code
is LGPL-3.0 licensed (see [LICENSE.md](./LICENSE.md)).

## Usage

Clone the repo and run `make`, and either copy the binary to somewhere in your
path or run `make install`. Write some code (see the examples), and run `cozy
./your-code.cz`. You can also run without a specified file, in which case your
entered code will be evaluated when you exit with `ctrl+d`.

## Important Notes

* `print` adds an ending newline, use `printf` or `STDOUT`/`STDERR` for raw text
* No null/nil, undefined, uninitialized variables
* Comments are Python/Shell style
* No switch statements
* Using `set` and `delete` on hashes returns a new hash
* `let` is for immutable variables; `mutable` is for mutable ones; this is
    because setting mutable variables should be more annoying to do than
    setting mutable ones.
* Uses Go's GC; porting to a different language might require writing a new GC.
* Semicolons are optional
* Most statements are expressions, including if/else; this also means implicit
    returns (without the `return` keyword) are possible
* No top level mutable variables, because all top level variables are exported
* Parens are optional in for and if conditions

### Builtins

Global functions:

* `async`/`await` for async functions
* `eval` evals a cozy string
* `float` converts to a float
* `import` imports another cozy file as a module
* `int` convert the given float/string to an integer
* `len` Yield the length of builtin containers
* `macro`/`quote`/`unquote` for building macros
* `match` Regular-expression matching
* `print` Write values to STDOUT with newlines
* `printf` Write values to STDOUT, via a format-string
* `sprintf` Create strings, via a format-string
* `string` convert the given item to a string
* `type` returns the type of a variable.

Core modules (see examples for docs):

* `fs`
* `json`
* `math`
* `net`
* `time`

See also the standard library (written mostly in cozy itself).
