// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/ui"
)

func offerDockerSetup(in *bufio.Reader, out io.Writer, root string) {
	stack := detectDockerStack(root)
	if stack == "" {
		return
	}
	fmt.Fprintln(out, "\n"+ui.Heading.Render("docker setup"))
	fmt.Fprintf(out, "  detected stack: %s — scaffold Dockerfile + docker-compose.yml?\n", ui.Accent.Render(stack))
	if !askYesNo(in, out, "write files?", false) {
		fmt.Fprintln(out, "  skipped")
		return
	}
	df, compose := renderDockerScaffold(stack)
	wrote := []string{}
	for _, f := range []struct {
		path string
		body string
	}{
		{filepath.Join(root, "Dockerfile"), df},
		{filepath.Join(root, "docker-compose.yml"), compose},
	} {
		if _, err := os.Stat(f.path); err == nil {
			fmt.Fprintf(out, "  %s %s already exists — skipping\n", ui.MarkWarn(), f.path)
			continue
		}
		if err := os.WriteFile(f.path, []byte(f.body), 0o644); err != nil {
			fmt.Fprintf(out, "  %s write %s: %v\n", ui.MarkFail(), f.path, err)
			continue
		}
		wrote = append(wrote, f.path)
	}
	for _, p := range wrote {
		fmt.Fprintf(out, "  %s wrote %s\n", ui.MarkSuccess(), ui.Accent.Render(p))
	}
}

func detectDockerStack(root string) string {
	probes := []struct {
		marker string
		stack  string
	}{
		{"go.mod", "go"},
		{"pyproject.toml", "python"},
		{"requirements.txt", "python"},
		{"package.json", "node"},
	}
	for _, p := range probes {
		if _, err := os.Stat(filepath.Join(root, p.marker)); err == nil {
			return p.stack
		}
	}
	return ""
}

func renderDockerScaffold(stack string) (string, string) {
	switch stack {
	case "go":
		return dockerfileGo, composeGeneric(8080)
	case "python":
		return dockerfilePython, composeGeneric(8000)
	case "node":
		return dockerfileNode, composeGeneric(3000)
	}
	return "", ""
}

const dockerfileGo = `FROM golang:1.21-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./...

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/app /app
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app"]
`

const dockerfilePython = `FROM python:3.12-slim
WORKDIR /app
COPY . .
RUN pip install --no-cache-dir -e .
EXPOSE 8000
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
`

const dockerfileNode = `FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev
COPY . .
EXPOSE 3000
CMD ["node", "server.js"]
`

func composeGeneric(port int) string {
	return fmt.Sprintf(`services:
  app:
    build: .
    ports:
      - "%d:%d"
    mem_limit: 2g
    memswap_limit: 2g
    cpus: '2.0'
    restart: unless-stopped
`, port, port)
}
