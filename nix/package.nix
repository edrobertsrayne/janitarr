{
  lib,
  buildGoModule,
  templ,
  fetchFromGitHub,
}:
buildGoModule rec {
  pname = "janitarr";
  version = "0.1.0";

  src = ./..;

  vendorHash = "sha256-ngeZvarE1m7PaVeEF3UTLHgS1IHObcAgdNGI1iG2yAQ=";

  nativeBuildInputs = [templ];

  preBuild = ''
    templ generate
  '';

  ldflags = [
    "-s"
    "-w"
  ];

  meta = with lib; {
    description = "Automation tool for Radarr and Sonarr media servers";
    homepage = "https://github.com/edrobertsrayne/janitarr";
    license = licenses.mit;
    maintainers = [];
    mainProgram = "janitarr";
  };
}
