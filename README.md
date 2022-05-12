GOALL
----

Simple program to compile your golang code to all known targets.

# Install

```
go install github.com/inkeliz/goall
```

# Usage

It's very simple to use, just use `goall -name {{pattern}} {{go args}}`:

```
goall -name "yourproject_{{OS}}_{{ARCH}}" build .
```

It will compile everything in to the current folder. You can pass any flag to golang compiler:

```
goall -name "yourproject_{{OS}}_{{ARCH}}" build -o "myfolder" .
```

# Flags

- `-name`: specify the name pattern  (`{{OS}}` and `{{ARCH}})` is replaced by each OS and ARCH).
- `-ignore-web`: removes compilation for web (js)
- `-ignore-mobile`: removes compilation for mobile devices (ios/android)