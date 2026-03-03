-- ╔══════════════════════════════════════════════════════════════╗
-- ║  Neovim — sumi Monochrome Config + lazy.nvim plugins     ║
-- ╚══════════════════════════════════════════════════════════════╝

-- ── Options ─────────────────────────────────────────────────
vim.g.mapleader = " "
vim.g.maplocalleader = " "

local opt = vim.opt
opt.number = true
opt.relativenumber = true
opt.signcolumn = "yes"
opt.cursorline = true
opt.scrolloff = 8
opt.sidescrolloff = 8

opt.tabstop = 4
opt.shiftwidth = 4
opt.expandtab = true
opt.smartindent = true

opt.wrap = false
opt.linebreak = true
opt.breakindent = true

opt.ignorecase = true
opt.smartcase = true
opt.hlsearch = true
opt.incsearch = true

opt.splitright = true
opt.splitbelow = true

opt.termguicolors = true
opt.background = "dark"
opt.showmode = false
opt.cmdheight = 1
opt.laststatus = 3

opt.undofile = true
opt.swapfile = false
opt.backup = false

opt.updatetime = 250
opt.timeoutlen = 300

opt.clipboard = "unnamedplus"
opt.mouse = "a"
opt.completeopt = "menuone,noselect"
opt.list = true
opt.listchars = { tab = "» ", trail = "·", nbsp = "␣" }

-- ── Bootstrap lazy.nvim ───────────────────────────────────────
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
    vim.fn.system({
        "git", "clone", "--filter=blob:none",
        "https://github.com/folke/lazy.nvim.git",
        "--branch=stable", lazypath,
    })
end
vim.opt.rtp:prepend(lazypath)

