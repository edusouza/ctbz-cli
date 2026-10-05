# ctbz extratos

Lista a situação dos extratos bancários por mês e conta

Lista, para cada conta bancária e mês, a situação do extrato na Contabilizei (ex.: ABERTO
enquanto o mês não foi fechado) e a integração com o banco. Os extratos são a base da
conciliação: meses sem extrato aparecem nas rotinas como "Importar extrato bancário".

--ano filtra pelo ano; a API devolve todos os meses disponíveis.

## Uso

```
ctbz extratos [flags]
```

## Exemplos

```sh
  ctbz extratos
  ctbz extratos --ano 2026 -o csv
```

## Subcomandos

- [`ctbz extratos classificar`](ctbz_extratos_classificar.md): Troca a classificação de um lançamento do extrato
- [`ctbz extratos contas`](ctbz_extratos_contas.md): Lista as classificações aceitas nos lançamentos do extrato de uma competência
- [`ctbz extratos desfazer-desmembramento`](ctbz_extratos_desfazer-desmembramento.md): Volta um lançamento desmembrado do extrato ao original
- [`ctbz extratos desmembrar`](ctbz_extratos_desmembrar.md): Divide um lançamento do extrato em partes com classificações diferentes
- [`ctbz extratos excluir`](ctbz_extratos_excluir.md): Exclui o extrato importado de uma conta num mês
- [`ctbz extratos importar`](ctbz_extratos_importar.md): Importa o extrato bancário (OFX ou PDF) de uma conta num mês
- [`ctbz extratos lancamentos`](ctbz_extratos_lancamentos.md): Lista os lançamentos do extrato de uma conta bancária num mês

## Flags

```
      --ano int   só os meses deste ano
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
