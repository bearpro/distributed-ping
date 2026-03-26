{
  description = "distributed-ping development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: 
    flake-utils.lib.eachDefaultSystem (system: 
      let pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            delve
            gotools
            golangci-lint
            air
            
            elmPackages.elm
            elmPackages.elm-language-server
            elmPackages.elm-format
            elmPackages.elm-test
            elmPackages.elm-live
          ];

          shellHook = ''
            export GOPATH="$PWD/.gopath"
            export GOBIN="$GOPATH/bin"
            export PATH="$GOBIN:$PATH"

            export CGO_ENABLED=1

            echo "Go dev shell"
            go version
          '';
        };
      });
}

