{
  description = "CSC 5610 Labs and Resources";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }: 
  let 
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};
    python-pkgs = pkgs.python312Packages;
  in {
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = with pkgs; [
          python3
          sqlite
          doctl
          # "${python-packages}.pandas"
      ];

      shellHook = ''
          exec zsh
      '';
    };
  };
}
