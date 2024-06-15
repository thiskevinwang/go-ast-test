Build the Go program as a wasm file

```
GOARCH=wasm GOOS=js go build -o html/main.wasm main.go
```

Copy Go's wasm_exec.js file to the html directory.
This file is necessary for calling `const go = new Go();` in JS.

```
cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./html
```

> [!NOTE]
>
> `<script src="wasm_exec.js"></script>` Needs to be specified early in the HTML file.

Local testing...

```
npx serve ./html
```