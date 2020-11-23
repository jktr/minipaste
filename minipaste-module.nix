{ config, pkgs, lib, ... }:

with lib;

let
  cfg = config.services.minipaste;
in {
  options.services.minipaste = {
    enable = mkEnableOption "Minipaste daemon";
    package = mkOption {
      type = types.package;
      default = pkgs.minipaste;
      defaultText = "pkgs.minipaste";
      description = "Use this package instead of the default one";
    };
    bind = mkOption {
      type = types.str;
      default = "[::1]:8080";
      description = "Bind to this IP and port";
    };
    retention = mkOption {
      type = types.str;
      default = "5m";
      description = "Retention time for uploaded pastes, zero is infinite";
    };
    index = mkOption {
      type = types.nullOr types.path;
      default = null;
      description = "Serve this file as the front page";
    };
    limit = mkOption {
      type = types.int;
      default = 16;
      description = "Upload size limit in MB";
    };
  };
  config = mkIf cfg.enable {
    systemd.services.minipaste = {
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];
      serviceConfig = {
        ExecStart = "${cfg.package}/bin/minipaste-server "
          + "-bind ${cfg.bind} -size-limit ${toString cfg.limit} "
          + "-retention ${cfg.retention} "
          + "${optionalString (cfg.index != null) "-index ${cfg.index}"}";
        DynamicUser = true;
      };
    };
  };
}
