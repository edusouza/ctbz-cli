#!/usr/bin/env bash
# Coleta, só com leituras e em paralelo, tudo o que o checkup da empresa precisa.
#
# Uso: coletar.sh [AAAA-MM]     (padrão: o mês atual)
#
# Imprime uma seção por comando: o JSON da saída, o stderr (totais e avisos) e o código de
# saída. Termina com 3 se não houver sessão válida (faça o login antes) e com 0 nos demais
# casos, mesmo que algum comando falhe: a falha aparece na seção dele.
# CTBZ_BIN escolhe o binário (padrão: ctbz).

set -u

CTBZ=${CTBZ_BIN:-ctbz}
MES=${1:-$(date +%Y-%m)}
ANO=${MES%%-*}
PARALELO=4

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

if ! "$CTBZ" status -o json >"$TMP/status.out" 2>"$TMP/status.err" ||
	grep -q '"situacao": *"expirada"' "$TMP/status.out"; then
	cat "$TMP/status.out" "$TMP/status.err"
	echo "Sem sessão válida: faça o login (ctbz login) e rode de novo." >&2
	exit 3
fi

# Um comando por linha; os argumentos não têm espaços.
COMANDOS=(
	"resumo"
	"impostos"
	"impostos debitos"
	"impostos faturamento"
	"pendencias"
	"pendencias termos"
	"pendencias procuracao"
	"pendencias conciliacao"
	"rotinas --mes $MES"
	"documentos pendentes"
	"notas entrada --lista a-manifestar --mes $MES"
	"notas entrada --lista a-classificar --mes $MES"
	"notas config"
	"extratos --ano $ANO"
	"certificado"
	"mensalidade"
	"mensalidade situacao"
	"prolabore"
	"prolabore fator-r"
)

rodar() {
	local n=$1
	shift
	# shellcheck disable=SC2086 # a divisão em palavras é intencional
	"$CTBZ" $* -o json >"$TMP/$n.out" 2>"$TMP/$n.err"
	echo $? >"$TMP/$n.rc"
}

i=0
for c in "${COMANDOS[@]}"; do
	i=$((i + 1))
	rodar "$i" "$c" &
	if ((i % PARALELO == 0)); then
		wait
	fi
done
wait

echo "### ctbz status"
cat "$TMP/status.out"
i=0
for c in "${COMANDOS[@]}"; do
	i=$((i + 1))
	echo
	echo "### ctbz $c (código $(cat "$TMP/$i.rc"))"
	cat "$TMP/$i.out"
	if [ -s "$TMP/$i.err" ]; then
		echo "--- stderr"
		cat "$TMP/$i.err"
	fi
done
