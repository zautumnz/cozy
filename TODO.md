# TODO

List of things to get done before a v1.

* Major things missing:
    * HTTP server and client
    * Cryptography builtins
    * Errors and exceptions:
        * As values?
        * Try/catch?
        * Move all os.Exit-ing up to top level and pass around exit code?
        * Stack traces? Or at least better file/line indicators?
* Minor things:
    * At least 50% code coverage
    * Confirm that everything under ./examples works
    * Complete all lingering TODOs
    * Write real docs
    * Add argument validation to all functions
    * Improve all error messages
    * Add tab completion and arrow support to the REPL
    * Flesh out the testing library with functions to test failure and errors
    * Remove parens in `if` and `for`
    * Remove `self`?
    * Cozy, memo, and other FP utils
