#!/usr/bin/env bash
# Coleta, só com leituras e em paralelo, tudo o que o checkup da empresa precisa.
#
# Uso: coletar.sh [AAAA-MM]     (padrão: o mês atual)
#
# Imprime uma seção por comando: o JSON da saída, o stderr (totais e avisos) e o código de
# saída. Termina com 3 se não houver sessão (nenhuma salva ou expirada: faça o login antes),
# com 1 se não der para conferir a sessão por outro motivo (rede, erro do servidor) e com 0
# nos demais casos, mesmo que algum comando falhe: a falha aparece na seção dele.
# CTBZ_BIN escolhe o binário (padrão: ctbz).

set -u

CTBZ=${CTBZ_BIN:-ctbz}
MES=${1:-$(date +%Y-%m)}
ANO=${MES%%-*}
# Mês anterior: as notas de entrada dele vencem no dia 05 deste mês.
M=$((10#${MES#*-} - 1))
if ((M == 0)); then
	MES_ANTERIOR=$(printf '%04d-12' $((10#$ANO - 1)))
else
	MES_ANTERIOR=$(printf '%s-%02d' "$ANO" "$M")
fi
# Sessão conferida antes: os processos em paralelo não refazem o login ao mesmo tempo.
PARALELO=4

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

if ! "$CTBZ" status -o json >"$TMP/status.out" 2>"$TMP/status.err"; then
	cat "$TMP/status.out" "$TMP/status.err"
	if grep -q "nenhuma sessão salva" "$TMP/status.err"; then
		echo "Sem sessão: faça o login (ctbz login) e rode de novo." >&2
		exit 3
	fi
	echo "Não foi possível conferir a sessão (rede ou servidor); tente de novo antes de refazer o login." >&2
	exit 1
fi
if grep -q '"situacao": *"expirada"' "$TMP/status.out"; then
	cat "$TMP/status.out" "$TMP/status.err"
	echo "Sessão expirada: faça o login (ctbz login) e rode de novo." >&2
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
	"notas entrada --lista a-manifestar --mes $MES_ANTERIOR"
	"notas entrada --lista a-classificar --mes $MES_ANTERIOR"
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
