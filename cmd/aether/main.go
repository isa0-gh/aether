package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/isa0-gh/aether/internal/builder"
)

func main() {
	content  := flag.String("content",   "content",     "Directory containing markdown posts")
	templates := flag.String("templates", "templates",   "Directory containing HTML templates")
	static   := flag.String("static",    "static",      "Directory containing static assets (css, js)")
	output   := flag.String("output",    "public",      "Output directory for the generated site")
	cfg      := flag.String("config",    "config.json", "Path to config.json")
	flag.Parse()

	opts := builder.Options{
		ContentDir:  *content,
		TemplateDir: *templates,
		StaticDir:   *static,
		OutputDir:   *output,
		ConfigFile:  *cfg,
	}

	if err := builder.Build(opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Printf("Site built → %s\n", *output)
}
