if exists("b:current_syntax")
    finish
endif

syn case match

" used in interpolations
syn cluster     cozyEverything      contains=cozyMutable,cozyLet,cozyDeclaration,cozyStatement,cozyConditional,cozyRepeat,cozyBuiltins,cozyBoolean,cozyString,cozyField,cozySingleDecl,cozyDecimalInt,cozyFloat,cozyOperator,cozyFunction,cozyFunctionCall

syn keyword     cozyImport          import  contained
syn keyword     cozyMutable         mutable contained
syn keyword     cozyLet             let     contained
hi def link     cozyImport          Statement
hi def link     cozyMutable         Keyword
hi def link     cozyLet             Keyword
hi def link     cozyDeclaration     Keyword

" Keywords within functions
syn keyword     cozyStatement         return null
syn keyword     cozyConditional       if else
syn keyword     cozyRepeat            for foreach in
hi def link     cozyStatement         Statement
hi def link     cozyConditional       Conditional
hi def link     cozyRepeat            Repeat

" Predefined functions and values
syn keyword     cozyBuiltins
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
syn keyword     cozyBoolean             true false
hi def link     cozyBuiltins            Identifier
hi def link     cozyBoolean             Boolean

" Comments; their contents
syn keyword     cozyTodo              contained TODO
syn cluster     cozyCommentGroup      contains=cozyTodo
syn region      cozyComment           start="#" end="$" contains=@cozyCommentGroup,@Spell
hi def link     cozyComment           Comment
hi def link     cozyTodo              Todo

" Strings and their contents
syn region      cozyString            start=+"+ skip=+\\\\\|\\"+ end=+"+ contains=@Spell
syn region      cozyDocString         start=+'+ skip=+\\\\\|\\'+ end=+'+ contains=@Spell
hi def link cozyString String
hi def link cozyDocString String

" Fields;
if cozy()
  " 1. Match a sequence of word characters coming after a '.'
  " 2. Require the following but dont match it: ( \@= see :h E59)
  "    - The symbols: / - + * %   OR
  "    - The symbols: [] {} <> )  OR
  "    - The symbols: \n \r space OR
  "    - The symbols: , : .
  " 3. Have the start of highlight (hs) be the start of matched
  "    pattern (s) offsetted one to the right (+1) (see :h E401)
  syn match       cozyField   /\.\w\+\
        \%(\%([\/\-\+*%]\)\|\
        \%([\[\]{}<\>\)]\)\|\
        \%([\!=\^|&]\)\|\
        \%([\n\r\ ]\)\|\
        \%([,\:.]\)\)\@=/hs=s+1
endif
hi def link    cozyField              Identifier


" Regions
syn region        cozyParen             start='(' end=')' transparent
syn region        cozyBlock             start="{" end="}" transparent

" import
syn region    cozyImport            start='import (' end=')' transparent contains=cozyImport,cozyString,cozyComment

