# Plan: AI-Sensors - Feature Breakdown

## Vision du Projet

Serveur qui raccourcit la boucle de feedback pour les agents de code en:
- Exécutant des commandes en mode watch/continu
- Capturant leur output en temps réel
- Exposant ces logs via une API simple

## Décisions Architecturales

- **Scope:** Mono-projet (un serveur = un projet)
- **Structure:** Modules Go séparés pour chaque composant logique (réutilisabilité)
- **Persistence:** Fichier JSON simple (load au démarrage, save à chaque modification)
- **Output format:** Lignes brutes (pas de metadata)

## Philosophie de Test

- **Modules découplés:** Chaque package (buffer, runner, command) doit être testable en isolation
- **Pas de mocks inutiles:** Les tests unitaires testent le module directement, sans mocker ses internals
- **Dépendances via interfaces:** Quand un module dépend d'un autre, utiliser des interfaces pour permettre l'injection
- **Tests par couche:**
  - F1-F3: Tests unitaires purs (pas de dépendances externes)
  - F4: Tests d'intégration légers (peut utiliser les vrais modules)
  - F5: Tests HTTP contre l'API (à la fin, quand tout est assemblé)

---

## Découpage en Features

### Feature 1: Ring Buffer
**But:** Stocker les N dernières lignes d'output sans exploser la mémoire

**Scope:**
- Struct `RingBuffer` avec capacité configurable (ex: 1000 lignes)
- Implémente `io.Writer` pour recevoir l'output
- Méthode `Lines()` pour récupérer le contenu actuel
- Méthode `LastN(n)` pour les N dernières lignes
- Thread-safe (mutex pour concurrent access)

**Package:** `buffer/`

**Tests:** Pur unit test - aucune dépendance externe
- Écrire des données, vérifier qu'on récupère les bonnes lignes
- Tester le comportement circulaire (overflow)
- Tester la concurrence (goroutines parallèles)

---

### Feature 2: Process Runner
**But:** Spawner un process externe et capturer son output

**Scope:**
- Struct `Process` qui encapsule un `exec.Cmd`
- Méthode `Start(output io.Writer)` qui lance le process
- Combine stdout/stderr vers le writer fourni
- Méthode `Stop()` pour arrêter proprement (SIGTERM puis SIGKILL après timeout)
- Méthode `Wait()` pour attendre la fin
- État observable: Running, Stopped, Errored

**Package:** `runner/`

**Tests:** Unit tests avec commandes simples (echo, sleep, etc.)
- Lancer un process court, vérifier l'output capturé
- Lancer un process long, le stopper, vérifier qu'il s'arrête
- Tester les différents états (Running -> Stopped, Running -> Errored)

---

### Feature 3: Command Definition + JSON Store
**But:** Définir et persister des commandes

**Scope:**
- Struct `Command` (Name, WorkDir, Cmd, Args, Env)
- `Store` interface avec impl JSON file
- Méthodes: Save, Load, Add, Remove, Get, List
- Auto-save à chaque modification

**Package:** `command/`

**Tests:** Unit tests avec fichier temp
- CRUD complet (Add, Get, Remove, List)
- Vérifier que Save/Load préserve les données
- Tester avec fichier inexistant (création automatique)

---

### Feature 4: Manager (orchestration)
**But:** Lier Store + Runner + Buffer

**Scope:**
- `Manager` struct avec:
  - Référence au command store
  - Map des processes actifs (name -> Process + Buffer)
- Méthodes:
  - `Start(name)` - charge la commande, crée buffer, lance process
  - `Stop(name)` - arrête le process
  - `Output(name)` - retourne le contenu du buffer
  - `Status(name)` - état du process

**Package:** `manager/`

**Tests:** Tests d'intégration légers
- Utilise les vrais modules (buffer, runner, command)
- Scénario complet: définir commande -> start -> lire output -> stop
- Peut utiliser un store en mémoire ou fichier temp

---

### Feature 5: API REST
**But:** Exposer tout via HTTP

**Endpoints Commandes:**
- `GET /commands` - Liste des commandes définies
- `POST /commands` - Créer une commande
- `GET /commands/{name}` - Détails d'une commande
- `DELETE /commands/{name}` - Supprimer

**Endpoints Execution:**
- `POST /commands/{name}/start` - Démarrer
- `POST /commands/{name}/stop` - Arrêter
- `GET /commands/{name}/status` - État (running/stopped/error)

**Endpoints Output:**
- `GET /commands/{name}/output` - Buffer complet
- `GET /commands/{name}/output?lines=N` - N dernières lignes

**Package:** `server/` (existant, à enrichir)

**Tests:** Tests HTTP (à la fin)
- Utiliser `httptest` pour tester les handlers
- Scénarios E2E: créer commande via API, démarrer, lire output, stopper
- Vérifier les codes de retour HTTP, les erreurs, etc.

---

## Ordre d'Implémentation

```
F1: buffer/     ─┐
                 ├──> F4: manager/ ──> F5: server/
F2: runner/     ─┤
                 │
F3: command/    ─┘
```

**F1, F2, F3** sont indépendants (peuvent être faits en parallèle ou dans n'importe quel ordre)
**F4** intègre les trois
**F5** expose le manager via HTTP

---

## Structure Finale des Packages

```
ai-sensors/
├── main.go
├── commands.json          # Persistence (créé au runtime)
├── buffer/
│   ├── ring.go
│   └── ring_test.go
├── runner/
│   ├── process.go
│   └── process_test.go
├── command/
│   ├── command.go
│   ├── store.go
│   └── store_test.go
├── manager/
│   ├── manager.go
│   └── manager_test.go
└── server/
    ├── server.go
    ├── handlers.go
    └── server_test.go
```

---

## Vérification (End-to-End)

Pour tester le système complet:

1. Démarrer le serveur: `go run main.go`
2. Créer une commande:
   ```bash
   curl -X POST localhost:3000/commands \
     -H "Content-Type: application/json" \
     -d '{"name":"watch-tests","cmd":"gotestsum","args":["--watch"]}'
   ```
3. Démarrer la commande: `curl -X POST localhost:3000/commands/watch-tests/start`
4. Lire l'output: `curl localhost:3000/commands/watch-tests/output`
5. Stopper: `curl -X POST localhost:3000/commands/watch-tests/stop`

---

## Backlog (futures features)

- **Streaming SSE:** Output en temps réel sans polling
- **Multi-projet:** Support de plusieurs projets
- **Templates:** Presets pour Go, Node, Rust, etc.
- **Filtrage:** Grep-like sur le buffer
- **MCP Integration:** Exposer comme outil MCP pour agents
