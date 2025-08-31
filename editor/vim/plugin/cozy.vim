" don't spam the user when Vim is started in Vi compatibility mode
let s:cpo_save = &cpo
set cpo&vim

" We take care to preserve the user's fileencodings and fileformats,
" because those settings are global (not buffer local), yet we want
" to override them for loading keai files, which are defined to be UTF-8.
let s:current_fileformats = ''
let s:current_fileencodings = ''

" define fileencodings to open as utf-8 encoding even if it's ascii.
function! s:keaifiletype_pre()
    let s:current_fileformats = &g:fileformats
    let s:current_fileencodings = &g:fileencodings
    set fileencodings=utf-8 fileformats=unix
endfunction

" restore fileencodings as others
function! s:keaifiletype_post()
    let &g:fileformats = s:current_fileformats
    let &g:fileencodings = s:current_fileencodings
endfunction

function! s:noop(...) abort
endfunction

augroup vim-keai
    autocmd!

    autocmd BufNewFile *.keai if &modifiable | setlocal fileencoding=utf-8 fileformat=unix | endif
    autocmd BufNewFile .keai_init if &modifiable | setlocal fileencoding=utf-8 fileformat=unix | endif

    autocmd BufRead *.keai call s:keaifiletype_pre()
    autocmd BufRead .keai_init call s:keaifiletype_pre()

    autocmd BufReadPost *.keai call s:keaifiletype_post()
    autocmd BufReadPost .keai_init call s:keaifiletype_post()
augroup end

" restore Vi compatibility settings
let &cpo = s:cpo_save
unlet s:cpo_save
