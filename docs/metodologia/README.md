# Metodologia

Como a investigação foi feita, para repetir quando algo mudar no site.

## Ferramentas

`curl` (com `--compressed`: o servidor sempre responde com gzip), `grep -o` e Python para
decodificar. Sem navegador e sem proxy de interceptação: tudo vem do HTML e do JavaScript
servidos.

## 1. Mapear o login

```sh
UA='Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0 Safari/537.36'
J=cookies.txt

curl -sS --compressed -c $J -b $J -A "$UA" https://app.contabilizei.com.br/login -o login.html
TOKEN=$(grep -o 'name="token" value="[^"]*"' login.html | sed 's/.*value="//;s/"//')

curl -sS --compressed -c $J -b $J -A "$UA" -X POST https://app.contabilizei.com.br/login \
  -H 'Origin: https://app.contabilizei.com.br' -H 'Referer: https://app.contabilizei.com.br/login' \
  --data-urlencode "user=$CTBZ_USER" --data-urlencode "password=$CTBZ_PASSWORD" \
  --data-urlencode "token=$TOKEN" -o otp.html -D -
```

Removendo as imagens inline, a lógica da página fica legível:

```sh
sed -E 's/data:[^"]{200,}/DATA/g' otp.html | sed -n '/<\/style>/,$p'
```

Foi daí que saíram as funções `send()`/`verify()` e o significado dos status 206/200/404.

## 2. OTP e empresa

```sh
curl -sS --compressed -c $J -b $J -A "$UA" -X POST https://app.contabilizei.com.br/login \
  -H 'Content-Type: text/plain;charset=UTF-8' --data-raw "v=2&otp=123456" -o sel.html -w '%{http_code}\n'

curl -sS --compressed -c $J -b $J -A "$UA" -X POST https://app.contabilizei.com.br/selecionarempresa \
  --data-urlencode "cnpj=22222222000122" --data-urlencode "token=$TOKEN" -o final.html -D -
```

Decodificando o `localStorage` gravado pelo `final.html`:

```python
import re, base64, json, urllib.parse
html = open("final.html").read()
for k, v in re.findall(r'localStorage.setItem\("([^"]*)","([^"]*)"\)', html):
    print(k, json.loads(urllib.parse.unquote(base64.b64decode(v).decode())).keys())
```

## 3. Achar as APIs

1. Baixar `/painel-de-controle/` com os cookies e listar os `<script src>`/`<link href>`.
2. No `app.<hash>.js`, procurar a configuração (`VUE_APP_[A-Z_]+:"…"`) e as instâncias
   (`axios.create({baseURL:…})`), e depois o cabeçalho do módulo (`t.d(a,"c",…)`) para
   saber qual letra exportada corresponde a qual base.
3. Extrair as chamadas: `grep -oE '\w+\["[a-f]"\]\.(get|post|put|delete|patch)\("[^"]+"'`.
4. Baixar os ≈165 chunks **com cookie** (sem ele, voltam 106 bytes de redirecionamento)
   e repetir.
5. Para os parâmetros de um endpoint, ler o código ao redor:

   ```sh
   grep -ohE '.{0,150}relatorios-ms/gerar-balancete/.{0,200}' js/*.js
   # → r["c"].get("relatorios-ms/gerar-balancete/".concat(ano,"/").concat(mes))
   ```

Além de `X["c"].get("…")`, o script reconhece `.concat(x, "/sufixo")`, caminhos guardados em
variáveis (`var e="…"; X["c"].post(e)`), `Object(X["c"])({method, url})` e o vue-resource do
front de notas de entrada (`this.$http.post("/api/emissor/…")`). Os testes ficam em
`scripts/test_extrair_endpoints.py` (`python3 -m unittest discover -s scripts`).

Os passos 1 a 4 estão automatizados em
[`scripts/extrair-endpoints.py`](https://github.com/edusouza/ctbz-cli/blob/main/scripts/extrair-endpoints.py):

```sh
ctbz login && scripts/extrair-endpoints.py > docs/api/catalogo.md
```

## 4. Validar

Com `ctbz api <caminho>`, chamando só `GET`s de leitura e registrando o status e as
chaves de primeiro nível da resposta (sem guardar valores pessoais). Resultado em
[endpoints-verificados.md](../api/endpoints-verificados.md).

Para isolar o papel de cada cookie, a mesma chamada foi repetida com cada um sozinho
(ver [sessão](../autenticacao/04-sessao-e-cookies.md#cookies)).

## Armadilhas encontradas

- **Dois OTPs quase juntos.** Uma tentativa da CLI falhou depois de enviar as
  credenciais (o servidor já tinha mandado o e-mail); a nova tentativa gerou um
  segundo e-mail. O código do primeiro foi recusado com 404. Só o mais recente vale.
- **`/dev/null` parece terminal.** Checar `os.ModeCharDevice` no stdin trata
  `/dev/null` como terminal, e a CLI ficava esperando um código que nunca viria.
  A correção foi usar `term.IsTerminal` e, se o prompt receber EOF, salvar como pendente.
- **404 com HTML** não quer dizer rota inexistente: geralmente faltam segmentos de
  caminho (`caixa/listpaginada/` precisa de `{ano}/{mes}/{porPagina}/{pagina}`).
- **Mesmas rotas, bases diferentes:** chamar uma rota da plataforma em `/api/legado/`
  dá 404.
- **Chunks sem cookie** voltam vazios (redirecionamento), o que esconde a maioria dos endpoints.

## Cuidados

- Não chamar `POST`/`PUT`/`PATCH`/`DELETE` em investigação: são ações reais na
  contabilidade da empresa (emitir nota, confirmar pagamento, aceitar termos).
- Evitar logins em sequência: cada um dispara um e-mail, há limite de sessões
  simultâneas e bloqueio temporário por tentativas.
- Não versionar respostas reais: contêm CPF, CNPJ, endereço e dados fiscais. Esta
  documentação registra só a estrutura.
