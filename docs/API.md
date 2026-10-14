# API das pulseiras QR

Guia das rotas, do modelo e das regras desta API. Para subir o projeto, migrar e rodar os testes, use o [README](../README.md).

A API gera tags, devolve os QR Codes em JPG, diz qual é o próximo passo de um QR e, se a pulseira estiver ativada, devolve só os campos públicos do perfil. Não cadastra usuário, não ativa tag, não faz login, checkout nem notificação.

A URL impressa no QR é `APP_URL/{id}`. Em produção isso é `https://hbtags.com.br/{id}`. Essa página não é servida aqui. Quem lê o QR chama esta API a partir dela.

## Modelo

Três tabelas, definidas em `migrations/000001_init.up.sql`. O seed cria as categorias `pet`, `elderly` e `neurodivergent`.

### user

Conta de quem cadastra a pulseira. A tabela se chama `"user"` porque `user` é palavra reservada no PostgreSQL.

| Coluna | Tipo | Regra |
| --- | --- | --- |
| `id` | uuid | gerado pelo banco |
| `name` | varchar(120) | obrigatório |
| `email` | varchar(254) | obrigatório e único |
| `phone` | varchar(20) | opcional, formato E.164 |
| `consent_at` | timestamptz | consentimento, quando existir |
| `created_at` | timestamptz | padrão `now()` |

Nenhuma rota devolve dados de `"user"`.

### category

Formulário de um tipo de pulseira.

| Coluna | Tipo | Regra |
| --- | --- | --- |
| `id` | uuid | gerado pelo banco |
| `slug` | varchar(40) | único: `pet`, `elderly` ou `neurodivergent` |
| `name` | varchar(80) | obrigatório |
| `fields_schema` | jsonb | objeto ou nulo. Descreve os campos do perfil |
| `active` | boolean | padrão `true`. Só as ativas aparecem em `GET /api/categories` |

### tag

A pulseira. O `id` vai impresso no QR.

| Coluna | Tipo | Regra |
| --- | --- | --- |
| `id` | varchar(10) | chave primária. Casa com `^[2-9A-HJKMNP-Z]{10}$` |
| `status` | enum | `GENERATED`, `ACTIVATED`, `LOST` ou `SUSPENDED`. Padrão `GENERATED` |
| `batch` | varchar(50) | lote de impressão, opcional |
| `pin_hash` | varchar(255) | nulo até existir ativação. Nunca sai na API |
| `user_id` | uuid | referência a `"user"`, opcional |
| `category_id` | uuid | referência a `category`, opcional |
| `profile_data` | jsonb | objeto ou nulo. Dados de quem usa a pulseira |
| `activated_at` | timestamptz | opcional |
| `created_at` | timestamptz | padrão `now()` |

O alfabeto do `id` é `23456789ABCDEFGHJKMNPQRSTUVWXYZ` (31 símbolos). Ficam de fora `0`, `1`, `I`, `L` e `O`, para não confundir na impressão.

Status e o que o banco exige:

- `GENERATED`: `user_id`, `category_id`, `profile_data` e `activated_at` são nulos. É o estado em que a tag nasce.
- `ACTIVATED`: `user_id`, `category_id` e `profile_data` são obrigatórios.
- `LOST` e `SUSPENDED`: sem exigência extra nesta versão.

## Regras comuns

Respostas de sucesso e de erro são `application/json`, com duas exceções: `GET /api/admin/tags?format=csv` devolve `text/csv`, e `/api/admin/tags/qrs` devolve `application/zip` quando dá certo. O erro dessas rotas continua JSON, antes do arquivo começar.

Erro:

```json
{"error":"codigo_curto"}
```

Não há detalhe interno (SQL, stack, corpo do perfil).

| Código | HTTP | Quando |
| --- | --- | --- |
| `unauthorized` | 401 | `x-api-key` ausente ou diferente de `ADMIN_API_KEY` |
| `invalid_body` | 400 | JSON inválido, `quantity` fora de 1 a 1000, ou `batch` vazio ou maior que 50 caracteres |
| `invalid_query` | 400 | `limit`, `offset`, `batch` ou `format` inválidos na listagem |
| `not_found` | 404 | rota desconhecida, ou tag pública que não está `ACTIVATED` |
| `method_not_allowed` | 405 | método HTTP que a rota não aceita |
| `rate_limited` | 429 | mais requisições do que o limite do IP |
| `timeout` | 503 | a requisição passou de 10s. Em `/api/admin/tags/qrs`, o prazo é 3 minutos |
| `internal` | 500 | falha interna ou panic |
| `id_generation_failed` | 500 | não foi possível gerar ids únicos para o lote |
| `unavailable` | 503 | `GET /healthz` e o banco não responde ao ping |

