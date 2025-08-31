if exists("b:current_syntax")
    finish
endif

syn case match

" used in interpolations
syn cluster     keaiEverything      contains=keaiMutable,keaiLet,keaiDeclaration,keaiStatement,keaiConditional,keaiRepeat,keaiBuiltins,keaiBoolean,keaiString,keaiField,keaiSingleDecl,keaiDecimalInt,keaiFloat,keaiOperator,keaiFunction,keaiFunctionCall

syn keyword     keaiImport          import  contained
syn keyword     keaiMutable         mutable contained
syn keyword     keaiLet             let     contained
hi def link     keaiImport          Statement
hi def link     keaiMutable         Keyword
hi def link     keaiLet             Keyword
hi def link     keaiDeclaration     Keyword

" Keywords within functions
syn keyword     keaiStatement         return null
syn keyword     keaiConditional       if else
syn keyword     keaiRepeat            for foreach in
hi def link     keaiStatement         Statement
hi def link     keaiConditional       Conditional
hi def link     keaiRepeat            Repeat

" Predefined functions and values
syn keyword     keaiBuiltins
            \ array
            \ core
            \ error
            \ float
            \ fs
            \ hash
            \ http
            \ import
            \ integer
            \ json
            \ math
            \ net
            \ object
            \ panic
            \ print
            \ string
            \ sys
            \ time
            \ util
syn keyword     keaiBoolean             true false
hi def link     keaiBuiltins            Identifier
hi def link     keaiBoolean             Boolean

" Comments; their contents
syn keyword     keaiTodo              contained TODO
syn cluster     keaiCommentGroup      contains=keaiTodo
syn region      keaiComment           start="#" end="$" contains=@keaiCommentGroup,@Spell
hi def link     keaiComment           Comment
hi def link     keaiTodo              Todo

" Strings and their contents
syn region      keaiString            start=+"+ skip=+\\\\\|\\"+ end=+"+ contains=@Spell
syn region      keaiDocString         start=+'+ skip=+\\\\\|\\'+ end=+'+ contains=@Spell
hi def link keaiString String
hi def link keaiDocString String

" 1. Match a sequence of word characters coming after a '.'
" 2. Require the following but dont match it: ( \@= see :h E59)
"    - The symbols: / - + * %   OR
"    - The symbols: [] {} <> )  OR
"    - The symbols: \n \r space OR
"    - The symbols: , : .
" 3. Have the start of highlight (hs) be the start of matched
"    pattern (s) offsetted one to the right (+1) (see :h E401)
syn match   keaiField   /\.\w\+\
            \%(\%([\/\-\+*%]\)\|\
            \%([\[\]{}<\>\)]\)\|\
            \%([\!=\^|&]\)\|\
            \%([\n\r\ ]\)\|\
            \%([,\:.]\)\)\@=/hs=s+1
hi def link    keaiField              Identifier

" Regions
syn region        keaiParen             start='(' end=')' transparent
syn region        keaiBlock             start="{" end="}" transparent

" import
syn region    keaiImport            start='import (' end=')' transparent contains=keaiImport,keaiString,keaiComment

