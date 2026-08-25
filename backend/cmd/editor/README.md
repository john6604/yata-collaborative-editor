# Editor CLI

Interactive terminal for editing a local CRDT document and persisting it in BoltDB.

## Run

From the repository root:

```powershell
go run ./cmd/editor
```

By default, it uses the database configured as `../../data/yata.db`, relative to the running process. You can also provide a custom path:

```powershell
go run ./cmd/editor -db ./data/editor.db
```

Or use the `YATA_DB_PATH` environment variable:

```powershell
$env:YATA_DB_PATH="./data/editor.db"
go run ./cmd/editor
```

On startup, the editor tries to restore an existing snapshot. If no valid snapshot exists, it creates a new document and immediately saves its initial identity.

## Commands

All commands are entered at the `>` prompt.

```text
print
insert <index> <char>
delete <index>
save
state
help
exit
```

## Basic Usage

Create the text `Hi`:

```text
> insert 0 H
Inserted 'H' at index 0.
> insert 1 i
Inserted 'i' at index 1.
> print
Visible:  "Hi"
Internal: START -> H -> i -> END
```

Delete the first visible character:

```text
> delete 0
Deleted the visible element at index 0.
> print
Visible:  "i"
Internal: START -> H(X) -> i -> END
```

Exit and save:

```text
> exit
Saving and closing...
```

## Editing Rules

- Indexes are zero-based.
- `insert <index> <char>` inserts before the visible character at that position.
- `insert` can append at the end by using `index = VisibleLength`.
- `delete <index>` deletes the visible character at that position.
- `<char>` must contain exactly one Unicode character, represented internally as a `rune`.
- `<char>` cannot contain spaces because the parser splits input with `strings.Fields`.
- If your terminal uses UTF-8, you can insert multibyte characters such as accented letters or emoji, as long as they are a single rune.

## Persistence

- `insert` and `delete` save the snapshot immediately after applying the change.
- `save` forces a manual snapshot save.
- `exit`, EOF, Ctrl+C, and SIGTERM perform a final save before closing BoltDB.
- If saving fails after a mutation, the change remains applied in memory, but the error is reported.

## Inspection

`print` shows the visible content and the internal linked list with tombstones.

`state` shows:

- `ClientID`
- `Clock`
- `VisibleLength`
- `InsertLog`
- `DeleteLog`
- `PendingInserts`
- `PendingDeletes`

This is useful for debugging synchronization, remote operations, and snapshot reconstruction.
