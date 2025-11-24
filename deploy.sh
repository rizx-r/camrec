#!/usr/bin/env bash
set -euo pipefail

compose() {
  if command -v docker-compose >/dev/null 2>&1; then
    docker-compose "$@"
  else
    docker compose "$@"
  fi
}

require_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "未检测到 docker 命令" >&2
    exit 1
  fi
}

deploy_all() {
  require_docker
  compose up -d --build
}

deploy_camrec() {
  require_docker
  compose up -d --build camrec
}

menu() {
  echo "请选择操作："
  echo "1) 一键部署"
  echo "2) 仅部署 camrec"
  echo "q) 退出"
  printf "> "
  read -r choice
  case "$choice" in
    1)
      deploy_all
      ;;
    2)
      deploy_camrec
      ;;
    q|Q)
      exit 0
      ;;
    *)
      echo "无效选项"
      ;;
  esac
}

menu
