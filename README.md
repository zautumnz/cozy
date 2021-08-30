# cozy

This is a WIP. See the [TODO](./TODO.md).

[![Support with PayPal](https://img.shields.io/badge/paypal-donate-yellow.png)](https://paypal.me/zacanger) [![Patreon](https://img.shields.io/badge/patreon-donate-yellow.svg)](https://www.patreon.com/zacanger) [![ko-fi](https://img.shields.io/badge/donate-KoFi-yellow.svg)](https://ko-fi.com/U7U2110VB)

`cozy` (always spelled all-lowercase) is a simple, medium-to-high-level,
interpreted, general purpose programming language. It can be used for many of
the same tasks as shell scripts, Python, Node, and Ruby. It's dynamically and
strongly typed, with some with features that work well with pseudo-functional
programming but C-language-family syntax. There are no OOP constructs like
classes; instead we have first-class functions, closures, plain data structures,
and file-based modules.

It's meant to feel comfortable and familiar no matter what languages you might
already know (hence the name). Performance isn't an explicit goal, but cozy is
fairly fast (time the examples and try a benchmarking tool against the http
server if you're curious). It's also meant to have a relatively simple
host-language implementation (for example, much of the standard libarary is
written in cozy itself) to make debugging easy and porting to other host
languages in the future straightforward.

## Example

```cozy
# contrived example to show off some features and syntax; see the ./examples
# directory for more.
# reduce is a built-in method on arrays, int? is a built-in type checking
# function, and sum is a built-in method on arrays, but this example shows
# how they might be implemented by a user.

# let is for immutable variables
let reduce = fn(fun, xs, init) {
    # mutable obvious
    # mutable is only available within blocks, not at at the top level
    mutable acc = init

    foreach i, x in xs {
        acc = fun(x, acc, i)
    }

    return acc
}

# identifiers can have unicode (for example, chinese characters),
# and question marks
let ints? = fn(xs) {
    foreach x in xs {
        if util.type(x) != "integer" && util.type(x) != "float" {
            return false
        }
    }

    # returns can be implicit (without the `return` keyword)
    true
}

let sum = fn(xs) {
    # basic assertions and a TAP-producing test library are built in
    util.assert(ints?(xs), "expected only numbers!")
    return reduce(
        fn(x, acc) {
            return x + acc
        }, xs, 0)
}

# only one level of equality checking, unlike JS's == vs ===
print(sum([1, 2, 3, 4]) == 10) # true
```

For more examples and documentation, see the [examples](./examples) and
[stdlib](./stdlib) directories. The examples also serve as a second test suite.
For vim, CLOC, Ctags, and Emacs support, see the [editor](./editor) directory.

## About

Originally designed by writing a bunch of examples and a small stdlib.
Implementation started as a fork from [skx's
version](https://github.com/skx/monkey) of the language from the [Go
Interpreters Book](https://interpreterbook.com), and also includes some pieces
of [prologic's](https://github.com/prologic/monkey-lang) upstream version and
[ABS](https://github.com/abs-lang), among others (see comments). The differences
between cozy and Monkey are too numerous to list here; it's best to think of it
as a totally separate language. It's written in pure Go with minimal third-party
dependencies, with a large amount of the standard library implemented in cozy
itself.

This is the first large Go program I've worked on, so contributions, especially
in areas where I didn't write idiomatic Go, are definitely welcome. See
[CONTRIBUTING](.github/CONTRIBUTING.md) for contribution guidelines.

## Usage

Clone the repo and run `make`, and either copy the binary to somewhere in your
path or run `make install`. Write some code (see the examples), and run `cozy
./your-code.cz`. You can also run without a specified file, in which case your
entered code will be evaluated when you exit with `ctrl+d`.

### Important Notes

* `print` adds an ending newline, use  or `sys.STDOUT`/`sys.STDERR` for raw text
* No undefined or uninitialized variables Comments are Python/Shell style Errors
* are values, so you can pass them around and use `panic` (like in Go) Using
* `set` and `delete` on hashes returns a new hash `let` is for immutable
* variables; `mutable` is for mutable ones; this is
    because setting mutable variables should be more annoying to do than setting
    mutable ones.
* Uses Go's GC; porting to a different language might require writing a new GC.
* Semicolons are optional Most statements are expressions, including if/else;
* this also means implicit
    returns (without the `return` keyword) are possible
* No top level mutable variables, because all top level variables are exported
* Parens and braces are optional in `for`, `foreach`, and `if` expressions, as
    long as what would be between them is only one expression (would normally be
    typed on one line)
* No ternary expressions, switch statements, or pattern matching; if statements
    are expressions and type-checking is dynamic, so there's no need for extra
    keywords or syntax
* REPL history is stored at `$HOME/.cozy_history`, and the size (in lines) can
    be configured with the env var `COZY_HISTSIZE`

### Builtins

Global functions:

* `error` creates a new error object
* `import` imports another cozy file as a module
* `panic` prints an error contents and exits
* `print` Write values to STDOUT with newlines

Builtin modules (see examples for docs):

* `core`
* `fs`
* `http`
* `json`
* `math`
* `net`
* `sys`
* `time`
* `util`

See also the standard library (written mostly in cozy itself).

## License

This code is licensed [LGPL-3.0](./LICENSE.md). I've used code from various
Monkey implementations, which are usually licensed MIT. Other code used in cozy
includes comments with notes on the licenses.