Toda resposta leva `X-Request-ID`. Se o cliente enviar um id com até 128 caracteres (letras, números, `-` e `_`), ele é reutilizado. Caso contrário a API gera outro.

O limite é 5 requisições por segundo por IP, com rajada de 20. Vale para todas as rotas exceto `GET /healthz`. O IP é o primeiro valor de `X-Forwarded-For` ou, se o header não vier, o endereço da conexão. Excesso: `429 {"error":"rate_limited"}`.

CORS só em `GET /api/tags/{id}/resolve`, `GET /api/public/tags/{id}` e `GET /api/categories`. `Access-Control-Allow-Origin` é exatamente `FRONTEND_ORIGIN`. Métodos liberados: `GET` e `OPTIONS`. Headers: `Content-Type` e `X-Request-ID`. `OPTIONS` responde `204` sem corpo. As rotas de admin não têm CORS.

## Rotas

### POST /api/admin/tags

Cria tags com status `GENERATED`. Exige o header `x-api-key` igual a `ADMIN_API_KEY`. Sem chave válida: `401 {"error":"unauthorized"}`, mesmo que o método seja outro.

Corpo JSON, no máximo 4 KB:

```json
{"quantity":2,"batch":"lote-2026-10"}
```

- `quantity`: inteiro de 1 a 1000. Se o campo não vier, vale 1.
- `batch`: texto de 1 a 50 caracteres. Se não vier, a resposta traz `"batch": null`.

`{}` cria uma tag sem lote. Corpo vazio, JSON inválido, quantidade fora da faixa ou lote inválido: `400 {"error":"invalid_body"}`.

`201`:

```json
{
  "count": 2,
  "batch": "lote-2026-10",
  "tags": [
    {
      "id": "23456789AB",
      "status": "GENERATED",
      "url": "https://hbtags.com.br/23456789AB"
    }
  ]
}
```

`url` é `APP_URL` + `/` + `id`. O `id` sai do alfabeto acima, sem viés de módulo. A inserção repete até completar a quantidade, no máximo 5 vezes, ignorando colisão de chave. Se não completar: `500 {"error":"id_generation_failed"}`.

### GET /api/admin/tags

Lista tags do admin. Mesma autenticação do POST. Não devolve `profile_data`, `pin_hash` nem `user_id`.

Query:

| Parâmetro | Regra |
| --- | --- |
| `batch` | opcional. Vazio ou ausente lista todas. Acima de 50 caracteres: 400 |
| `limit` | padrão 50, máximo 200, mínimo 1 |
| `offset` | padrão 0, mínimo 0 |
| `format` | vazio ou `json` devolve JSON. `csv` devolve CSV. Outro valor: 400 |

`200` JSON:

```json
{
  "tags": [
    {
      "id": "23456789AB",
      "status": "GENERATED",
      "batch": "lote-2026-10",
      "url": "https://hbtags.com.br/23456789AB"
    }
  ],
  "limit": 50,
  "offset": 0
}
```

`format=csv` usa `text/csv; charset=utf-8`, com a mesma paginação:

```csv
id,status,batch,url
23456789AB,GENERATED,lote-2026-10,https://hbtags.com.br/23456789AB
```

Query inválida: `400 {"error":"invalid_query"}`.

### POST /api/admin/tags/qrs

Gera tags e devolve um ZIP com um JPG por id. Mesma autenticação do POST de tags. O corpo é o mesmo, com uma diferença: `quantity` ausente vale **1000**, não 1. Continua de 1 a 1000. `batch` continua opcional, de 1 a 50 caracteres. `{}` grava 1000 tags.

Sucesso é `200` com `Content-Type: application/zip` e `Content-Disposition: attachment`. O arquivo se chama `tags-qrs.zip`, ou `tags-{batch}.zip` quando o lote vem no corpo. Dentro há um `{id}.jpg` por tag. Cada imagem é o QR de `APP_URL/{id}`, 512 px, correção de erro média, JPEG qualidade 95. As linhas ficam `GENERATED`.

Erro de validação: `400 {"error":"invalid_body"}`. Falha ao completar os ids: `500 {"error":"id_generation_failed"}`. O erro sai em JSON, antes de qualquer byte do ZIP.

POST e GET desta rota podem levar até 3 minutos. As outras rotas continuam em 10 segundos.

### GET /api/admin/tags/qrs

Exporta o ZIP das tags ainda `GENERATED`, para envio ao fornecedor. `ACTIVATED`, `LOST` e `SUSPENDED` ficam de fora. Mesma autenticação e o mesmo formato de arquivo do POST.

| Parâmetro | Regra |
| --- | --- |
| `batch` | opcional. Vazio ou ausente lista todas as `GENERATED`. Acima de 50 caracteres: 400 |
| `limit` | padrão 1000, máximo 1000, mínimo 1 |
| `offset` | padrão 0, mínimo 0 |

