# Skyharbor Content

Skyharbor Content is the complete learning project for Datapack. It models a small multiplayer game's combat content: effects, items, playable characters, and level-growth balancing data.

The project is intentionally small enough to understand in one sitting, but each table demonstrates a different part of the editor. Open `skyharbor-content.dtproj` to begin.

## Learning Path

1. Open the `common` schema file and inspect `#GameID`, `#Rarity`, `#EffectKind`, `#NonEmptyString`, `#NonNegative`, and `#Probability`.
2. Open `content.effect` and edit a name, enum, number, nullable description, and array of tags.
3. Open `content.item`. Use the foreign-key suggestions for `effectID`, then open the cell context menu to add a comment. The `effectIDs` array shows that an array is one cell, not several spreadsheet columns.
4. Edit the `position` structure as one structured value. Its fields are defined by the table's local `_local.#Position` schema.
5. Inspect `sellPrice` in `content.item`. It is a formula preview, not a value written into `rows`; changing `price` changes the preview after compilation.
6. Open `content.character` and follow the `startingItem` foreign key to an item. Its `attack` column is calculated from the same row.
7. Open `balance.growth` and switch to manual compilation. Change `baseValue`, observe the pending compile state, then compile explicitly. The formulas reference other rows using the complete logical table name.
8. Select a cell and use the formula bar, paste a rectangular range, and use Undo/Redo to compare the source change with the displayed result.
9. Use the Schema editor to create or rename a definition. The tree shows the difference between a public Schema file and a table's local Schema definitions.
10. Configure project export selection and the table export columns, then export CSV. The exported files are in `exports/`; arrays and structures use JSON encoding.
11. In the `content.item` column menu, set a display name for `price`. The header becomes `Buy Price (price)`; formulas, source fields, and export configuration continue to use `price`.

Display names are UI metadata and never change a field's stable identity. Without a display name, the header shows only the field name. Export settings show both names, but persist the field name.

The application-level actions use the same project as a safe playground: use File > New project and File > Save project as to practice project lifecycle, use the project-tree context menus to create a table or Schema file, and use the table/Schema context menus to rename, move, or delete objects. These actions are deliberately not encoded in the source project because they change project structure; the functional and Playwright test suites exercise them on temporary copies.

## Datapack Concepts

Datapack is not a spreadsheet with a hidden formula layer:

- A table's `#Row` is the contract for every data row. A cell display value is derived from its CUE expression and can differ from the source text.
- A formula lives in `_tool.formulas` and is evaluated for preview and export. Its result is not copied back into `rows`.
- A foreign key validates that an ID exists in another table and provides navigation. It does not let a formula read arbitrary fields from that table.
- Public Schemas are shared rules. Local Schemas belong to one table and cannot be used by another table.
- Arrays and structures remain single typed cells. Their editor is structured, and CSV encoding is configured explicitly.
- A nullable field is not the same as an empty string. Clearing a cell uses the type's safe default behavior.
- Export always performs a full project compile first, even when project export selection excludes a table's output file.

## Experiments

The initial project compiles cleanly. Make a copy before trying destructive experiments:

- Change an `effectID` to an unknown value and use the diagnostic to locate it.
- Try deleting `burn` from the effect table while items still reference it. The editor protects referenced rows.
- Rename `mend` and inspect the foreign-key updates in `content.item`.
- Set `chance` above `1` to see a constraint diagnostic, then undo the edit.
- Delete a local Schema definition while it is used by a column to see reference protection.

The complete baseline coverage is split between this project and the test suite. The project covers source-backed data behavior: public and local Schemas, type and constraint rules, defaults and nullable values, scalar and array foreign keys, arrays, structures, formulas, comments, diagnostics, navigation, CSV configuration, compilation, and history. The temporary-project workflows cover project creation, Save As, table and Schema lifecycle, column lifecycle, row lifecycle, paste, settings, language/theme, compile mode, export selection, and undo/redo.

## Source Layout

```text
skyharbor-content.dtproj     project entry point
cue.mod/module.cue           CUE module identity
schema/common.cue            public shared types and constraints
tables/content/effect/effect.cue       enum, nullable text, arrays, and constraints
tables/content/item/item.cue           scalar/array foreign keys and a structure
tables/content/character/character.cue same-row formula and item foreign key
tables/balance/growth/growth.cue       cross-row formulas and manual compile practice
```

The generated `exports/` directory is excluded from source scanning and is not part of the maintained input data.
