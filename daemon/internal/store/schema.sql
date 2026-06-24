-- Aqno local database schema (SQLite via libSQL-compatible modernc.org/sqlite).
-- Every statement is idempotent (IF NOT EXISTS) so it doubles as the migration.
-- See docs/context.md §10 for the source of truth.

-- ===== Configuration and persona =====
CREATE TABLE IF NOT EXISTS config (
  chave         TEXT PRIMARY KEY,
  valor         TEXT NOT NULL,
  atualizado_em TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS persona (
  id          INTEGER PRIMARY KEY,
  nome        TEXT NOT NULL,
  tipo_avatar TEXT NOT NULL DEFAULT 'orbe',   -- orbe | animal | personagem | imagem
  avatar_ref  TEXT,
  cor_aura    TEXT NOT NULL DEFAULT '#8B5CF6',
  voz         TEXT,
  tom         TEXT NOT NULL DEFAULT 'amigavel',
  wake_word   TEXT NOT NULL DEFAULT 'aqno',
  usuario     TEXT,                            -- name of the human owner
  criado_em   TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ===== Contexts (companies + personal) =====
CREATE TABLE IF NOT EXISTS contextos (
  id        INTEGER PRIMARY KEY,
  nome      TEXT NOT NULL UNIQUE,
  cor       TEXT NOT NULL,                     -- palette key: violet|teal|amber|rose|blue
  ai_mode   TEXT NOT NULL DEFAULT 'cloud',     -- cloud | local_only
  arquivado INTEGER NOT NULL DEFAULT 0,
  criado_em TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ===== Calendar =====
CREATE TABLE IF NOT EXISTS eventos (
  id           INTEGER PRIMARY KEY,
  contexto_id  INTEGER REFERENCES contextos(id) ON DELETE SET NULL,
  titulo       TEXT NOT NULL,
  tipo         TEXT NOT NULL DEFAULT 'reuniao', -- reuniao | bloco_foco | tarefa | pessoal
  inicio       TEXT NOT NULL,                   -- 'HH:MM' (recurring) or ISO (single)
  fim          TEXT,
  rrule        TEXT,                            -- iCalendar RRULE; NULL = single event
  data_unica   TEXT,                            -- 'YYYY-MM-DD' for non-recurring
  lembrete_min INTEGER,
  ativo        INTEGER NOT NULL DEFAULT 1,
  origem_voz   TEXT,
  criado_em    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS excecoes (
  id          INTEGER PRIMARY KEY,
  evento_id   INTEGER NOT NULL REFERENCES eventos(id) ON DELETE CASCADE,
  data        TEXT NOT NULL,                    -- affected occurrence (YYYY-MM-DD)
  tipo        TEXT NOT NULL DEFAULT 'cancelado',-- cancelado | remarcado
  novo_inicio TEXT,
  novo_fim    TEXT
);

-- ===== Tasks =====
CREATE TABLE IF NOT EXISTS tarefas (
  id           INTEGER PRIMARY KEY,
  contexto_id  INTEGER REFERENCES contextos(id) ON DELETE SET NULL,
  titulo       TEXT NOT NULL,
  status       TEXT NOT NULL DEFAULT 'aberta',  -- aberta | em_andamento | concluida
  prioridade   INTEGER NOT NULL DEFAULT 0,
  prazo        TEXT,
  origem_voz   TEXT,
  criado_em    TEXT NOT NULL DEFAULT (datetime('now')),
  concluido_em TEXT
);

-- ===== Voice interactions (history / audit) =====
CREATE TABLE IF NOT EXISTS interacoes (
  id          INTEGER PRIMARY KEY,
  contexto_id INTEGER REFERENCES contextos(id) ON DELETE SET NULL,
  transcricao TEXT NOT NULL,
  intencao    TEXT,
  resultado   TEXT,
  criado_em   TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ===== Chat =====
CREATE TABLE IF NOT EXISTS conversas (
  id        INTEGER PRIMARY KEY,
  titulo    TEXT,
  criado_em TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS mensagens (
  id          INTEGER PRIMARY KEY,
  conversa_id INTEGER NOT NULL REFERENCES conversas(id) ON DELETE CASCADE,
  papel       TEXT NOT NULL,                    -- user | assistant | tool
  conteudo    TEXT NOT NULL,
  ref_kind    TEXT,                             -- memory | action (optional UI hint)
  ref_label   TEXT,
  ref_tone    TEXT,
  criado_em   TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ===== Knowledge graph =====
CREATE TABLE IF NOT EXISTS entidades (
  id          INTEGER PRIMARY KEY,
  tipo        TEXT NOT NULL,                    -- empresa | projeto | tarefa | evento | pessoa | decisao | nota | context
  rotulo      TEXT NOT NULL,
  contexto_id INTEGER REFERENCES contextos(id) ON DELETE SET NULL,
  meta        TEXT,                             -- free JSON
  embedding   TEXT,                             -- JSON float array (in lieu of sqlite-vec)
  criado_em   TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS relacoes (
  id         INTEGER PRIMARY KEY,
  origem_id  INTEGER NOT NULL REFERENCES entidades(id) ON DELETE CASCADE,
  destino_id INTEGER NOT NULL REFERENCES entidades(id) ON DELETE CASCADE,
  tipo       TEXT NOT NULL,                     -- pertence_a | depende_de | mencionou | etc.
  peso       REAL NOT NULL DEFAULT 1.0
);

-- ===== Servers / VPS (credentials live in the Keychain, only a ref here) =====
CREATE TABLE IF NOT EXISTS servidores (
  id           INTEGER PRIMARY KEY,
  nome         TEXT NOT NULL,
  host         TEXT NOT NULL,
  porta        INTEGER NOT NULL DEFAULT 22,
  usuario      TEXT NOT NULL,
  auth_tipo    TEXT NOT NULL DEFAULT 'senha',   -- senha | chave
  keychain_ref TEXT NOT NULL,
  criado_em    TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ===== Audit log (VPS actions + cloud-boundary crossings) =====
CREATE TABLE IF NOT EXISTS auditoria (
  id        INTEGER PRIMARY KEY,
  categoria TEXT NOT NULL,                      -- vps | llm | config
  acao      TEXT NOT NULL,
  detalhe   TEXT,
  criado_em TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ===== Full-text search (FTS5; created at runtime when available) =====

-- ===== Indexes =====
CREATE INDEX IF NOT EXISTS idx_eventos_contexto   ON eventos(contexto_id);
CREATE INDEX IF NOT EXISTS idx_eventos_ativo       ON eventos(ativo);
CREATE INDEX IF NOT EXISTS idx_tarefas_status      ON tarefas(status);
CREATE INDEX IF NOT EXISTS idx_mensagens_conversa  ON mensagens(conversa_id);
CREATE INDEX IF NOT EXISTS idx_relacoes_origem     ON relacoes(origem_id);
CREATE INDEX IF NOT EXISTS idx_relacoes_destino    ON relacoes(destino_id);
CREATE INDEX IF NOT EXISTS idx_excecoes_evento     ON excecoes(evento_id);