" mutable, let, and import.
syn match       cozySingleDecl        /\%(import\|mutable\|let\) [^(]\@=/ contains=cozyImport,cozyMutable,cozyLet

" Integers
syn match       cozyDecimalInt        "\<-\=\(0\|[1-9]_\?\(\d\|\d\+_\?\d\+\)*\)\%([Ee][-+]\=\d\+\)\=\>"

hi def link     cozyDecimalInt        Integer
hi def link     Integer               Number

" Floating point
syn match       cozyFloat             "\<-\=\d\+\.\d*\%([Ee][-+]\=\d\+\)\=\>"
syn match       cozyFloat             "\<-\=\.\d\+\%([Ee][-+]\=\d\+\)\=\>"

hi def link     cozyFloat             Float

" Comments; their contents
syn keyword     cozyTodo              contained NOTE
hi def link     cozyTodo              Todo

syn match cozyMutableArgs /\.\.\./

" Operators;
" match single-char operators:          - + % < > ! & | ^ * =
" and corresponding two-char operators: -= += %= <= >= != &= |= ^= *= ==
syn match cozyOperator /[-+%<>!&|^*=]=\?/
" match / and /=
syn match cozyOperator /\/\%(=\|\ze[^/*]\)/
" match two-char operators:               << >> &^
" and corresponding three-char operators: <<= >>= &^=
syn match cozyOperator /\%(<<\|>>\|&^\)=\?/
" match remaining two-char operators: := && || <- ++ --
syn match cozyOperator /:=\|||\|<-\|++\|--/
" match ...
hi def link     cozyMutableArgs       cozyOperator
hi def link     cozyOperator          Operator

" Functions;
" TODO: due to order I think this is highlighting function names as generic
" identifiers rather than functions, expecting
" `fn foo(x)`, but it needs to match `let foo = fn(x)`.
syn match cozyDeclaration       /\<fn\>/ nextgroup=cozyReceiver,cozyFunction,cozySimpleParams skipwhite skipnl
syn match cozyFunction          /\w\+/ nextgroup=cozySimpleParams contained skipwhite skipnl
syn match cozySimpleParams      /(\%(\w\|\_s\|[*\.\[\],\{\}<>-]\)*)/ contained contains=cozyParamName nextgroup=cozyFunctionReturn skipwhite skipnl
syn match cozyFunctionReturn   /(\%(\w\|\_s\|[*\.\[\],\{\}<>-]\)*)/ contained contains=cozyParamName skipwhite skipnl
syn match cozyParamName        /\w\+\%(\s*,\s*\w\+\)*\ze\s\+\%(\w\|\.\|\*\|\[\)/ contained nextgroup=cozyParamName skipwhite skipnl
            \ contains=cozyMutableArgs,cozyBlock
hi def link   cozyReceiverVar    cozyParamName
hi def link   cozyParamName      Identifier
syn match cozyReceiver          /(\s*\w\+\%(\s\+\*\?\s*\w\+\)\?\s*)\ze\s*\w/ contained nextgroup=cozyFunction contains=cozyReceiverVar skipwhite skipnl
hi def link     cozyFunction          Function

" Function calls;
syn match       cozyFunctionCall      /\w\+\ze(/ contains=cozyBuiltins,cozyDeclaration
hi def link     cozyFunctionCall      Type

" Interpolations
syn region cozyInterp       matchgroup=cozyInterpDelim start="{{" end="}}" contained containedin=cozyString contains=@cozyEverything
hi def link cozyInterpDelim Delimiter

" Variable Assignments
syn match cozyMutableAssign /\v[_.[:alnum:]]+(,\s*[_.[:alnum:]]+)*\ze(\s*([-^+|^\/%&]|\*|\<\<|\>\>|\&\^)?\=[^=])/
hi def link   cozyMutableAssign         Special

" Variable Declarations
syn match cozyMutableDefs /\v\w+(,\s*\w+)*\ze(\s*:\=)/
hi def link   cozyMutableDefs           Special

function! s:hi()
    hi def link cozySameId Search
    hi def link cozyDiagnosticError SpellBad
    hi def link cozyDiagnosticWarning SpellRare

    if has('textprop')
        if empty(prop_type_get('cozySameId'))
            call prop_type_add('cozySameId', {'highlight': 'cozySameId'})
        endif
        if empty(prop_type_get('cozyDiagnosticError'))
            call prop_type_add('cozyDiagnosticError', {'highlight': 'cozyDiagnosticError'})
        endif
        if empty(prop_type_get('cozyDiagnosticWarning'))
            call prop_type_add('cozyDiagnosticWarning', {'highlight': 'cozyDiagnosticWarning'})
        endif
    endif
endfunction

augroup vim-cozy-hi
    autocmd!
    autocmd ColorScheme * call s:hi()
augroup end
call s:hi()

syn sync minlines=200

let b:current_syntax = "cozy"
