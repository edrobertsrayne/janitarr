{
  config,
  lib,
  pkgs,
  ...
}:
with lib; let
  cfg = config.services.janitarr;
in {
  options.services.janitarr = {
    enable = mkEnableOption "Janitarr media server automation";

    port = mkOption {
      type = types.port;
      default = 3434;
      description = "Port for the web interface.";
    };

    openFirewall = mkOption {
      type = types.bool;
      default = false;
      description = "Whether to open the firewall for the web interface.";
    };

    dataDir = mkOption {
      type = types.path;
      default = "/var/lib/janitarr";
      description = "Directory for database and application data.";
    };

    logLevel = mkOption {
      type = types.enum ["debug" "info" "warn" "error"];
      default = "info";
      description = "Log verbosity level.";
    };

    user = mkOption {
      type = types.str;
      default = "janitarr";
      description = "User account under which janitarr runs.";
    };

    group = mkOption {
      type = types.str;
      default = "janitarr";
      description = "Group under which janitarr runs.";
    };

    package = mkOption {
      type = types.package;
      default = pkgs.janitarr or (pkgs.callPackage ./package.nix {});
      description = "The janitarr package to use.";
    };
  };

  config = mkIf cfg.enable {
    users.users.janitarr = mkIf (cfg.user == "janitarr") {
      isSystemUser = true;
      group = cfg.group;
      home = cfg.dataDir;
      description = "Janitarr service user";
    };

    users.groups.janitarr = mkIf (cfg.group == "janitarr") {};

    networking.firewall.allowedTCPPorts = mkIf cfg.openFirewall [cfg.port];

    systemd.services.janitarr = {
      description = "Janitarr media server automation";
      wantedBy = ["multi-user.target"];
      after = ["network.target"];

      serviceConfig = {
        Type = "simple";
        User = cfg.user;
        Group = cfg.group;
        ExecStart = "${cfg.package}/bin/janitarr start --host 0.0.0.0 --port ${toString cfg.port}";
        Restart = "on-failure";
        RestartSec = 5;

        # State management
        StateDirectory = "janitarr";
        StateDirectoryMode = "0750";
        WorkingDirectory = cfg.dataDir;

        # Security hardening
        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = ["AF_INET" "AF_INET6" "AF_UNIX"];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        MemoryDenyWriteExecute = true;
        LockPersonality = true;
        SystemCallFilter = ["@system-service" "~@privileged" "~@resources"];
        SystemCallArchitectures = "native";
        CapabilityBoundingSet = "";
        AmbientCapabilities = "";
      };

      environment = {
        JANITARR_DB_PATH = "${cfg.dataDir}/janitarr.db";
        JANITARR_LOG_LEVEL = cfg.logLevel;
      };
    };
  };
}
