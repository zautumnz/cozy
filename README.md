# cozy

[![Support with PayPal](https://img.shields.io/badge/paypal-donate-yellow.png)](https://paypal.me/zacanger) [![Patreon](https://img.shields.io/badge/patreon-donate-yellow.svg)](https://www.patreon.com/zacanger) [![ko-fi](https://img.shields.io/badge/donate-KoFi-yellow.svg)](https://ko-fi.com/U7U2110VB)

---

Example from the ./stdlib.cz:

```
let array.reduce = fn(fun, init) {
    let acc = init;
    foreach _, x in self {
        acc = fun(x, acc);
    }

    return acc;
};

let sum = fn(xs) {
    return xs.reduce(
        fn(x, acc) {
            return x + acc;
        }, 0);
};
```

## About

Simple, high-ish-level programming language that sits somewhere between
scripting and general-purpose programming. Dynamically and strongly typed, with
some with semantics that work well with pseudo-functional programming but syntax
similar to Python, Go, and Shell. No OOP constructs like classes.

Originally designed by writing a bunch of examples and a small stdlib;
implementation started as a fork from [skx's
version](https://github.com/skx/monkey) of the language from the [Go
Interpreters Book](https://interpreterbook.com). Written in pure Go with no
third-party dependencies.

See the ./examples directory for how it all looks. There are also syntax files
for Emacs (untested) and Vim (WIP) in the ./editor directory.

## Usage

Clone the repo and run `make`, and either copy the binary to somewhere in your
path or run `make install`. Write some code (see the examples), and run `cozy
./your-code.cz`. You can also run without a specified file, in which case your
entered code will be evaluated when you exit with `ctrl+d`.

## Important Notes

* `print` adds an ending newline
* No null/nil, no undefined
* Comments are Python/Shell style
* No switch statements
* `let` and `const` are for declarations (see TODOs about this)
* Using `set` and `delete` on hashes returns a new hash

### Builtin Functions

The core primitives are:

* `delete` Deletes a hash-key.
* `int` convert the given float/string to an integer.
* `keys` Return the keys of the specified array.
* `len` Yield the length of builtin containers.
* `match` Regular-expression matching.
* `push` push an elements into the array.
* `print` Write literal value of objects to STDOUT.
* `printf` Write values to STDOUT, via a format-string.
* `set` insert key value pair into the map.
* `sprintf` Create strings, via a format-string.
* `string` convert the given item to a string.
* `type` returns the type of a variable.

Many more functions are defined in the stdlib. See that file for details because
it's always growing.

## TODO

* Major things missing:
    * Async/futures/generators
    * Timers
    * Concurrency
    * Modules
    * Garbage collection
    * Generally make it look more like the initial example code
    * Automatic semicolon insertion
    * JSON, YAML, and TOML built-in support
    * Networking builtins
    * OS/sys/process builtins
    * Cryptography builtins
    * Rework the whole REPL to make it usable
    * Time and date
* Minor things:
    * Variadic arguments
    * Improve core assertion library
    * Add a TAP-compatible testing library on top of `assert`
    * Remove parens in `if` conditions
    * Remove `self`
    * Change declaration keywords to make mutable variables explicit
    * Curry, memo, and other FP utils
    * Docstrings
    * Improve Vim and Emacs files
