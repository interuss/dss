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
)

var (
	metaname = flag.String("metaname", "cluster-metadata", `metaname is the name 
		of the config map holding the cluster context name. It must be in the meta
		namespace.`)
	confirm = flag.Bool("confirm", true, `whether to prompt for confirmation on
		mutating commands`)
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s <command> <input.jsonnet>]\n", os.Args[0])
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
	// TODO(steeling): consider leveraging a CLI framework like cobra.
	// If these features ever make it into kubecfg, we can ditch this tool
	// completely.
	if len(os.Args) != 3 {
		usage()
	}
	ctx := context.Background()
	command, err := kubejsonnet.New(os.Args[1], os.Args[2], *metaname)
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
		log.Fatalf("error running command %s: %v", os.Args[1], err)
	}
}
