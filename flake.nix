{
  description = "soyweb - ssg wrapper with additional functionality";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
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
          vendorHash = "sha256-1cj4L2eGRXFuU4LNPBL4/9Ucw3kkWjUNDJAOtkJNlwQ=";
          meta = {
            homepage = "https://github.com/soyart/soyweb";
            description = "ssg-go wrapper";
          };
        };
      });

      devShells = forAllSystems ({ pkgs }: {
        default = pkgs.mkShell {
          shellHook = ''
            echo "Entering Nix shell";
            echo "Go version:";
            go version
          '';
          packages = with pkgs; [
            nixd
            nixpkgs-fmt

            bash-language-server
            shellcheck
            shfmt

            coreutils
            lowdown

            go
            gopls
            gotools
            go-tools
          ];
        };
      });
    };
}
