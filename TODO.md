# TODO

# Remaining v1 Work

* Bugs:
    * Spread isn't quite right, see the comment in curry in the stdlib
    * Vim config bugs:
        * Function defs in syntax, see comment there
        * Comments aren't indented when using `>>`/`<<` and `=`
    * More than one level of dot access doesn't seem to always work
    * Assertions are somehow creating globals? Reproduction:
        * `let a = 1; a` in REPL. Same with `let b`... see assertions that have
          inline IIFEs
    * Allow listing empty root-level modules such as http (http itself isn't
        anything, but http.server, etc, are)
    * Evaluator errors are getting printed twice
* Features:
    * http.client: add form support
* Chores:
    * Clean up newerror and similar calls now that errors are useful values
    * Confirm that everything under ./examples works
    * Add argument validation to all internal functions and stdlib
    * Improve all Go error messages

## Possible Future Features

* Consider changing how module exports work to allow top-level (but still
    non-exported) mutable variables; maybe a new keyword (capital letters aren't
    an option because we allow unicode identifiers)
* Utility like Node's `__filename` (which can also be used to get dirname)
* Allow ANSI escape codes (for colors and whatnot) in print calls
* Change import, http.server, and other paths to allow relative paths/from the
    cozy file being executed
* Add basic module management: some kind of module manifest, vcs manager, and
    automatic COZYPATH modification
* Add option to compile a program (along with cozy itself) to a binary
* 80%+ code coverage
* Nested interpolations
* Add tab-completion to the REPL
* Maybe combine float/integer to just one number type?
* Move as much of the stdlib into cozy (out of Go) as possible
* Write in cozy stdlib:
    * Date object or additions to core time module
    * Markdown parser
    * Simplify registerBuiltin calls so they can be looped over
    * Microblogging site real-life example, or something similar in scope
    * Cryptography builtins
        * GUID
        * crypto/rand
        * common hashes
        * AES
        * RSA
    * YAML support
    * TOML support
    * Websocket support
    * Multiple-db ORM
