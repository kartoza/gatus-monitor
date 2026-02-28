{
  description = "Gatus Monitor - Cross-platform system tray app for monitoring Gatus endpoints";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        # Version from VERSION file or default
        version = if builtins.pathExists ./VERSION
                  then builtins.readFile ./VERSION
                  else "0.1.0";

        # Build the application
        gatus-monitor = pkgs.buildGoModule {
          pname = "gatus-monitor";
          version = version;
          src = ./.;

          vendorHash = null; # Will be updated after first build

          nativeBuildInputs = with pkgs; [
            pkg-config
          ];

          buildInputs = with pkgs; [
            # GUI libraries for Fyne
            xorg.libX11
            xorg.libXcursor
            xorg.libXrandr
            xorg.libXinerama
            xorg.libXi
            xorg.libXxf86vm
            libGL

            # System tray support
            gtk3
            libappindicator-gtk3
          ] ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
            # Linux-specific
            xorg.libXext
          ] ++ pkgs.lib.optionals pkgs.stdenv.isDarwin [
            # macOS-specific
            darwin.apple_sdk.frameworks.Cocoa
            darwin.apple_sdk.frameworks.IOKit
            darwin.apple_sdk.frameworks.Kernel
          ];

          meta = with pkgs.lib; {
            description = "Cross-platform system tray app for monitoring Gatus endpoints";
            homepage = "https://github.com/kartoza/gatus-monitor";
            license = licenses.mit;
            maintainers = [];
            platforms = platforms.linux ++ platforms.darwin ++ platforms.windows;
          };
        };

      in
      {
        packages = {
          default = gatus-monitor;
          gatus-monitor = gatus-monitor;
        };

        # Development shell
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go toolchain
            go_1_21
            gopls
            gotools
            go-tools

            # Development tools
            golangci-lint
            git
            gh

            # Pre-commit framework
            pre-commit

            # Documentation
            python311
            python311Packages.mkdocs
            python311Packages.mkdocs-material
            python311Packages.mkdocs-mermaid2-plugin

            # Build tools
            pkg-config

            # GUI libraries (same as buildInputs)
            xorg.libX11
            xorg.libXcursor
            xorg.libXrandr
            xorg.libXinerama
            xorg.libXi
            xorg.libXxf86vm
            libGL
            gtk3
            libappindicator-gtk3
          ] ++ lib.optionals stdenv.isLinux [
            xorg.libXext

            # Packaging tools for Linux
            dpkg
            rpm
            fpm
            appimagekit
          ] ++ lib.optionals stdenv.isDarwin [
            darwin.apple_sdk.frameworks.Cocoa
            darwin.apple_sdk.frameworks.IOKit
            darwin.apple_sdk.frameworks.Kernel
          ];

          shellHook = ''
            echo "Gatus Monitor Development Environment"
            echo "======================================"
            echo ""
            echo "Go version: $(go version)"
            echo "Project: github.com/kartoza/gatus-monitor"
            echo ""
            echo "Available commands:"
            echo "  go run ./cmd/gatus-monitor  - Run the application"
            echo "  go test ./...               - Run tests"
            echo "  golangci-lint run           - Run linters"
            echo "  nix run .#build-all         - Build for all platforms"
            echo "  nix run .#docs              - Build and serve documentation"
            echo "  nix run .#test              - Run full test suite"
            echo ""
            echo "Neovim shortcuts available via <leader>p (see .exrc)"
            echo ""

            # Set up Go environment
            export GOPATH="$HOME/go"
            export PATH="$GOPATH/bin:$PATH"

            # Set up pre-commit hooks if not already installed
            if [ -f .pre-commit-config.yaml ] && [ ! -f .git/hooks/pre-commit ]; then
              echo "Installing pre-commit hooks..."
              pre-commit install
            fi
          '';
        };

        # Convenience apps
        apps = {
          default = {
            type = "app";
            program = "${gatus-monitor}/bin/gatus-monitor";
          };

          # Build for all platforms
          build-all = {
            type = "app";
            program = toString (pkgs.writeShellScript "build-all" ''
              set -e
              echo "Building for all platforms..."

              # Linux amd64
              GOOS=linux GOARCH=amd64 go build -o dist/gatus-monitor_linux_amd64 ./cmd/gatus-monitor

              # Linux arm64
              GOOS=linux GOARCH=arm64 go build -o dist/gatus-monitor_linux_arm64 ./cmd/gatus-monitor

              # macOS amd64
              GOOS=darwin GOARCH=amd64 go build -o dist/gatus-monitor_darwin_amd64 ./cmd/gatus-monitor

              # macOS arm64 (Apple Silicon)
              GOOS=darwin GOARCH=arm64 go build -o dist/gatus-monitor_darwin_arm64 ./cmd/gatus-monitor

              # Windows amd64
              GOOS=windows GOARCH=amd64 go build -o dist/gatus-monitor_windows_amd64.exe ./cmd/gatus-monitor

              echo "Build complete! Binaries in dist/"
              ls -lh dist/
            '');
          };

          # Run documentation server
          docs = {
            type = "app";
            program = toString (pkgs.writeShellScript "docs" ''
              set -e
              if [ ! -d "docs" ]; then
                echo "Error: docs/ directory not found"
                exit 1
              fi
              cd docs
              echo "Building and serving documentation..."
              echo "Navigate to http://127.0.0.1:8000"
              mkdocs serve
            '');
          };

          # Run full test suite
          test = {
            type = "app";
            program = toString (pkgs.writeShellScript "test" ''
              set -e
              echo "Running test suite..."
              go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
              echo ""
              echo "Coverage report:"
              go tool cover -func=coverage.txt
            '');
          };

          # Format code
          fmt = {
            type = "app";
            program = toString (pkgs.writeShellScript "fmt" ''
              set -e
              echo "Formatting Go code..."
              go fmt ./...
              echo "Done!"
            '');
          };

          # Run linters
          lint = {
            type = "app";
            program = toString (pkgs.writeShellScript "lint" ''
              set -e
              echo "Running linters..."
              golangci-lint run ./...
            '');
          };
        };
      }
    );
}
