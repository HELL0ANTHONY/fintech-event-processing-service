#!/bin/bash

# Este script genera una entrada de CHANGELOG usando un modelo de lenguaje.
# Requiere que 'ollama' esté instalado y configurado con el modelo 'llama3.2:1b'.
# El script asume que la versión actual está en 'lambda.yaml' y obtiene los commits desde la última etiqueta git.
# El resultado se imprime en formato markdown.
# Uso: ./scripts/generate-changelog.sh
# Asegúrate de tener permisos de ejecución: chmod +x scripts/generate-changelog.sh
VERSION=$(grep 'version:' lambda.yaml | awk '{print $2}')
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

if [ -z "$LAST_TAG" ]; then
  COMMITS=$(git log --oneline --no-merges -20)
else
  COMMITS=$(git log ${LAST_TAG}..HEAD --oneline --no-merges)
fi

if [ -z "$COMMITS" ]; then
  echo "No hay commits nuevos"
  exit 0
fi

PROMPT="Genera una entrada de CHANGELOG en español para la versión $VERSION.
Fecha: $(date +%Y-%m-%d)

Commits:
$COMMITS

Usa este formato exacto:
## [$VERSION] - $(date +%Y-%m-%d)
### Added
- descripción sin el prefijo feat/fix

### Fixed
- descripción

Solo incluye secciones que tengan contenido. Responde SOLO con el markdown, sin explicaciones."

echo "$PROMPT" | ollama run llama3.2:1b