-- ── Plugins ───────────────────────────────────────────────────
require("lazy").setup({
    -- Treesitter (syntax highlighting)
    {
        "nvim-treesitter/nvim-treesitter",
        build = ":TSUpdate",
        config = function()
            require("nvim-treesitter.configs").setup({
                ensure_installed = {
                    "bash", "c", "css", "html", "json", "lua",
                    "markdown", "python", "rust", "toml", "yaml",
                    "javascript", "typescript", "go",
                },
                highlight = { enable = true },
                indent = { enable = true },
            })
        end,
    },

    -- Telescope (fuzzy finder)
    {
        "nvim-telescope/telescope.nvim",
        branch = "0.1.x",
        dependencies = { "nvim-lua/plenary.nvim" },
        config = function()
            local telescope = require("telescope")
            local actions = require("telescope.actions")
            telescope.setup({
                defaults = {
                    layout_strategy = "vertical",
                    layout_config = { height = 0.8, width = 0.7 },
                    border = true,
                    borderchars = { "─", "│", "─", "│", "┌", "┐", "┘", "└" },
                    prompt_prefix = "> ",
                    selection_caret = "│ ",
                    mappings = {
                        i = {
                            ["<C-j>"] = actions.move_selection_next,
                            ["<C-k>"] = actions.move_selection_previous,
                            ["<Esc>"] = actions.close,
                        },
                    },
                },
            })
        end,
    },

    -- Gitsigns (git gutters)
    {
        "lewis6991/gitsigns.nvim",
        config = function()
            require("gitsigns").setup({
                signs = {
                    add          = { text = "│" },
                    change       = { text = "│" },
                    delete       = { text = "_" },
                    topdelete    = { text = "‾" },
                    changedelete = { text = "~" },
                },
            })
        end,
    },

    -- Comment.nvim (gc to comment)
    { "numToStr/Comment.nvim", config = true },

    -- Autopairs
    { "windwp/nvim-autopairs", event = "InsertEnter", config = true },

    -- Surround (ys/ds/cs)
    { "kylechui/nvim-surround", event = "VeryLazy", config = true },

    -- Indent guides
    {
        "lukas-reineke/indent-blankline.nvim",
        main = "ibl",
        config = function()
            require("ibl").setup({
                indent = { char = "│" },
                scope = { enabled = true, show_start = false, show_end = false },
            })
        end,
    },

    -- Which-key (shows pending keybinds)
    {
        "folke/which-key.nvim",
        event = "VeryLazy",
        config = function()
            require("which-key").setup({
                win = { border = "single" },
            })
        end,
    },

    -- ── LSP ─────────────────────────────────────────────────────

    -- Mason (LSP/DAP/linter/formatter installer)
    {
        "williamboman/mason.nvim",
        config = function()
            require("mason").setup({
                ui = {
                    border = "single",
                    icons = { package_installed = "+", package_pending = "~", package_uninstalled = "-" },
                },
            })
        end,
    },

    -- Mason-lspconfig bridge
    {
        "williamboman/mason-lspconfig.nvim",
        dependencies = { "williamboman/mason.nvim", "neovim/nvim-lspconfig" },
        config = function()
            require("mason-lspconfig").setup({
                ensure_installed = {
                    "lua_ls", "pyright", "rust_analyzer", "ts_ls",
                    "bashls", "cssls", "html", "jsonls", "gopls",
                    "clangd", "taplo",
                },
                automatic_installation = true,
            })
        end,
    },

    -- LSP config
    {
        "neovim/nvim-lspconfig",
        dependencies = { "williamboman/mason-lspconfig.nvim" },
        config = function()
            local lspconfig = require("lspconfig")
            local capabilities = vim.lsp.protocol.make_client_capabilities()

            -- Try to enhance capabilities with cmp if available
            local ok_cmp, cmp_lsp = pcall(require, "cmp_nvim_lsp")
            if ok_cmp then
                capabilities = cmp_lsp.default_capabilities(capabilities)
            end

            -- On-attach keybinds (only active when LSP connects)
            local on_attach = function(_, bufnr)
                local bmap = function(mode, lhs, rhs, desc)
                    vim.keymap.set(mode, lhs, rhs, { buffer = bufnr, desc = desc })
                end
                bmap("n", "gd", vim.lsp.buf.definition, "Go to definition")
                bmap("n", "gD", vim.lsp.buf.declaration, "Go to declaration")
                bmap("n", "gr", vim.lsp.buf.references, "References")
                bmap("n", "gi", vim.lsp.buf.implementation, "Implementation")
                bmap("n", "gy", vim.lsp.buf.type_definition, "Type definition")
                bmap("n", "K", vim.lsp.buf.hover, "Hover docs")
                bmap("n", "<leader>ca", vim.lsp.buf.code_action, "Code action")
                bmap("n", "<leader>cr", vim.lsp.buf.rename, "Rename symbol")
                bmap("n", "<leader>cf", function() vim.lsp.buf.format({ async = true }) end, "Format buffer")
                bmap("n", "<leader>cs", vim.lsp.buf.signature_help, "Signature help")
                bmap("i", "<C-s>", vim.lsp.buf.signature_help, "Signature help")
            end

            -- Setup each server
            local servers = {
                lua_ls = {
                    settings = {
                        Lua = {
                            runtime = { version = "LuaJIT" },
                            workspace = { checkThirdParty = false },
                            telemetry = { enable = false },
                            diagnostics = { globals = { "vim" } },
                        },
                    },
                },
                pyright = {},
                rust_analyzer = {
                    settings = {
                        ["rust-analyzer"] = {
                            checkOnSave = { command = "clippy" },
                        },
                    },
                },
                ts_ls = {},
                bashls = {},
                cssls = {},
                html = {},
                jsonls = {},
                gopls = {},
                clangd = {},
                taplo = {},
            }

            for server, opts in pairs(servers) do
                opts.on_attach = on_attach
                opts.capabilities = capabilities
                lspconfig[server].setup(opts)
            end
        end,
    },

    -- Autocompletion
    {
        "hrsh7th/nvim-cmp",
        event = "InsertEnter",
        dependencies = {
            "hrsh7th/cmp-nvim-lsp",
            "hrsh7th/cmp-buffer",
            "hrsh7th/cmp-path",
            "L3MON4D3/LuaSnip",
            "saadparwaiz1/cmp_luasnip",
        },
        config = function()
            local cmp = require("cmp")
            local luasnip = require("luasnip")
            cmp.setup({
                snippet = {
                    expand = function(args) luasnip.lsp_expand(args.body) end,
                },
                window = {
                    completion = cmp.config.window.bordered({ border = "single" }),
                    documentation = cmp.config.window.bordered({ border = "single" }),
                },
                mapping = cmp.mapping.preset.insert({
                    ["<C-n>"] = cmp.mapping.select_next_item(),
                    ["<C-p>"] = cmp.mapping.select_prev_item(),
                    ["<C-b>"] = cmp.mapping.scroll_docs(-4),
                    ["<C-f>"] = cmp.mapping.scroll_docs(4),
                    ["<C-Space>"] = cmp.mapping.complete(),
                    ["<C-e>"] = cmp.mapping.abort(),
                    ["<CR>"] = cmp.mapping.confirm({ select = false }),
                    ["<Tab>"] = cmp.mapping(function(fallback)
                        if cmp.visible() then
                            cmp.select_next_item()
                        elseif luasnip.expand_or_jumpable() then
                            luasnip.expand_or_jump()
                        else
                            fallback()
                        end
                    end, { "i", "s" }),
                    ["<S-Tab>"] = cmp.mapping(function(fallback)
                        if cmp.visible() then
                            cmp.select_prev_item()
                        elseif luasnip.jumpable(-1) then
                            luasnip.jump(-1)
                        else
                            fallback()
                        end
                    end, { "i", "s" }),
                }),
                sources = cmp.config.sources({
                    { name = "nvim_lsp" },
                    { name = "luasnip" },
                    { name = "path" },
                }, {
                    { name = "buffer" },
                }),
            })
        end,
    },

    -- Conform (formatter)
    {
        "stevearc/conform.nvim",
        event = "BufWritePre",
        cmd = { "ConformInfo" },
        config = function()
            require("conform").setup({
                formatters_by_ft = {
                    lua = { "stylua" },
                    python = { "ruff_format", "black", stop_after_first = true },
                    rust = { "rustfmt" },
                    javascript = { "prettierd", "prettier", stop_after_first = true },
                    typescript = { "prettierd", "prettier", stop_after_first = true },
                    json = { "prettierd", "prettier", stop_after_first = true },
                    html = { "prettierd", "prettier", stop_after_first = true },
                    css = { "prettierd", "prettier", stop_after_first = true },
                    yaml = { "prettierd", "prettier", stop_after_first = true },
                    markdown = { "prettierd", "prettier", stop_after_first = true },
                    go = { "gofmt" },
                    c = { "clang-format" },
                    cpp = { "clang-format" },
                    sh = { "shfmt" },
                    bash = { "shfmt" },
                    toml = { "taplo" },
                    ["_"] = { "trim_whitespace" },
                },
                format_on_save = {
                    timeout_ms = 3000,
                    lsp_format = "fallback",
                },
            })
        end,
    },

    -- Lint (nvim-lint for extra linters beyond LSP)
    {
        "mfussenegger/nvim-lint",
        event = { "BufReadPre", "BufNewFile" },
        config = function()
            local lint = require("lint")
            lint.linters_by_ft = {
                python = { "ruff" },
                sh = { "shellcheck" },
                bash = { "shellcheck" },
            }
            vim.api.nvim_create_autocmd({ "BufWritePost", "BufReadPost", "InsertLeave" }, {
                callback = function() lint.try_lint() end,
            })
        end,
    },

    -- Trouble (diagnostics list)
    {
        "folke/trouble.nvim",
        cmd = { "Trouble" },
        config = function()
            require("trouble").setup({
                icons = false,
                fold_open = "v",
                fold_closed = ">",
                indent_lines = true,
                use_diagnostic_signs = true,
            })
        end,
    },

    -- Todo comments (highlight TODO/FIXME/HACK)
    {
        "folke/todo-comments.nvim",
        dependencies = { "nvim-lua/plenary.nvim" },
        event = "VeryLazy",
        config = function()
            require("todo-comments").setup({
                signs = false,
                highlight = { pattern = [[.*<(KEYWORDS)\s*:?]] },
                search = { pattern = [[\b(KEYWORDS)\b]] },
            })
        end,
    },

    -- Fidget (LSP progress indicator)
    {
        "j-hui/fidget.nvim",
        event = "LspAttach",
        config = function()
            require("fidget").setup({
                notification = {
                    window = { winblend = 0, border = "single" },
                },
            })
        end,
    },
}, {
    ui = {
        border = "single",
        icons = { cmd = ":", config = "cfg", event = "ev", ft = "ft",
                  init = "init", keys = "key", plugin = "plug",
                  runtime = "rt", source = "src", start = "▶",
                  task = "✓", lazy = "…" },
    },
    performance = {
        rtp = {
            disabled_plugins = {
                "gzip", "matchit", "matchparen", "netrwPlugin",
                "tarPlugin", "tohtml", "tutor", "zipPlugin",
            },
        },
    },
})

