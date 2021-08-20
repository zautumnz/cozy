# TODO

# Remaining v1 Work

* Bugs:
    * Spread isn't quite right, see the comment in curry in the stdlib
    * Vim syntax needs work: function definitions
    * More than one level of dot access doesn't seem to always work
    * Assertions are somehow creating globals? Reproduction:
        * `let a = 1; a` in REPL. Same with `let b`... see assertions that have
          inline IIFEs
* Features:
    * http.server: add form support
    * http.client: add form support
    * Utility like Node's `__filename` (which can also be used to get dirname)
* Chores:
    * Clean up newerror and similar calls now that errors are useful values
    * Confirm that everything under ./examples works
    * Add argument validation to all internal functions and stdlib
    * Improve all Go error messages

## Possible Future Features

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
        * Guid
        * crypto/rand
        * common hashes
        * aes stuff
    * YAML support
    * TOML support
    * Websocket support
    * Multiple-db ORM
