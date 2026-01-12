#!/bin/bash

# Se necesita tener instalado Claude CLI y configurado
# Descargar con brew: brew install claude-code para mac

VERSION=$(grep 'version:' lambda.yaml | awk '{print $2}')
DATE=$(date +%Y-%m-%d)
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

# Obtener commits: desde el último tag o todos si no hay tags
if [ -z "$LAST_TAG" ]; then
  COMMITS=$(git log --oneline --no-merges)
else
  COMMITS=$(git log ${LAST_TAG}..HEAD --oneline --no-merges)
fi

if [ -z "$COMMITS" ]; then
  echo "No hay commits nuevos"
  exit 0
fi

PROMPT="Genera una entrada de CHANGELOG para la versión $VERSION con fecha $DATE.

Commits:
$COMMITS

REGLAS:
1. Solo crea secciones para tipos que EXISTAN en los commits:
   - feat = ### Added
   - fix = ### Fixed
   - perf = ### Performance
   - refactor, chore, test, docs = IGNORAR
2. NO inventes información
3. NO incluyas prefijos (feat/fix) ni scopes en el texto final
4. Descripciones limpias y claras en español

Formato:
## [$VERSION] - $DATE

### Added
- descripción

Responde SOLO con el markdown, sin explicaciones ni bloques de código."

# Generar nuevo entry con Claude
NEW_ENTRY=$(NODE_TLS_REJECT_UNAUTHORIZED=0 claude -p "$PROMPT" 2>/dev/null)

# Verificar si CHANGELOG.md existe
if [ -f CHANGELOG.md ]; then
  if grep -q "## \[$VERSION\]" CHANGELOG.md; then
    echo "⚠️  La versión $VERSION ya existe en CHANGELOG.md"
    exit 1
  fi
  
  if head -1 CHANGELOG.md | grep -q "^# "; then
    head -1 CHANGELOG.md > CHANGELOG.tmp
    echo "" >> CHANGELOG.tmp
    echo "$NEW_ENTRY" >> CHANGELOG.tmp
    tail -n +2 CHANGELOG.md >> CHANGELOG.tmp
  else
    echo "$NEW_ENTRY" > CHANGELOG.tmp
    echo "" >> CHANGELOG.tmp
    cat CHANGELOG.md >> CHANGELOG.tmp
  fi
  mv CHANGELOG.tmp CHANGELOG.md
else
  echo "# Changelog" > CHANGELOG.md
  echo "" >> CHANGELOG.md
  echo "$NEW_ENTRY" >> CHANGELOG.md
fi

echo "✅ CHANGELOG.md actualizado para v$VERSION"