-- ── Monochrome colorscheme (inline — applied AFTER plugins) ──
local function set_colors()
    local colors = {
        bg       = "#0a0a0a",
        bg1      = "#1a1a1a",
        bg2      = "#2a2a2a",
        bg3      = "#3a3a3a",
        fg       = "#c8c8c8",
        fg_dim   = "#8a8a8a",
        fg_muted = "#6a6a6a",
        accent   = "#7aa2f7",
        red      = "#f7768e",
        green    = "#9ece6a",
        yellow   = "#e0af68",
        cyan     = "#7dcfff",
    }

    local hl = vim.api.nvim_set_hl

    -- Base
    hl(0, "Normal",       { fg = colors.fg, bg = "NONE" })
    hl(0, "NormalFloat",  { fg = colors.fg, bg = colors.bg1 })
    hl(0, "FloatBorder",  { fg = colors.bg3, bg = colors.bg1 })
    hl(0, "CursorLine",   { bg = colors.bg1 })
    hl(0, "CursorLineNr", { fg = colors.accent, bold = true })
    hl(0, "LineNr",       { fg = colors.bg3 })
    hl(0, "SignColumn",   { bg = "NONE" })
    hl(0, "VertSplit",    { fg = colors.bg3 })
    hl(0, "WinSeparator", { fg = colors.bg3 })
    hl(0, "StatusLine",   { fg = colors.fg_dim, bg = colors.bg1 })
    hl(0, "StatusLineNC", { fg = colors.fg_muted, bg = colors.bg1 })
    hl(0, "Pmenu",        { fg = colors.fg, bg = colors.bg1 })
    hl(0, "PmenuSel",     { fg = colors.bg, bg = colors.accent })
    hl(0, "Visual",       { bg = colors.bg2 })
    hl(0, "Search",       { fg = colors.bg, bg = colors.yellow })
    hl(0, "IncSearch",    { fg = colors.bg, bg = colors.accent })
    hl(0, "MatchParen",   { fg = colors.accent, bold = true, underline = true })
    hl(0, "NonText",      { fg = colors.bg3 })
    hl(0, "SpecialKey",   { fg = colors.bg3 })
    hl(0, "Folded",       { fg = colors.fg_muted, bg = colors.bg1 })
    hl(0, "FoldColumn",   { fg = colors.bg3 })
    hl(0, "Directory",    { fg = colors.accent })
    hl(0, "Title",        { fg = colors.fg, bold = true })
    hl(0, "ErrorMsg",     { fg = colors.red })
    hl(0, "WarningMsg",   { fg = colors.yellow })
    hl(0, "WildMenu",     { fg = colors.bg, bg = colors.accent })

    -- Syntax (monochrome with subtle differentiation)
    hl(0, "Comment",      { fg = colors.fg_muted, italic = true })
    hl(0, "Constant",     { fg = colors.fg })
    hl(0, "String",       { fg = colors.fg_dim })
    hl(0, "Number",       { fg = colors.accent })
    hl(0, "Boolean",      { fg = colors.accent })
    hl(0, "Identifier",   { fg = colors.fg })
    hl(0, "Function",     { fg = colors.fg, bold = true })
    hl(0, "Statement",    { fg = colors.fg })
    hl(0, "Keyword",      { fg = colors.fg, bold = true })
    hl(0, "Operator",     { fg = colors.fg_dim })
    hl(0, "PreProc",      { fg = colors.fg_dim })
    hl(0, "Type",         { fg = colors.fg })
    hl(0, "Special",      { fg = colors.accent })
    hl(0, "Delimiter",    { fg = colors.fg_muted })
    hl(0, "Todo",         { fg = colors.bg, bg = colors.accent, bold = true })
    hl(0, "Error",        { fg = colors.red })

    -- Treesitter overrides (monochrome philosophy)
    hl(0, "@keyword",     { fg = colors.fg, bold = true })
    hl(0, "@function",    { fg = colors.fg, bold = true })
    hl(0, "@string",      { fg = colors.fg_dim })
    hl(0, "@number",      { fg = colors.accent })
    hl(0, "@boolean",     { fg = colors.accent })
    hl(0, "@comment",     { fg = colors.fg_muted, italic = true })
    hl(0, "@variable",    { fg = colors.fg })
    hl(0, "@property",    { fg = colors.fg })
    hl(0, "@punctuation", { fg = colors.fg_muted })
    hl(0, "@operator",    { fg = colors.fg_dim })
    hl(0, "@type",        { fg = colors.fg })
    hl(0, "@tag",         { fg = colors.fg, bold = true })
    hl(0, "@tag.attribute", { fg = colors.fg_dim })

    -- Gitsigns
    hl(0, "GitSignsAdd",    { fg = colors.green })
    hl(0, "GitSignsChange", { fg = colors.yellow })
    hl(0, "GitSignsDelete", { fg = colors.red })

    -- Telescope
    hl(0, "TelescopeBorder",  { fg = colors.bg3 })
    hl(0, "TelescopeTitle",   { fg = colors.accent, bold = true })
    hl(0, "TelescopePromptPrefix", { fg = colors.accent })
    hl(0, "TelescopeSelection", { bg = colors.bg2 })
    hl(0, "TelescopeMatching", { fg = colors.accent, bold = true })

    -- Indent-blankline
    hl(0, "IblIndent", { fg = colors.bg2 })
    hl(0, "IblScope",  { fg = colors.bg3 })

    -- Which-key
    hl(0, "WhichKey",      { fg = colors.accent })
    hl(0, "WhichKeyGroup", { fg = colors.fg_dim })
    hl(0, "WhichKeyDesc",  { fg = colors.fg })

    -- Diagnostics
    hl(0, "DiagnosticError", { fg = colors.red })
    hl(0, "DiagnosticWarn",  { fg = colors.yellow })
    hl(0, "DiagnosticInfo",  { fg = colors.accent })
    hl(0, "DiagnosticHint",  { fg = colors.fg_muted })

    -- Diff
    hl(0, "DiffAdd",    { fg = colors.green, bg = "NONE" })
    hl(0, "DiffChange", { fg = colors.yellow, bg = "NONE" })
    hl(0, "DiffDelete", { fg = colors.red, bg = "NONE" })

    -- LSP reference highlighting
    hl(0, "LspReferenceText",  { bg = colors.bg2 })
    hl(0, "LspReferenceRead",  { bg = colors.bg2 })
    hl(0, "LspReferenceWrite", { bg = colors.bg2 })

    -- cmp menu
    hl(0, "CmpItemAbbr",           { fg = colors.fg })
    hl(0, "CmpItemAbbrMatch",      { fg = colors.accent, bold = true })
    hl(0, "CmpItemAbbrMatchFuzzy", { fg = colors.accent })
    hl(0, "CmpItemKind",           { fg = colors.fg_dim })
    hl(0, "CmpItemMenu",           { fg = colors.fg_muted })

    -- Fidget
    hl(0, "FidgetTitle",  { fg = colors.accent })
    hl(0, "FidgetTask",   { fg = colors.fg_muted })

    -- Trouble
    hl(0, "TroubleNormal", { bg = "NONE" })
