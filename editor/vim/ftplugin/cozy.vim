" cozy.vim: Vim filetype plugin for Cozy.

if exists("b:did_ftplugin")
  finish
endif
let b:did_ftplugin = 1

" don't spam the user when Vim is started in Vi compatibility mode
let s:cpo_save = &cpo
set cpo&vim

let b:undo_ftplugin = "setl fo< com< cms< tw< tabstop< softtabtop< sw< et< smartindent< smarttab< autoindent<"

setlocal comments=:#
setlocal commentstring=#%s

setlocal tw=80
setlocal tabstop=4
setlocal softtabstop=4
setlocal shiftwidth=4
setlocal expandtab
setlocal smartindent
setlocal smarttab
setlocal autoindent

" restore Vi compatibility settings
let &cpo = s:cpo_save
unlet s:cpo_save
