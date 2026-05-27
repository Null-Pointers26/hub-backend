#!/usr/bin/env bash
# install-hub.sh — Nainstaluje a spustí Hub (backend + frontend).
#
# Použití:
#   bash install-hub.sh [--backend-repo URL] [--frontend-repo URL]
#
# Předpoklady:
#   - Docker Engine >= 24 a Docker Compose Plugin ("docker compose")
#   - Git

set -euo pipefail

# ── Konfigurovatelné hodnoty ──────────────────────────────────────────────────
BACKEND_REPO="${BACKEND_REPO:-https://github.com/Null-Pointers26/hub-backend.git}"
FRONTEND_REPO="${FRONTEND_REPO:-https://github.com/Null-Pointers26/hub-frontend.git}"
INSTALL_DIR="${INSTALL_DIR:-hub}"
NETWORK_NAME="hub-shared"

# ── Barvy ─────────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

info()  { echo -e "${CYAN}[hub-install]${NC} $*"; }
ok()    { echo -e "${GREEN}[hub-install] ✔${NC} $*"; }
warn()  { echo -e "${YELLOW}[hub-install] ⚠${NC} $*"; }
err()   { echo -e "${RED}[hub-install] ✘${NC} $*" >&2; exit 1; }

# ── Parsování argumentů ───────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --backend-repo)  BACKEND_REPO="$2";  shift 2 ;;
    --frontend-repo) FRONTEND_REPO="$2"; shift 2 ;;
    --dir)           INSTALL_DIR="$2";   shift 2 ;;
    *) err "Neznámý argument: $1" ;;
  esac
done

# ── Kontrola závislostí ───────────────────────────────────────────────────────
info "Kontrola závislostí..."
command -v docker &>/dev/null || err "Docker není nainstalován."
docker compose version &>/dev/null || err "Docker Compose Plugin není dostupný. Nainstalujte 'docker-compose-plugin'."
command -v git    &>/dev/null || err "Git není nainstalován."
ok "Závislosti jsou splněny."

# ── Vytvoření instalačního adresáře ──────────────────────────────────────────
info "Instalace do adresáře: $(pwd)/$INSTALL_DIR"
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

# ── Klonování repozitářů ──────────────────────────────────────────────────────
clone_or_pull() {
  local repo="$1" dest="$2"
  if [[ -d "$dest/.git" ]]; then
    info "Repozitář '$dest' již existuje — provádím git pull."
    git -C "$dest" pull --ff-only
  else
    info "Klonuji $repo → $dest ..."
    git clone "$repo" "$dest"
  fi
}

clone_or_pull "$BACKEND_REPO"  "hub-backend"
clone_or_pull "$FRONTEND_REPO" "hub-frontend"
ok "Repozitáře jsou připraveny."

# ── Docker síť ───────────────────────────────────────────────────────────────
if docker network inspect "$NETWORK_NAME" &>/dev/null; then
  ok "Síť '$NETWORK_NAME' již existuje."
else
  info "Vytvářím Docker síť '$NETWORK_NAME'..."
  docker network create "$NETWORK_NAME"
  ok "Síť '$NETWORK_NAME' vytvořena."
fi

# ── Konfigurace prostředí ─────────────────────────────────────────────────────
if [[ ! -f hub-backend/.env ]]; then
  if [[ -f hub-backend/.env.example ]]; then
    cp hub-backend/.env.example hub-backend/.env
    warn "Soubor hub-backend/.env byl vytvořen z .env.example."
    warn "Upravte hub-backend/.env (zejména DOMAIN a CERT_EMAIL) před ostrým nasazením."
  else
    warn "Soubor .env.example nenalezen. Vytvořte hub-backend/.env ručně."
  fi
else
  ok "Soubor hub-backend/.env již existuje, přeskočeno."
fi

# ── Spuštění systému ──────────────────────────────────────────────────────────
info "Sestavuji a spouštím Hub (docker compose up --build -d)..."
docker compose -f hub-backend/docker-compose.yml up --build -d

ok "Hub byl úspěšně spuštěn!"
echo ""
echo -e "  Frontend + Backend:  ${CYAN}http://localhost:80${NC}"
echo -e "  API seznam her:      ${CYAN}http://localhost:80/api/games${NC}"
echo ""
echo -e "  Pro logy spusťte:    ${YELLOW}docker compose -f hub-backend/docker-compose.yml logs -f${NC}"
echo ""