end

set_colors()

-- Reapply colors after any colorscheme change (e.g. plugins overriding)
vim.api.nvim_create_autocmd("ColorScheme", { callback = set_colors })

-- ── Diagnostics config ──────────────────────────────────────
vim.diagnostic.config({
    virtual_text = { prefix = "│", spacing = 4 },
    signs = true,
    underline = true,
    update_in_insert = false,
    severity_sort = true,
    float = { border = "single", source = true },
})

local signs = { Error = "x", Warn = "!", Hint = ".", Info = "i" }
for type, icon in pairs(signs) do
    local hl = "DiagnosticSign" .. type
    vim.fn.sign_define(hl, { text = icon, texthl = hl, numhl = "" })
end

-- ── Statusline (minimal boxy) ───────────────────────────────
function _G.statusline()
    local mode_map = {
        n = "NRM", i = "INS", v = "VIS", V = "VLN",
        ["\22"] = "VBL", c = "CMD", R = "REP", t = "TRM",
    }
    local mode = mode_map[vim.fn.mode()] or vim.fn.mode()
    local file = vim.fn.expand("%:t")
    if file == "" then file = "[no file]" end
    local modified = vim.bo.modified and " [+]" or ""
    local readonly = vim.bo.readonly and " [ro]" or ""
    local ft = vim.bo.filetype ~= "" and vim.bo.filetype or "none"
    local pos = vim.fn.line(".") .. ":" .. vim.fn.col(".")
    local pct = math.floor(vim.fn.line(".") / vim.fn.line("$") * 100) .. "%%"

    -- Git branch from gitsigns
    local branch = ""
    local gs = vim.b.gitsigns_head
    if gs then branch = " " .. gs .. " │" end

    -- LSP client name
    local lsp = ""
    local clients = vim.lsp.get_clients({ bufnr = 0 })
    if #clients > 0 then lsp = " " .. clients[1].name .. " │" end

    -- Diagnostic counts
    local diag = ""
    local errs = #vim.diagnostic.get(0, { severity = vim.diagnostic.severity.ERROR })
    local warns = #vim.diagnostic.get(0, { severity = vim.diagnostic.severity.WARN })
    if errs > 0 then diag = diag .. " x" .. errs end
    if warns > 0 then diag = diag .. " !" .. warns end
    if diag ~= "" then diag = diag .. " │" end

    return " [" .. mode .. "] │" .. branch .. " " .. file .. modified .. readonly
        .. "%="
        .. diag .. lsp .. " " .. ft .. " │ " .. pos .. " │ " .. pct .. " "
