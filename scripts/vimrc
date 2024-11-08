" Common settings for both vi and vim
set noerrorbells
set vb t_vb=
set mouse=a
set path+=**
let mapleader = " "
syntax on
set incsearch
set ignorecase
set smartcase
set noswapfile
set signcolumn=no
set wildignorecase

set grepformat=%f:%l:%c:%m

" Configure grepprg to use rg if available, otherwise fall back to grep
if executable('rg')
    set grepprg=rg\ --vimgrep\ --smart-case\ --no-heading
else
    set grepprg=grep\ -rI\ -n\ -i
endif

" Mapping for <Leader>fg to prompt for input and open the quickfix list

" Key mappings compatible with vi
" Adjust or remove mappings not supported in vi as needed
" (Note: vi has limited support for custom mappings)

" Vim-specific settings
if v:progname =~? 'vim'
    " Vim-only options
    set undofile
    let &undodir = expand('$HOME/.vim/undodir_vim')
    set signcolumn=yes
    set tabstop=4
    set softtabstop=4
    set shiftwidth=4
    set textwidth=79
    set expandtab
    set autoindent
    set showmatch
    set updatetime=50
    set background=dark
    set numberwidth=4

    " Key mappings
    nnoremap <leader>e :find 
    nnoremap <leader><Space> :Explore<CR>
    nnoremap <C-d> <C-d>zz
    nnoremap <C-u> <C-u>zz
    nnoremap <leader>s :%s/\<<C-r><C-w>\>/<C-r><C-w>/gI<Left><Left><Left>
    nnoremap <leader>x :!chmod +x %<CR>
    nnoremap <leader>fr :bro ol<CR>
    nnoremap <leader>y "+y
    vnoremap <leader>y "+y
    nnoremap <leader>dd "+dd
    nnoremap <leader>Y "+Y
    vnoremap J :m '>+1<CR>gv=gv
    vnoremap K :m '<-2<CR>gv=gv
    nmap <C-p> mzyyP`z
    nnoremap <leader>o :execute 'tcd ' . fnameescape(expand('%:h'))<CR>
    nnoremap <M-e> :marks ABCDEFGHIJKLNOPQRSTUVWXYZ<CR>
    nnoremap <M-a> :marks abcdefghijklnopqrstuvwxyz<CR>
    nnoremap <Leader>cc :make 
       " Paste mode configuration
    function! XTermPasteBegin()
        set pastetoggle=<Esc>[201~
        set paste
        return ""
    endfunction

    inoremap <special> <expr> <Esc>[201~ XTermPasteBegin()

    " Auto-closing pairs
    inoremap ( ()<Left>
    inoremap [ []<Left>
    inoremap { {}<Left>
    inoremap ' ''<Left>
    inoremap " ""<Left>

    " File-specific settings
    autocmd BufWritePre *.py,*.f90,*.f95,*.for :%s/\s\+$//e

    " Cursor configuration
    let &t_SI = "\e[6 q"
    let &t_EI = "\e[2 q"

    " Keybinds for buffer management
" <Leader>ba - List all buffers except quickfix buffers
nnoremap <Leader>ba :ls<CR>

" <Leader>bn - Go to the next buffer
nnoremap <Leader>bn :bnext<CR>

" <Leader>bp - Go to the previous buffer
nnoremap <Leader>bp :bprevious<CR>

    " Custom Functions
nnoremap <leader>fg :call SearchWithGrep()<CR>
function! SearchWithGrepSilent()
    let search_term = input('Grep for: ')
    if search_term != ''
        execute 'silent! grepadd ' . shellescape(search_term)
        copen
    endif
endfunction

function! SearchWithGrep()
    " Prompt for a search term
    let search_term = input('Grep for: ')
    if search_term != ''
        " Run grep with the specified term and open the quickfix list
        execute 'silent! grep! ' . search_term
        copen
    endif
endfunction

    function! RunOnVisual() range
        let saved_reg = getreg('a')
        normal! gv"ay
        let selected_text = getreg('a')
        call setreg('a', saved_reg)
        let command = input('Command to run: ', '', 'shellcmd')
        if empty(command)
            return
        endif
        let output = system(command, selected_text)
        let output = substitute(output, '\n$', '', '')
        execute 'normal! gv"_c' . escape(output, '\|')
    endfunction

    xnoremap f :<C-U>call RunOnVisual()<CR>

    function! LimitOldFiles(max_length)
        let oldfiles = v:oldfiles
        if len(oldfiles) > a:max_length
            let v:oldfiles = oldfiles[0:a:max_length - 1]
        endif
    endfunction

    autocmd VimEnter * call LimitOldFiles(7)

    function! SetZellnerColorscheme()
        if !empty(globpath(&runtimepath, "colors/zellner.vim"))
            colorscheme zellner
        else
            colorscheme default
        endif
    endfunction

command! CloseQuickfixBuffers call CloseQuickfixBuffers()

nnoremap <Leader>cq :CloseQuickfixBuffers<CR>


function! CloseQuickfixBuffers()
    " Iterate over all buffers
    for buf in range(1, bufnr('$'))
        " Check if the buffer type is quickfix or location list
        if getbufvar(buf, '&buftype') ==# 'quickfix'
            " Close the quickfix buffer
            execute 'bdelete' buf
        endif
    endfor
endfunction
    " Initialize colorscheme
    call SetZellnerColorscheme()

    " Colors and Highlights
    hi VertSplit ctermfg=black ctermbg=black guifg=black guibg=black
    hi StatusLine ctermbg=8 ctermfg=0 guibg=#808080 guifg=black
    hi StatusLineNC ctermbg=8 ctermfg=7 guibg=#808080 guifg=darkgrey
    highlight LineNr ctermbg=black ctermfg=8 guibg=#000000 guifg=#808080
    highlight CursorLineNr ctermbg=black ctermfg=8 guibg=#000000 guifg=#808080
    highlight SignColumn ctermbg=NONE guibg=NONE

    " Catppuccin Mocha Theme Highlights
    hi Normal          ctermbg=0       ctermfg=7       guibg=#1e1e2e   guifg=#cdd6f4
    hi Pmenu           ctermbg=8       ctermfg=7       guibg=#313244   guifg=#cdd6f4
    hi PmenuSel        ctermbg=0       ctermfg=7       guibg=#45475a   guifg=#cdd6f4
    hi Visual          ctermbg=8       ctermfg=7       guibg=#45475a   guifg=NONE
    hi Comment         ctermbg=NONE    ctermfg=8       guibg=NONE      guifg=#6c7086
    hi String          ctermbg=NONE    ctermfg=2       guibg=NONE      guifg=#a6e3a1
    hi Function        ctermbg=NONE    ctermfg=4       guibg=NONE      guifg=#89b4fa
    hi Identifier      ctermbg=NONE    ctermfg=3       guibg=NONE      guifg=#cba6f7
    hi Keyword         ctermbg=NONE    ctermfg=5       guibg=NONE      guifg=#f5c2e7
    hi Type            ctermbg=NONE    ctermfg=6       guibg=NONE      guifg=#fab387
    hi Constant        ctermbg=NONE    ctermfg=1       guibg=NONE      guifg=#f38ba8
    hi QuickFixLine    ctermbg=5       ctermfg=0       guibg=#cba6f7   guifg=#1e1e2e
    hi ErrorMsg        ctermbg=1       ctermfg=15      guibg=#f38ba8   guifg=#cdd6f4
    hi WarningMsg      ctermbg=3       ctermfg=0       guibg=#fab387   guifg=#1e1e2e
    hi MoreMsg         ctermbg=2       ctermfg=0       guibg=#a6e3a1   guifg=#1e1e2e
    hi StatusLine      ctermbg=0       ctermfg=8       guibg=#181825   guifg=#bac2de
    hi StatusLineNC    ctermbg=0       ctermfg=8       guibg=#11111b   guifg=#6c7086

    
endif