Query inválida: `400 {"error":"invalid_query"}`. Página sem tags: `200` com um ZIP sem entradas.

### GET /api/tags/{id}/resolve

Sem autenticação. Diz só o próximo passo. Não devolve perfil, dono nem pin. Sempre `200`, inclusive quando o id não existe, e `Cache-Control: no-store`.

| Situação | Corpo |
| --- | --- |
| `GENERATED` | `{"action":"activate"}` |
| `ACTIVATED` | `{"action":"informations"}` |
| `LOST` ou `SUSPENDED` | `{"action":"lost"}` |
| id inexistente ou fora do formato | `{"action":"not_found"}` |

Id malformado e id inexistente usam o mesmo status e o mesmo corpo. O malformado nem consulta o banco.

```http
GET /api/tags/23456789AB/resolve
```

### GET /api/public/tags/{id}

Sem autenticação. Só responde tag `ACTIVATED`. Qualquer outro caso (id malformado, inexistente ou outro status) é `404 {"error":"not_found"}`, sem dizer qual dos casos foi. `Cache-Control: no-store`.

`200`:

```json
{
  "id": "23456789AC",
  "category": {"slug": "pet", "name": "Pets"},
  "profile": {
    "identity": {"display_name": "Rex"},
    "contacts": [{"name": "Ana", "phone": "+5511999999999"}]
  }
}
```

`profile` contém só campos com visibilidade efetiva `public`. Não vêm `user_id`, `pin_hash`, dados de `"user"` nem o mapa `visibility`.

### GET /api/categories

Sem autenticação. Lista categorias com `active = true`, ordenadas pelo nome.

`200`:

```json
{
  "categories": [
    {
      "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "slug": "pet",
      "name": "Pets",
      "fields_schema": {"version": 1, "fields": []}
    }
  ]
}
```

`fields_schema` é o JSON gravado na categoria. Se estiver nulo no banco, a resposta traz `null`.

### GET /healthz

Sem autenticação e sem rate limit. `200 {"status":"ok"}` se o banco responder ao ping em até 2 segundos. Caso contrário `503 {"error":"unavailable"}`.

## Visibilidade do perfil

Cada campo do formulário tem uma chave, muitas vezes com ponto: `health.blood_type`, `contacts`, `health.health_plan.card_number`.

A visibilidade efetiva é:

1. `profile_data.visibility[chave]`, se o valor for exatamente `public`, `restricted` ou `private`.
2. Senão, `visibility` do mesmo campo em `fields_schema.fields`.
3. Qualquer outro texto de override é ignorado e vale o schema.

Só entra na resposta o campo cuja visibilidade efetiva é `public` e que tem valor presente. `null` e campo ausente ficam de fora. `false` e `0` entram. `restricted` e `private` nunca saem. Chave que não está no schema também não sai, mesmo que exista no perfil.

A resposta reconstrói o objeto. `health.blood_type` vira `profile.health.blood_type`. Um irmão privado, como `health.health_plan.card_number`, não vem junto. `contacts`, quando público, é copiado inteiro (a lista de contatos é um único campo).

O mapa `visibility` em si não aparece em `profile`. Schema ou perfil ilegível produz `"profile": {}` e a rota continua `200` se a tag estiver `ACTIVATED`.

Exemplo. Schema com `identity.display_name` público, `extra.microchip` restrito e `extra.diagnosis` privado. Perfil:

```json
{
  "identity": {"display_name": "Rex"},
  "extra": {"microchip": "999", "diagnosis": "secreto"},
  "visibility": {"extra.diagnosis": "public"}
}
```

Resposta pública:

```json
{
  "identity": {"display_name": "Rex"},
  "extra": {"diagnosis": "secreto"}
}
```

`microchip` fica de fora. `diagnosis` sai porque o override é `public`.

Quem tem o id da pulseira vê esses campos públicos. Não há login nesse passo.

## Ambiente

A API não sobe se faltar uma destas variáveis, ou se o valor for inválido.

| Variável | Uso |
| --- | --- |
| `DATABASE_URL` | Postgres da API, das migrations e do seed |
| `ADMIN_API_KEY` | valor esperado no header `x-api-key` |
| `APP_URL` | origem da URL do QR (`APP_URL/{id}`), absoluta, `http` ou `https`, sem barra no final |
| `FRONTEND_ORIGIN` | única origem do CORS, sem path |
| `PORT` | porta HTTP, de 1 a 65535 |

`TEST_DATABASE_URL` não é necessária para a API. Os testes que gravam no banco a leem e são ignorados se ela não existir.

Ordem no banco: `make migrate-up`, depois `make seed`. A API não migra sozinha. Os comandos estão no [README](../README.md).
