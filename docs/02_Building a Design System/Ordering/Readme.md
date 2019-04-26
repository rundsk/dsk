# Manually Ordering Aspects and Documents

Aspects and documents by default appear in the same order as they are stored on your disk. But sometimes
order matters. To manually set the order you prefix aspects or documents with an _order number_ like so: 

```
example
├── DataEntry
│   ├── 01_TextField       <-- now comes before "Button"
│   │   ├── ...
│   │   ├── 01_explain.md  <-- now comes before "api.md"
│   │   └── 02_api.md
│   ├── 02_Button
│   │   └── ...
```

Valid order number prefixes look like `01_`, `01-`, `1_` or `1-`.


