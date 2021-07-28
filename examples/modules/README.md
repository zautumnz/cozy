Every cozy file with the extension `.cz` is a module. `cwd` is considered the
module root, as can be seen in the import statements here which assume you're
running cozy from this repo's root. This can be changed with the environment
variable `COZYPATH`. There is no package management, but using `COZYPATH` with
Git submodules is a pretty obvious way to go. All top-level variables are
exported (and `mutable` variables are not allowed to be top level). Most of the
module code is taken directly from github.com/prologic/monkey-lang (MIT
licensed), with some modifications to make it work in this version of the
language.
