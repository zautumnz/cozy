# cozy

[![Support with PayPal](https://img.shields.io/badge/paypal-donate-yellow.png)](https://paypal.me/zacanger) [![Patreon](https://img.shields.io/badge/patreon-donate-yellow.svg)](https://www.patreon.com/zacanger) [![ko-fi](https://img.shields.io/badge/donate-KoFi-yellow.svg)](https://ko-fi.com/U7U2110VB)

WIP, see ./TODO.md

---

Example:

```cozy
let reduce = fn(fun, xs, init) {
    mutable acc = init;
    foreach _, x in xs {
        acc = fun(x, acc);
    }

    return acc;
};

let ints? = fn(xs) {
    foreach x in xs {
        if (type(x) != "integer" && type(x) != "float") {
            return false;
        }
    }
    return true;
};

let sum = fn(xs) {
    assert(ints?(xs), "expected only numbers!")
    return reduce(
        fn(x, acc) {
            return x + acc;
        }, xs, 0);
};

print(sum([1, 2, 3, 4]) == 10); # true
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

* `print` adds an ending newline, use `printf` or `STDOUT`/`STDERR` for raw text
* No null/nil, no undefined
* Comments are Python/Shell style
* No switch statements
* `let` and `const` are for declarations (see TODOs about this)
* Using `set` and `delete` on hashes returns a new hash
* `let` is for immutable variables; `mutable` is for mutable ones; this is
    because setting mutable variables should be more annoying to do than
    setting mutable ones.

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
