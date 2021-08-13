# TODO

v1 work:
* Errors and exceptions: As values? Try/catch? Stack traces?
* At least 50% code coverage
* More than one level of dot access doesn't seem to always work
* Confirm that everything under ./examples works
* Complete all lingering TODOs in the code
* Add argument validation to all internal functions and stdlib
* Improve all Go error messages
* Remove other extraneous things from readline fork
* Find and remove any unused Go code
* Remove requirement for braces in if/else expressions
* Nested interpolations
* Maybe combine float/integer to just one number type?
* Convenience method for creating new hashes
* Assertions are somehow creating globals? Reproduction:
    * `let a = 1; a` in REPL. Same with `let b`... see assertions that have
        inline IIFEs

v2 work:
* Markdown parser
* Move as much of the stdlib into cozy (out of Go) as possible
* Optionally indent json in json.serialize
* Simplify registerBuiltin calls so they can be looped over
* Microblogging site real-life example, or something similar in scope
* Cryptography builtins
    * Guid
    * crypto/rand
    * common hashes
    * aes stuff
* YAML support
* TOML support
* Multiple-db ORM
* Add some properties to functions (and to all objects?):
    * name
    * maybe arguments as a named array, like es5
