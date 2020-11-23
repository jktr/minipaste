{
  description = "A tiny in-memory pastebin server";

  outputs = { self, nixpkgs }: {

    overlay = final: prev: {
      minipaste = with final; buildGoModule {
        name = "minipaste";
        src = self;
        vendorSha256 = "sha256-xfGAHxXOUcNLGYmsicVDa5I3wlyW8aZ2eDILYobyBgc=";
      };
    };

    packages.x86_64-linux.minipaste = (import nixpkgs {
      system = "x86_64-linux";
      overlays = [ self.overlay ];
    }).minipaste;

    apps.x86_64-linux = with self.packages.x86_64-linux; {
      client = {
        type = "app"; 
        program = "${minipaste}/bin/minipaste"; 
      };
      server = {
        type = "app"; 
        program = "${minipaste}/bin/minipaste-server"; 
      };
    };

    defaultApp.x86_64-linux =  self.apps.x86_64-linux.client;
    defaultPackage.x86_64-linux = self.packages.x86_64-linux.minipaste;

    nixosModule = self.nixosModules.minipaste;
    nixosModules.minipaste = { ... }: {
      nixpkgs.overlays = [ self.overlay ];
      imports = [ ./minipaste-module.nix ];
    };
  };
}
