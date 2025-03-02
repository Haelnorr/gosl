package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	args := setupFlags()
	ctx := context.Background()
	if err := run(ctx, os.Stdout, args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
