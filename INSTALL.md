# Installing Entropy Money

## Prebuilt binaries

Download the archive for your platform from the
[releases page](https://github.com/yzoz/entropy/releases), unpack it and put the
`entropy` binary somewhere on your `PATH`.

```bash
# example: Linux x86-64
tar -xzf entropy_1.0.0_linux_amd64.tar.gz
sudo install entropy /usr/local/bin/
entropy version
```

Released archives are accompanied by `checksums.txt` so you can verify integrity.

## With Go

Requires Go 1.18 or newer.

```bash
go install github.com/yzoz/entropy/cmd/entropy@v1.0.0
```

This installs the binary into `$(go env GOPATH)/bin`.

## From source

```bash
git clone https://github.com/yzoz/entropy
cd entropy
make build          # produces ./bin/entropy with version metadata
# or run without building:
make run ARGS="balance"
```

## Verifying a release build

Every release binary reports its embedded version so you can confirm what you
installed:

```bash
entropy version
entropy version --json
```