" mutable, let, and import.
syn match       keaiSingleDecl        /\%(import\|mutable\|let\) [^(]\@=/ contains=keaiImport,keaiMutable,keaiLet

" Integers
syn match       keaiDecimalInt        "\<-\=\(0\|[1-9]_\?\(\d\|\d\+_\?\d\+\)*\)\%([Ee][-+]\=\d\+\)\=\>"

hi def link     keaiDecimalInt        Integer
hi def link     Integer               Number

" Floating point
syn match       keaiFloat             "\<-\=\d\+\.\d*\%([Ee][-+]\=\d\+\)\=\>"
syn match       keaiFloat             "\<-\=\.\d\+\%([Ee][-+]\=\d\+\)\=\>"

hi def link     keaiFloat             Float

" Comments; their contents
syn keyword     keaiTodo              contained NOTE
hi def link     keaiTodo              Todo

syn match keaiMutableArgs /\.\.\./

" Operators;
" match single-char operators:          - + % < > ! & | ^ * =
" and corresponding two-char operators: -= += %= <= >= != &= |= ^= *= ==
syn match keaiOperator /[-+%<>!&|^*=]=\?/
" match / and /=
syn match keaiOperator /\/\%(=\|\ze[^/*]\)/
" match two-char operators:               << >> &^
" and corresponding three-char operators: <<= >>= &^=
syn match keaiOperator /\%(<<\|>>\|&^\)=\?/
" match remaining two-char operators: := && || <- ++ --
syn match keaiOperator /:=\|||\|<-\|++\|--/
" match ...
hi def link     keaiMutableArgs       keaiOperator
hi def link     keaiOperator          Operator

" Functions;
syn match keaiDeclaration       /\<fn\>/ nextgroup=keaiReceiver,keaiFunction,keaiSimpleParams skipwhite skipnl
syn match keaiFunction          /\w\+/ nextgroup=keaiSimpleParams contained skipwhite skipnl
syn match keaiSimpleParams      /(\%(\w\|\_s\|[*\.\[\],\{\}<>-]\)*)/ contained contains=keaiParamName nextgroup=keaiFunctionReturn skipwhite skipnl
syn match keaiFunctionReturn   /(\%(\w\|\_s\|[*\.\[\],\{\}<>-]\)*)/ contained contains=keaiParamName skipwhite skipnl
syn match keaiParamName        /\w\+\%(\s*,\s*\w\+\)*\ze\s\+\%(\w\|\.\|\*\|\[\)/ contained nextgroup=keaiParamName skipwhite skipnl
            \ contains=keaiMutableArgs,keaiBlock
hi def link   keaiReceiverVar    keaiParamName
hi def link   keaiParamName      Identifier
syn match keaiReceiver          /(\s*\w\+\%(\s\+\*\?\s*\w\+\)\?\s*)\ze\s*\w/ contained nextgroup=keaiFunction contains=keaiReceiverVar skipwhite skipnl
hi def link     keaiFunction          Function

" Function calls;
syn match       keaiFunctionCall      /\w\+\ze(/ contains=keaiBuiltins,keaiDeclaration
hi def link     keaiFunctionCall      Type

" Interpolations
syn region keaiInterp       matchgroup=keaiInterpDelim start="{{" end="}}" contained containedin=keaiString contains=@keaiEverything
hi def link keaiInterpDelim Delimiter

" Variable Assignments
syn match keaiMutableAssign /\v[_.[:alnum:]]+(,\s*[_.[:alnum:]]+)*\ze(\s*([-^+|^\/%&]|\*|\<\<|\>\>|\&\^)?\=[^=])/
hi def link   keaiMutableAssign         Special

" Variable Declarations
syn match keaiMutableDefs /\v\w+(,\s*\w+)*\ze(\s*:\=)/
hi def link   keaiMutableDefs           Special

function! s:hi()
    hi def link keaiSameId Search
    hi def link keaiDiagnosticError SpellBad
    hi def link keaiDiagnosticWarning SpellRare

    if has('textprop')
        if empty(prop_type_get('keaiSameId'))
            call prop_type_add('keaiSameId', {'highlight': 'keaiSameId'})
        endif
        if empty(prop_type_get('keaiDiagnosticError'))
            call prop_type_add('keaiDiagnosticError', {'highlight': 'keaiDiagnosticError'})
        endif
        if empty(prop_type_get('keaiDiagnosticWarning'))
            call prop_type_add('keaiDiagnosticWarning', {'highlight': 'keaiDiagnosticWarning'})
        endif
    endif
endfunction

augroup vim-keai-hi
    autocmd!
    autocmd ColorScheme * call s:hi()
augroup end
call s:hi()

syn sync minlines=200

let b:current_syntax = "keai"
