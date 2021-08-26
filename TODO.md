# TODO

# Remaining v1 Work

* Bugs:
    * Spread isn't quite right, see the comment in curry in the stdlib
    * Vim config bugs:
        * Function defs in syntax, see comment there
        * Comments aren't indented when using `>>`/`<<` and `=`
    * Ctags config:
        * Identifiers can be unicode, and also can include dots
    * More than one level of index access doesn't seem to always work on
        non-standard objects (like files)
    * Assertions are somehow creating globals? Reproduction:
        * `let a = 1; a` in REPL. Same with `let b`... see assertions that have
          inline IIFEs
* Features:
    * http.client: add form support
* Chores:
    * Confirm that everything under ./examples works
    * Add argument validation to all internal functions and stdlib
    * Improve all Go error messages

## Possible Future Features

* Allow listing empty root-level modules and non-object modules such as http and
    fs using just the root word (`http` or `fs`).
* Consider changing how module exports work to allow top-level (but still
    non-exported) mutable variables; maybe a new keyword (capital letters aren't
    an option because we allow unicode identifiers)
* Utility like Node's `__filename` (which can also be used to get dirname)
* Allow ANSI escape codes (for colors and whatnot) in print calls
* Change import, http.server, and other paths to allow relative paths/from the
    cozy file being executed
* Add basic module management: some kind of module manifest, vcs manager, and
    automatic COZY_PATH modification
* Add option to compile a program (along with cozy itself) to a binary
* 80%+ code coverage
* Nested interpolations
* Add tab-completion to the REPL
* Maybe combine float/integer to just one number type?
* Move as much of the stdlib into cozy (out of Go) as possible
* Full-featured examples:
    * Twitter/Tumblr clone
    * Ranger clone
    * Text editor
* Date object or additions to core time module
* Markdown parser
* Simplify registerBuiltin calls so they can be looped over
* Cryptography builtins: GUID, hashes, AES, RSA, crypto/rand, etc.
* YAML support
* TOML support
* Websocket support
* Multiple-db ORM