end
vim.o.statusline = "%!v:lua.statusline()"

-- ── Keymaps ─────────────────────────────────────────────────
local map = vim.keymap.set

-- Better window navigation
map("n", "<C-h>", "<C-w>h", { desc = "Window left" })
map("n", "<C-j>", "<C-w>j", { desc = "Window down" })
map("n", "<C-k>", "<C-w>k", { desc = "Window up" })
map("n", "<C-l>", "<C-w>l", { desc = "Window right" })

-- Move lines
map("v", "J", ":m '>+1<CR>gv=gv", { desc = "Move line down" })
map("v", "K", ":m '<-2<CR>gv=gv", { desc = "Move line up" })

-- Keep centered
map("n", "<C-d>", "<C-d>zz")
map("n", "<C-u>", "<C-u>zz")
map("n", "n", "nzzzv")
map("n", "N", "Nzzzv")

-- Clear search
map("n", "<Esc>", "<cmd>nohlsearch<CR>")

-- Quick save
map("n", "<leader>w", "<cmd>w<CR>", { desc = "Save" })
map("n", "<leader>q", "<cmd>q<CR>", { desc = "Quit" })
map("n", "<leader>x", "<cmd>bd<CR>", { desc = "Close buffer" })

-- Buffer navigation
map("n", "<S-h>", "<cmd>bprevious<CR>", { desc = "Prev buffer" })
map("n", "<S-l>", "<cmd>bnext<CR>", { desc = "Next buffer" })

