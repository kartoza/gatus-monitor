-- Gatus Monitor - Neovim Project-Specific Configuration
-- This file is automatically loaded when opening this project in Neovim

-- Project root directory
local project_root = vim.fn.getcwd()

-- LSP Configuration for Go
-- Use vim.lsp.config instead of deprecated lspconfig
vim.lsp.config.gopls = {
  cmd = {'gopls'},
  filetypes = {'go', 'gomod', 'gowork', 'gotmpl'},
  root_markers = {'go.mod', '.git'},
  settings = {
    gopls = {
      analyses = {
        unusedparams = true,
        shadow = true,
        nilness = true,
        unusedwrite = true,
      },
      staticcheck = true,
      gofumpt = true,
      usePlaceholders = true,
      completeUnimported = true,
    },
  },
}

-- Enable gopls for Go files
vim.api.nvim_create_autocmd('FileType', {
  pattern = 'go',
  callback = function()
    vim.lsp.enable('gopls')
  end,
})

-- LSP keymaps (set globally)
vim.api.nvim_create_autocmd('LspAttach', {
  callback = function(args)
    local bufnr = args.buf
    local bufopts = { noremap=true, silent=true, buffer=bufnr }
    vim.keymap.set('n', 'gD', vim.lsp.buf.declaration, bufopts)
    vim.keymap.set('n', 'gd', vim.lsp.buf.definition, bufopts)
    vim.keymap.set('n', 'K', vim.lsp.buf.hover, bufopts)
    vim.keymap.set('n', 'gi', vim.lsp.buf.implementation, bufopts)
    vim.keymap.set('n', '<C-k>', vim.lsp.buf.signature_help, bufopts)
    vim.keymap.set('n', '<leader>rn', vim.lsp.buf.rename, bufopts)
    vim.keymap.set('n', '<leader>ca', vim.lsp.buf.code_action, bufopts)
    vim.keymap.set('n', 'gr', vim.lsp.buf.references, bufopts)
    vim.keymap.set('n', '<leader>f', function() vim.lsp.buf.format { async = true } end, bufopts)
  end,
})

-- Old lspconfig setup (fallback for older neovim versions)
local ok, lspconfig = pcall(require, 'lspconfig')
if ok then
  lspconfig.gopls.setup{
  cmd = {'gopls'},
  settings = {
    gopls = {
      analyses = {
        unusedparams = true,
        shadow = true,
        nilness = true,
        unusedwrite = true,
      },
      staticcheck = true,
      gofumpt = true,
      usePlaceholders = true,
      completeUnimported = true,
    },
  },
  on_attach = function(client, bufnr)
    -- Enable completion triggered by <c-x><c-o>
    vim.api.nvim_buf_set_option(bufnr, 'omnifunc', 'v:lua.vim.lsp.omnifunc')
  end,
  }
end

-- Auto-format on save for Go files
vim.api.nvim_create_autocmd("BufWritePre", {
  pattern = "*.go",
  callback = function()
    vim.lsp.buf.format({ async = false })
  end,
})

-- Organize imports on save
vim.api.nvim_create_autocmd("BufWritePre", {
  pattern = "*.go",
  callback = function()
    local params = vim.lsp.util.make_range_params()
    params.context = {only = {"source.organizeImports"}}
    local result = vim.lsp.buf_request_sync(0, "textDocument/codeAction", params, 3000)
    for _, res in pairs(result or {}) do
      for _, r in pairs(res.result or {}) do
        if r.edit then
          vim.lsp.util.apply_workspace_edit(r.edit, "utf-8")
        else
          vim.lsp.buf.execute_command(r.command)
        end
      end
    end
  end,
})

-- Diagnostics configuration
vim.diagnostic.config({
  virtual_text = true,
  signs = true,
  underline = true,
  update_in_insert = false,
  severity_sort = true,
})

-- Diagnostic signs
local signs = { Error = " ", Warn = " ", Hint = " ", Info = " " }
for type, icon in pairs(signs) do
  local hl = "DiagnosticSign" .. type
  vim.fn.sign_define(hl, { text = icon, texthl = hl, numhl = hl })
end

-- Treesitter configuration for Go (optional)
local ok_treesitter, treesitter = pcall(require, 'nvim-treesitter.configs')
if ok_treesitter then
  treesitter.setup {
    ensure_installed = { "go", "gomod", "gowork", "gosum" },
    highlight = {
      enable = true,
      additional_vim_regex_highlighting = false,
    },
    indent = {
      enable = true,
    },
  }
