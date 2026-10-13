# hb-tags-api-register

API do MVP de pulseiras com QR code. Gera tags, resolve o próximo passo de um QR e devolve só os campos públicos de uma pulseira ativada.

A API não cadastra usuários, não ativa tags, não faz login, checkout nem notificações.

## Variáveis

| Variável | Uso |
| --- | --- |
| `DATABASE_URL` | Postgres da API, do `make migrate-*` e do `make seed` |
| `TEST_DATABASE_URL` | Postgres dos testes que gravam dados. Sem ela, esses testes são ignorados |
| `ADMIN_API_KEY` | Valor do header `x-api-key` nas rotas `/api/admin` |
| `APP_URL` | Origem da URL impressa no QR (`APP_URL/{id}`), sem barra no final. Em produção: `https://hbtags.com.br` |
| `FRONTEND_ORIGIN` | Única origem liberada no CORS das rotas públicas |
| `PORT` | Porta HTTP |

Veja `.env.example`.

## Comandos

Suba o Postgres e crie também o banco de testes, se for rodar `go test`:

```bash
docker compose up -d postgres
docker compose exec postgres psql -U hbtags -d postgres -c "CREATE DATABASE hbtags_test;"
```

Copie `.env.example` para `.env` e exporte as variáveis. Os alvos do Makefile leem `.env` quando o arquivo existe. É preciso Go 1.23+ e GNU Make.

```bash
make migrate-up
make seed
go run ./cmd/api
```

```bash
make migrate-down
go test ./...
```

A API não aplica migrations na subida. `make seed` executa `seed/seed_categories.sql` (idempotente por `ON CONFLICT`).

Rate limit: 5 requisições por segundo por IP, rajada de 20, em todas as rotas exceto `GET /healthz`. O IP é o primeiro valor de `X-Forwarded-For` ou, se o header não vier, o endereço da conexão.

## Exemplos

`ADMIN_API_KEY=change-me` e a API em `http://localhost:8080`.

### POST /api/admin/tags

`201`

```http
POST /api/admin/tags
x-api-key: change-me
Content-Type: application/json

{"quantity": 2, "batch": "lote-2026-10"}
```

```json
{"count":2,"batch":"lote-2026-10","tags":[{"id":"23456789AB","status":"GENERATED","url":"http://localhost:3000/23456789AB"}]}
```

Sem `x-api-key`: `401 {"error":"unauthorized"}`.

`quantity` fora de 1 a 1000, `batch` vazio ou maior que 50 caracteres, ou JSON inválido: `400 {"error":"invalid_body"}`.

`{}` cria 1 tag com `batch` nulo.

### GET /api/admin/tags

```http
GET /api/admin/tags?batch=lote-2026-10&limit=50&offset=0
x-api-key: change-me
```

```json
{"tags":[{"id":"23456789AB","status":"GENERATED","batch":"lote-2026-10","url":"http://localhost:3000/23456789AB"}],"limit":50,"offset":0}
```

CSV (`text/csv`), mesmas query params:

```http
GET /api/admin/tags?batch=lote-2026-10&format=csv
x-api-key: change-me
```

```csv
id,status,batch,url
23456789AB,GENERATED,lote-2026-10,http://localhost:3000/23456789AB
```

`limit` padrão 50, máximo 200. `offset` padrão 0. Sem `batch`, lista todas as tags.

### GET /api/tags/{id}/resolve

Sem autenticação. Sempre `200` e `Cache-Control: no-store`.

```http
GET /api/tags/23456789AB/resolve
```

| Status da tag | Corpo |
| --- | --- |
| `GENERATED` | `{"action":"activate"}` |
| `ACTIVATED` | `{"action":"informations"}` |
| `LOST` ou `SUSPENDED` | `{"action":"lost"}` |
| id inexistente ou malformado | `{"action":"not_found"}` |

### GET /api/public/tags/{id}

Sem autenticação. Só responde tag `ACTIVATED`. Nos outros casos: `404 {"error":"not_found"}`. `Cache-Control: no-store`.

```json
{
  "id": "23456789AC",
  "category": {"slug": "pet", "name": "Pets"},
  "profile": {"identity": {"display_name": "Rex"}}
}
```

Entram apenas campos com visibilidade efetiva `public`. A visibilidade vem de `profile_data.visibility[chave]` e, se não houver override válido, de `fields_schema`. `restricted` e `private` não saem.

### GET /api/categories

```json
{"categories":[{"id":"...","slug":"pet","name":"Pets","fields_schema":{}}]}
```

Só categorias com `active = true`.

### GET /healthz

`200 {"status":"ok"}` quando o banco responde ao ping. Caso contrário `503 {"error":"unavailable"}`.
