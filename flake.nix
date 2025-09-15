{
  description = "soyweb - ssg wrapper with additional functionality";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    ssg-go = {
      url = "github:soyart/ssg-go";
      flake = false;
    };
    ssg-testdata = {
      url = "github:soyart/ssg-testdata";
      flake = false;
    };
  };

  outputs = { self, nixpkgs, ssg-go, ssg-testdata }:
    let
      homepage = "https://github.com/soyart/soyweb";

      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
      version = builtins.substring 0 8 lastModifiedDate;

      # The set of systems to provide outputs for
      allSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      # A function that provides a system-specific Nixpkgs for the desired systems
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        pkgs = import nixpkgs { inherit system; };
      });
    in

    {
      packages = forAllSystems ({ pkgs }: {
        default = pkgs.buildGoModule {
          inherit version;
          pname = "soyweb";
          
          src = ./.;
          
          # Go module configuration
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # This will be updated by nix
          
          # Build configuration
          buildPhase = ''
            runHook preBuild
            
            # Ensure test data is available during build
            ln -sf ${ssg-testdata} ./testdata
            
            # Build all binaries
            go build -o $out/bin/soyweb ./cmd/soyweb
            go build -o $out/bin/ssg-minifier ./cmd/ssg-minifier
            go build -o $out/bin/minifier ./cmd/minifier
            
            runHook postBuild
          '';
          
          # Install configuration
          installPhase = ''
            runHook preInstall
            
            mkdir -p $out/bin
            cp soyweb $out/bin/soyweb
            cp ssg-minifier $out/bin/ssg-minifier
            cp minifier $out/bin/minifier
            
            runHook postInstall
          '';
          
          # Test configuration
          checkPhase = ''
            runHook preCheck
            
            # Ensure test data is available for tests
            ln -sf ${ssg-testdata} ./testdata
            
            # Run tests
            go test -v ./...
            
            runHook postCheck
          '';
          
          meta = {
            inherit homepage;
            description = "soyweb - ssg wrapper with additional functionality";
            license = pkgs.lib.licenses.mit;
            maintainers = [ ];
            platforms = pkgs.lib.platforms.unix;
          };
        };

        # Individual binary packages
        soyweb = pkgs.buildGoModule {
          inherit version;
          pname = "soyweb";
          
          src = ./.;
          
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          
          buildPhase = ''
            runHook preBuild
            ln -sf ${ssg-testdata} ./testdata
            go build -o $out/bin/soyweb ./cmd/soyweb
            runHook postBuild
          '';
          
          installPhase = ''
            runHook preInstall
            mkdir -p $out/bin
            cp soyweb $out/bin/soyweb
            runHook postInstall
          '';
          
          checkPhase = ''
            runHook preCheck
            ln -sf ${ssg-testdata} ./testdata
            go test -v ./...
            runHook postCheck
          '';
          
          meta = {
            inherit homepage;
            description = "soyweb - main ssg wrapper binary";
            license = pkgs.lib.licenses.mit;
            maintainers = [ ];
            platforms = pkgs.lib.platforms.unix;
          };
        };

        ssg-minifier = pkgs.buildGoModule {
          inherit version;
          pname = "ssg-minifier";
          
          src = ./.;
          
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          
          buildPhase = ''
            runHook preBuild
            ln -sf ${ssg-testdata} ./testdata
            go build -o $out/bin/ssg-minifier ./cmd/ssg-minifier
            runHook postBuild
          '';
          
          installPhase = ''
            runHook preInstall
            mkdir -p $out/bin
            cp ssg-minifier $out/bin/ssg-minifier
            runHook postInstall
          '';
          
          checkPhase = ''
            runHook preCheck
            ln -sf ${ssg-testdata} ./testdata
            go test -v ./...
            runHook postCheck
          '';
          
          meta = {
            inherit homepage;
            description = "ssg-minifier - minification tool";
            license = pkgs.lib.licenses.mit;
            maintainers = [ ];
            platforms = pkgs.lib.platforms.unix;
          };
        };

        minifier = pkgs.buildGoModule {
          inherit version;
          pname = "minifier";
          
          src = ./.;
          
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          
          buildPhase = ''
            runHook preBuild
            ln -sf ${ssg-testdata} ./testdata
            go build -o $out/bin/minifier ./cmd/minifier
            runHook postBuild
          '';
          
          installPhase = ''
            runHook preInstall
            mkdir -p $out/bin
            cp minifier $out/bin/minifier
            runHook postInstall
          '';
          
          checkPhase = ''
            runHook preCheck
            ln -sf ${ssg-testdata} ./testdata
            go test -v ./...
            runHook postCheck
          '';
          
          meta = {
            inherit homepage;
            description = "minifier - standalone minification tool";
            license = pkgs.lib.licenses.mit;
            maintainers = [ ];
            platforms = pkgs.lib.platforms.unix;
          };
        };
      });

      devShells = forAllSystems ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            # Nix development tools
            nixd
            nixpkgs-fmt
            
            # Go development tools
            go
            gopls
            gotools
            go-tools
            
            # Testing tools
            ginkgo
            gomega
          ];
          
          # Environment variables
          shellHook = ''
            echo "soyweb development environment"
            echo "Go version: $(go version)"
            echo "Test data will be available at ./testdata (submodule)"
            
            # Check if testdata submodule exists
            if [ ! -d "./testdata" ]; then
              echo "Warning: testdata submodule not found. Run:"
              echo "  git submodule add https://github.com/soyart/ssg-testdata.git testdata"
            fi
          '';
        };
      });

      # Nix flake checks
      checks = forAllSystems ({ pkgs }: {
        # Build check
        build = self.packages.${pkgs.system}.default;
        
        # Test check
        test = pkgs.runCommand "soyweb-tests" {
          nativeBuildInputs = with pkgs; [ go ];
        } ''
          cp -r ${self}/* .
          chmod -R +w .
          
          # Ensure test data is available
          ln -sf ${ssg-testdata} ./testdata
          
          # Run tests
          go test -v ./...
          
          touch $out
        '';
        
        # Lint check
        lint = pkgs.runCommand "soyweb-lint" {
          nativeBuildInputs = with pkgs; [ go-tools ];
        } ''
          cp -r ${self}/* .
          chmod -R +w .
          
          # Run linter
          golangci-lint run
          
          touch $out
        '';
      });
    };
}