-- Telescope
map("n", "<leader>f", "<cmd>Telescope find_files<CR>", { desc = "Find files" })
map("n", "<leader>/", "<cmd>Telescope live_grep<CR>", { desc = "Live grep" })
map("n", "<leader>b", "<cmd>Telescope buffers<CR>", { desc = "Buffers" })
map("n", "<leader>r", "<cmd>Telescope oldfiles<CR>", { desc = "Recent files" })
map("n", "<leader>h", "<cmd>Telescope help_tags<CR>", { desc = "Help" })
map("n", "<leader>d", "<cmd>Telescope diagnostics<CR>", { desc = "Diagnostics" })

-- Gitsigns
map("n", "]h", "<cmd>Gitsigns next_hunk<CR>", { desc = "Next hunk" })
map("n", "[h", "<cmd>Gitsigns prev_hunk<CR>", { desc = "Prev hunk" })
map("n", "<leader>gp", "<cmd>Gitsigns preview_hunk<CR>", { desc = "Preview hunk" })
map("n", "<leader>gr", "<cmd>Gitsigns reset_hunk<CR>", { desc = "Reset hunk" })
map("n", "<leader>gb", "<cmd>Gitsigns blame_line<CR>", { desc = "Blame line" })

-- Diagnostic navigation
map("n", "[d", vim.diagnostic.goto_prev, { desc = "Prev diagnostic" })
map("n", "]d", vim.diagnostic.goto_next, { desc = "Next diagnostic" })
map("n", "<leader>e", vim.diagnostic.open_float, { desc = "Show diagnostic" })

