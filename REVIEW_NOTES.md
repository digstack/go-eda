## Revue - checklist actionnable

Date: 2026-01-23

Format:
- [ ] [SEVERITY] fichier:ligne - description
  - Problem: ...
  - Impact: ...
  - Fix: ...

- [ ] [HIGH] pkg/db/event_store.go:83 - Concurrence optimiste ignoree
  - Problem: expectedVersion n'est jamais utilise dans SaveEvents.
  - Impact: ecritures concurrentes non detectees, corruption du flux d'evenements.
  - Fix: verifier la version actuelle par aggregate et refuser si mismatch.

- [ ] [HIGH] pkg/db/event_store.go:88 - Event-Type manquant a la publication
  - Problem: SaveEvents publie sans header Event-Type alors que GetEvents* en depend.
  - Impact: relecture impossible (events ignores faute de type).
  - Fix: definir msg.Header.Set("Event-Type", event.Event.GetType()) avant publish.

- [ ] [HIGH] pkg/db/event_store.go:116 - Relecture JetStream incorrecte
  - Problem: SubscribeSync ne relit pas l'historique JetStream.
  - Impact: GetEventsFromVersion ne reconstruit pas l'etat complet.
  - Fix: utiliser un consumer JetStream avec DeliverAll/DeliverByStartSequence + Pull/Fetch.

- [ ] [HIGH] pkg/db/event_store.go:168 - PullSubscribe avec sujet invalide
  - Problem: PullSubscribe sur s.stream+".CONSUMER" sans consumer explicite.
  - Impact: GetAllEvents ne recupere rien ou echoue selon la config.
  - Fix: creer un consumer sur le stream, puis PullSubscribe sur le sujet du stream.

- [ ] [MEDIUM] pkg/db/event_store.go:23 - ReconnectWait en nanosecondes
  - Problem: nats.ReconnectWait(2) = 2ns, pas 2s.
  - Impact: boucle de reconnexion trop aggressive.
  - Fix: nats.ReconnectWait(2 * time.Second).

- [ ] [MEDIUM] pkg/cqrs/cqrs.go:18 - Command ID jamais renseigne
  - Problem: NewCommand ne met pas ID.
  - Impact: tracabilite/observabilite degradee.
  - Fix: generer un UUID.

- [ ] [MEDIUM] pkg/cqrs/cqrs.go:38 - Query ID jamais renseigne
  - Problem: NewQuery ne met pas ID.
  - Impact: tracabilite/observabilite degradee.
  - Fix: generer un UUID.

- [ ] [MEDIUM] pkg/cqrs/cqrs.go:90 - Buses non thread-safe
  - Problem: maps non protegees pour Register/Dispatch.
  - Impact: data races en concurrence.
  - Fix: mutex ou sync.Map.

- [ ] [LOW] pkg/module/registry.go:119 - Exposition de map interne
  - Problem: GetAllModules renvoie la map interne.
  - Impact: modifications externes difficiles a tracer.
  - Fix: retourner une copie ou un slice.

- [ ] [MEDIUM] Tests manquants
  - Problem: aucun fichier *_test.go.
  - Impact: regressions faciles sur event store/CQRS.
  - Fix: tests unitaires de base (SaveEvents/GetEvents, concurrency).
