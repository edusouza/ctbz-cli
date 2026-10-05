#!/usr/bin/env sh
# Verifica se a Contabilizei mudou a API ou o front:
#   1. testes de contrato ao vivo (CTBZ_CONTRACT_LIVE=1), que só fazem GET;
#   2. catálogo de endpoints regenerado a partir dos bundles do painel, comparado com o
#      docs/api/catalogo.md versionado (ignorando a linha com a data de geração);
#   3. escritas usadas pela CLI (goldens de internal/api/testdata/requisicoes) que sumiram
#      do catálogo regenerado.
#
# Requer uma sessão válida (ctbz login, em $CTBZ_HOME) e Python 3.
# Escreve um relatório em Markdown no stdout e sai com 1 se algo mudou, 0 se nada mudou.
# Usado pelo job agendado .github/workflows/monitor.yml; também roda localmente.
set -u

mudou=0
echo "# Monitoramento da API da Contabilizei"
echo

if saida=$(CTBZ_CONTRACT_LIVE=1 go test ./internal/api -run Live -count=1 -v 2>&1); then
	echo "Contratos ao vivo: sem mudanças."
else
	mudou=1
	echo "## Contratos ao vivo"
	echo
	echo "Algum contrato quebrou (campo removido ou tipo mudou). Atualize o tipo em"
	echo "\`internal/api\` e a fixture com \`go run ./tools/capture NOME\`."
	echo
	echo '```text'
	echo "$saida" | grep -E -- '--- FAIL|campo removido|tipo mudou|HTTP [0-9]|Error' | head -100
	echo '```'
fi
echo

if ! python3 scripts/extrair-endpoints.py > docs/api/catalogo.md; then
	mudou=1
	echo "## Catálogo de endpoints"
	echo
	echo "Não foi possível regenerar o catálogo (veja o log do job)."
elif git diff --quiet -I '^> Gerado por' -- docs/api/catalogo.md; then
	echo "Catálogo de endpoints: sem mudanças."
else
	mudou=1
	echo "## Catálogo de endpoints"
	echo
	echo "O front passou a chamar endpoints diferentes. Revise e versione o catálogo novo"
	echo "(\`scripts/extrair-endpoints.py > docs/api/catalogo.md\`)."
	echo
	echo '```diff'
	git diff -I '^> Gerado por' -U0 -- docs/api/catalogo.md | grep -E '^[+-]\|' | head -200
	echo '```'
fi

echo

if saida=$(CTBZ_CATALOGO_ESCRITAS=1 go test ./internal/api -run EscritasNoCatalogo -count=1 -v 2>&1); then
	echo "Escritas da CLI no catálogo: todas presentes."
else
	mudou=1
	echo "## Escritas da CLI"
	echo
	echo "Alguma escrita usada pela CLI não aparece mais no front (caminho ou método mudou)."
	echo "Confira em \`docs/api/escrita/\` e corrija a escrita e o golden antes de usá-la."
	echo
	echo '```text'
	echo "$saida" | grep -E 'ausente do catálogo|--- FAIL|Error' | head -100
	echo '```'
fi

exit "$mudou"