-- Trouble
map("n", "<leader>xx", "<cmd>Trouble diagnostics toggle<CR>", { desc = "Diagnostics (Trouble)" })
map("n", "<leader>xd", "<cmd>Trouble diagnostics toggle filter.buf=0<CR>", { desc = "Buffer diagnostics" })
map("n", "<leader>xl", "<cmd>Trouble loclist toggle<CR>", { desc = "Location list" })
map("n", "<leader>xq", "<cmd>Trouble qflist toggle<CR>", { desc = "Quickfix list" })

-- Todo comments
map("n", "<leader>xt", "<cmd>TodoTrouble<CR>", { desc = "TODOs (Trouble)" })
map("n", "]t", function() require("todo-comments").jump_next() end, { desc = "Next TODO" })
map("n", "[t", function() require("todo-comments").jump_prev() end, { desc = "Prev TODO" })

-- ── Autocommands ────────────────────────────────────────────
-- Highlight yanked text
vim.api.nvim_create_autocmd("TextYankPost", {
    callback = function()
        vim.highlight.on_yank({ timeout = 150 })
    end,
})

-- Return to last edit position
vim.api.nvim_create_autocmd("BufReadPost", {
    callback = function()
        local mark = vim.api.nvim_buf_get_mark(0, '"')
        if mark[1] > 0 and mark[1] <= vim.api.nvim_buf_line_count(0) then
            vim.api.nvim_win_set_cursor(0, mark)
        end
    end,
})

-- Trim trailing whitespace on save
vim.api.nvim_create_autocmd("BufWritePre", {
    pattern = "*",
    callback = function()
        local pos = vim.api.nvim_win_get_cursor(0)
        vim.cmd([[%s/\s\+$//e]])
        vim.api.nvim_win_set_cursor(0, pos)
    end,
})

-- Auto-resize splits when terminal resizes
vim.api.nvim_create_autocmd("VimResized", {
    callback = function()
        vim.cmd("tabdo wincmd =")
    end,
})
