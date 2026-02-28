" Gatus Monitor - Neovim Project Configuration
" All shortcuts are under <leader>p (project)

" Define project-specific key mappings using which-key
lua << EOF
local wk = require("which-key")

wk.register({
  p = {
    name = "Project",

    -- Build and Run
    b = { "<cmd>!go build -o gatus-monitor ./cmd/gatus-monitor<cr>", "Build project" },
    r = { "<cmd>!./gatus-monitor<cr>", "Run application" },
    R = { "<cmd>!go run ./cmd/gatus-monitor<cr>", "Build and run" },

    -- Testing
    t = { "<cmd>!go test ./...<cr>", "Run all tests" },
    T = { "<cmd>!go test -v -race -coverprofile=coverage.txt ./...<cr>", "Run tests with coverage" },
    c = { "<cmd>!go tool cover -html=coverage.txt<cr>", "View coverage report" },

    -- Documentation
    d = { "<cmd>!mkdocs serve<cr>", "Serve documentation" },
    D = { "<cmd>!mkdocs build<cr>", "Build documentation" },
    o = { "<cmd>!xdg-open http://127.0.0.1:8000<cr>", "Open documentation in browser" },

    -- Code Quality
    f = { "<cmd>!go fmt ./...<cr>", "Format code" },
    l = { "<cmd>!golangci-lint run<cr>", "Run linter" },
    L = { "<cmd>!golangci-lint run --fix<cr>", "Run linter with fixes" },

    -- Git Operations
    g = {
      name = "Git",
      s = { "<cmd>!git status<cr>", "Git status" },
      d = { "<cmd>!git diff<cr>", "Git diff" },
      l = { "<cmd>!git log --oneline -10<cr>", "Git log" },
      a = { "<cmd>!git add .<cr>", "Git add all" },
      c = { "<cmd>!git commit<cr>", "Git commit" },
      p = { "<cmd>!git push<cr>", "Git push" },
    },

    -- Nix operations
    n = {
      name = "Nix",
      d = { "<cmd>!nix develop<cr>", "Enter nix development shell" },
      b = { "<cmd>!nix run .#build-all<cr>", "Build for all platforms" },
      t = { "<cmd>!nix run .#test<cr>", "Run test suite" },
      f = { "<cmd>!nix run .#fmt<cr>", "Format code" },
      l = { "<cmd>!nix run .#lint<cr>", "Run linters" },
    },

    -- Package Management
    m = {
      name = "Modules",
      d = { "<cmd>!go mod download<cr>", "Download dependencies" },
      t = { "<cmd>!go mod tidy<cr>", "Tidy go.mod" },
      v = { "<cmd>!go mod verify<cr>", "Verify dependencies" },
      g = { "<cmd>!go mod graph<cr>", "Show dependency graph" },
    },

    -- Icons
    i = {
      name = "Icons",
      g = { "<cmd>!cd internal/icons && go run generate_icons.go<cr>", "Generate icons" },
    },

    -- Clean
    x = {
      name = "Clean",
      a = { "<cmd>!rm -rf dist/ build/ gatus-monitor<cr>", "Clean all build artifacts" },
      c = { "<cmd>!rm -f coverage.txt coverage.html<cr>", "Clean coverage files" },
      t = { "<cmd>!go clean -testcache<cr>", "Clean test cache" },
    },

    -- Help
    h = { "<cmd>e README.md<cr>", "Open README" },
    s = { "<cmd>e SPECIFICATION.md<cr>", "Open specification" },
    P = { "<cmd>e PACKAGES.md<cr>", "Open packages documentation" },
  },
}, { prefix = "<leader>" })
EOF

" Project-specific settings
set expandtab
set tabstop=4
set shiftwidth=4
set softtabstop=4

" Go-specific settings
autocmd FileType go setlocal noexpandtab tabstop=4 shiftwidth=4

" Automatically format Go code on save
autocmd BufWritePre *.go lua vim.lsp.buf.format()
