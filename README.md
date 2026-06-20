# Go Scaling

Go Scaling is a small, module-by-module learning repo for Go.

Each module is a root-level directory with a tiny runnable program, lesson text, exercises, and Dockerfiles. The lessons build from first principles: source code becomes a compiled binary, and the binary becomes an operating-system process.

## Shared Runtime Image

Build the shared runtime image once from the repo root:

```bash
docker build --target runtime-base -t go-scaling:runtime .
```

Module Dockerfiles use this image after compiling their own Go binary.

## Start With Module 001

After building the shared runtime image, enter the first module:

```bash
cd module-001-hello-go
```

Then follow that module's README to build and run its Docker image.
