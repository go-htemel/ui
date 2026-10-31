# 0012 — Variadic children win over trailing options

**Status:** Accepted (known wart)

## Context
Go allows one variadic parameter. Components whose natural last argument is `children...`
cannot also take `opts ...Option`.

## Decision
Children stay variadic (it is what makes pages read well). Where an option is needed
afterwards, a sibling constructor carries it: `BreadcrumbsLabel`, `StepsLabel`,
`NavbarSignal`, `SettingsLayoutLabel`, `DrawerSized`, `PaginationSignal`. Components with
few children take `opts ...Option` and put children in `WithContent`.

## Consequences
- Slight API sprawl; the naming rule (`XxxLabel`, `XxxSignal`, `XxxSized`) keeps it
  predictable. Revisit if Go ever grows optional parameters.