end

-- DAP (Debug Adapter Protocol) configuration for Go (optional)
local ok_dap, dap = pcall(require, 'dap')
if ok_dap then
  dap.adapters.go = {
    type = 'executable',
    command = 'dlv',
    args = {'dap'},
  }

  dap.configurations.go = {
    {
      type = 'go',
      name = 'Debug',
      request = 'launch',
      program = "${file}",
    },
    {
      type = 'go',
      name = 'Debug Package',
      request = 'launch',
      program = "${fileDirname}",
    },
    {
      type = 'go',
      name = 'Debug test',
      request = 'launch',
      mode = 'test',
      program = "${file}",
    },
    {
      type = 'go',
      name = 'Attach to running process',
      mode = 'local',
      request = 'attach',
      processId = require('dap.utils').pick_process,
    },
  }

  -- Additional keymaps for debugging
  vim.keymap.set('n', '<F5>', dap.continue, { desc = 'Debug: Continue' })
  vim.keymap.set('n', '<F10>', dap.step_over, { desc = 'Debug: Step Over' })
  vim.keymap.set('n', '<F11>', dap.step_into, { desc = 'Debug: Step Into' })
  vim.keymap.set('n', '<F12>', dap.step_out, { desc = 'Debug: Step Out' })
  vim.keymap.set('n', '<leader>b', dap.toggle_breakpoint, { desc = 'Debug: Toggle Breakpoint' })
end

-- Test runner configuration
vim.keymap.set('n', '<leader>tn', ':TestNearest<CR>', { desc = 'Test: Run nearest test' })
vim.keymap.set('n', '<leader>tf', ':TestFile<CR>', { desc = 'Test: Run current file tests' })
vim.keymap.set('n', '<leader>ts', ':TestSuite<CR>', { desc = 'Test: Run test suite' })
vim.keymap.set('n', '<leader>tl', ':TestLast<CR>', { desc = 'Test: Run last test' })
vim.keymap.set('n', '<leader>tv', ':TestVisit<CR>', { desc = 'Test: Visit last test file' })

-- File explorer (nvim-tree) project-specific settings (optional)
local ok_tree, nvim_tree = pcall(require, 'nvim-tree')
if ok_tree then
  nvim_tree.setup({
    update_cwd = true,
    update_focused_file = {
      enable = true,
      update_cwd = true,
    },
    view = {
      width = 35,
    },
    filters = {
      custom = { ".git", "node_modules", ".cache", "dist", "build", "*.exe" },
    },
  })
end

-- Telescope project-specific settings for Go (optional)
local ok_telescope, telescope = pcall(require, 'telescope')
if ok_telescope then
  telescope.setup{
    defaults = {
      file_ignore_patterns = {
        "vendor/",
        "%.pb.go",
        "node_modules/",
        ".git/",
      },
    },
  }
end

-- Project-specific snippets for common patterns (optional)
local ok_luasnip, ls = pcall(require, "luasnip")
if ok_luasnip then
  local s = ls.snippet
  local t = ls.text_node
  local i = ls.insert_node

  ls.add_snippets("go", {
    -- Test function
    s("test", {
      t("func Test"), i(1, "Name"), t("(t *testing.T) {"), t({"", "\t"}),
      i(0),
      t({"", "}"})
    }),

    -- Benchmark function
    s("bench", {
      t("func Benchmark"), i(1, "Name"), t("(b *testing.B) {"), t({"", "\t"}),
      t("for i := 0; i < b.N; i++ {"), t({"", "\t\t"}),
      i(0),
      t({"", "\t}"}),
      t({"", "}"})
    }),

    -- Error check
    s("iferr", {
      t("if err != nil {"), t({"", "\t"}),
      t("return "), i(1, "err"),
      t({"", "}"})
    }),
  })
end

-- Status line customization
vim.g.project_name = "Gatus Monitor"

-- Print loaded features
print("=== Gatus Monitor Project Configuration ===")
print("✓ LSP: gopls configured")
if ok_treesitter then print("✓ Treesitter: enabled for Go") end
if ok_dap then print("✓ DAP: delve debugger configured") end
if ok_tree then print("✓ nvim-tree: file explorer configured") end
if ok_telescope then print("✓ Telescope: fuzzy finder configured") end
if ok_luasnip then print("✓ LuaSnip: Go snippets loaded") end
print("Use <leader>p for project-specific commands")
print("===========================================")
