# Editor CLI

Terminal interactiva para editar un documento CRDT local y persistirlo en BoltDB.

## Ejecutar

Desde la raiz del repositorio:

```powershell
go run ./cmd/editor
```

Por defecto usa la base configurada en `../../data/yata.db`, relativa al proceso. Tambien puedes indicar otra ruta:

```powershell
go run ./cmd/editor -db ./data/editor.db
```

O usar la variable de entorno `YATA_DB_PATH`:

```powershell
$env:YATA_DB_PATH="./data/editor.db"
go run ./cmd/editor
```

Al iniciar, el editor intenta restaurar el snapshot existente. Si no hay uno valido, crea un documento nuevo y guarda su identidad inicial.

## Comandos

Todos los comandos se escriben en el prompt `>`.

```text
print
insert <index> <char>
delete <index>
save
state
help
exit
```

## Uso Basico

Crear el texto `Hi`:

```text
> insert 0 H
Insertado 'H' en el indice 0.
> insert 1 i
Insertado 'i' en el indice 1.
> print
Visible:  "Hi"
Interno:  START -> H -> i -> END
```

Eliminar el primer caracter visible:

```text
> delete 0
Eliminado el elemento visible del indice 0.
> print
Visible:  "i"
Interno:  START -> H(X) -> i -> END
```

Salir guardando:

```text
> exit
Guardando y cerrando...
```

## Reglas de Edicion

- Los indices son base cero.
- `insert <index> <char>` inserta antes del caracter visible en esa posicion.
- `insert` permite insertar al final usando `index = VisibleLength`.
- `delete <index>` elimina el caracter visible en esa posicion.
- `<char>` debe contener exactamente un caracter Unicode, representado internamente como un `rune`.
- `<char>` no puede contener espacios porque el parser separa la linea con `strings.Fields`.
- Si tu terminal esta en UTF-8, puedes insertar caracteres multibyte como letras con tilde o emoji, siempre que sean un solo rune.

## Persistencia

- `insert` y `delete` guardan el snapshot inmediatamente despues de aplicar el cambio.
- `save` fuerza un guardado manual.
- `exit`, EOF, Ctrl+C y SIGTERM hacen un guardado final antes de cerrar BoltDB.
- Si un guardado falla despues de una mutacion, el cambio queda aplicado en memoria, pero se reporta el error.

## Inspeccion

`print` muestra el contenido visible y la lista interna con tombstones.

`state` muestra:

- `ClientID`
- `Clock`
- `VisibleLength`
- `InsertLog`
- `DeleteLog`
- `PendingInserts`
- `PendingDeletes`

Esto es util para depurar sincronizacion, operaciones remotas y reconstruccion de snapshots.
