# TODO

v1 work:
* Errors and exceptions: As values? Try/catch? Stack traces?
* At least 50% code coverage
* Confirm that everything under ./examples works
* Complete all lingering TODOs in the code
* Add argument validation to all internal functions and stdlib
* Improve all Go error messages
* Remove other extraneous things from readline fork
* Find and remove any unused Go code
* Docstrings cause parsing errors, fix that
* Add splat/spread operator and finish curry fn
* Possibly change the ... literal to return a regular array
* Remove requirement for braces in if/else expressions
* Nested interpolations
* Assertions are somehow creating globals? Reproduction:
    * `let a = 1; a` in REPL. Same with `let b`... see assertions that have
        inline IIFEs

v2 work:
* Move as much of the stdlib into cozy (out of Go) as possible
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
