package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/interuss/dss/pkg/tools/kubejsonnet"
	cliflag "k8s.io/component-base/cli/flag"
)

var (
	confirm = flag.Bool("confirm", true, `whether to prompt for confirmation on
		mutating commands`)
)

func usage() {
	fmt.Fprint(os.Stderr, "usage: kubejsonnet <command> <input.jsonnet|input.json>]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func maybeConfirm(cmd *kubejsonnet.Command) error {
	if !cmd.IsMutatingCommand() {
		return nil
	}
	// Print the diff even if we're not prompting for confirm.
	diff, err := cmd.Diff()
	if err != nil {
		return fmt.Errorf("error computing diff: %v", err)
	}
	fmt.Println(diff)
	if *confirm {
		return waitForConfirm()
	}
	return nil
}

func waitForConfirm() error {
	fmt.Printf("\nProceed? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("could not get user response: %v", err)
	}
	response = strings.ToLower(strings.TrimSpace(response))
	if len(response) == 0 || !(response == "y" || response == "yes") {
		return errors.New("cancelled by user, aborting")
	}
	return nil
}

func main() {
	flag.Parse()
	// Setup kubectl flags
	cliflag.InitFlags()

	// TODO(steeling): consider leveraging a CLI framework like cobra.
	if len(os.Args) < 3 {
		usage()
	}
	// Remove kubejsonnet and input file.
	input, args := os.Args[len(os.Args)-1], os.Args[1:len(os.Args)-1]

	ctx := context.Background()
	command, err := kubejsonnet.New(input, args)
	defer command.Cleanup()

	if err != nil {
		log.Fatalf("error creating kubejsonnet command: %v", err)
	}

	if err := maybeConfirm(command); err != nil {
		log.Fatal(err)
	}

	out, err := command.Run(ctx)

	fmt.Println(out)
	if err != nil {
		log.Fatalf("error running command %v: %v", os.Args, err)
	}
}
